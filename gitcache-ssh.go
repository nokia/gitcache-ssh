// Git wrapper for use on a local gitlab-cache machine.
//
// TODO: explain what this means such that the purpose of the program is clear
//       to a first-time reader.
//
// This wrapper is intended to be called directly by SSH via the ForceCommand
// sshd_config directive that overloads whatever command was requested to be
// executed by the remote client.
//
// Author: Nicola Worthington <nicola.worthington@nokia.com>, <nicolaw@tfb.net>
// Caveat: This is (was) the first Go code that I've ever written. While I hope
//         that obvious obscenities will have been picked up in code review,
//         please don't view me too harshly if I committed any sins.
//
// Test on the command line:
//       SSH_ORIGINAL_COMMAND="git-upload-pack git@github.com:nokia/gitcache-ssh.git" ./gitcache-ssh
//
// TODO: Accept optional fallback default remote repo from environment variable.
//
// TODO: Fix logging so that STDERR and Syslog logging can have independent log
//       levels. Ideally, we want to do the following:
//     - Always log everything (up to DEBUG loglevel) to syslog, no matter what.
//     - Change the loglevel filter for messages going to the console (Stderr),
//       independently of the syslog loglevel. By default console messages
//       should have a loglevel of INFO.
//     - When the console (Stderr) isn't an IsTerminal, then a different
//       loglevel and different (simpler) textFormatter should be used for the
//       console logs. This should be independent of the syslog loglevel and
//       textFormatter, which should remain unchanged.
//     - The above will allow all message output to be send through a single
//       log interface, allowing constant detail in syslog, but differing output
//       to the console depending on runtime context.

package main

import (
	"bufio"
	"log/syslog"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"golang.org/x/crypto/ssh/terminal"

	log "github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	shellwords "github.com/mattn/go-shellwords"
	giturl "github.com/neechbear/gogiturl"
	gitconfig "github.com/tcnksm/go-gitconfig"
)

const (
	logIdentity   = "gitcache-ssh"   // Syslog identity to use in log messages
	envPath       = "/usr/bin:/bin"  // Safe $PATH environment variable contents
	cacheRootDir  = "/var/cache/git" // Local cached git repositories location
	gitBinaryPath = "/usr/bin"       // Used to qualify gitAction path w/Exec()
	gitBinary     = "/usr/bin/git"   // Full path to git command binary

	// TODO: It would be nice to allow this default to be set by an environment
	//       variable like GIT_CACHESSH_DEFAULT_REMOTE maybe.
	defaultRemoteHostname = "github.com"

	// Documentation URLs for this program, for use in fatal log messages.
	thisSourceURL = "https://github.com/nokia/gitcache-ssh"
)

// Variables to identify the build.
var (
	Version string
	Build   string
)

var (
	trustedGitCommands = []string{
		"git-upload-pack",    // git clone
		"git-upload-archive", // git archive
	}
	trustedEnvironment = []string{
		"SSH_ORIGINAL_COMMAND",       // Set by sshd when ForceCommand is set
		"GIT_CACHESSH_[a-zA-Z0-9_]+", // Toggle custom cache functionality
		"GIT_TRACE",                  // Increate git command debug log output
		"GIT_CONFIG",                 // Location of ~/.gitconfig /etc/gitconfig
		"USER",                       // Current username
		"HOME",                       // Home directory of current user
	}
	shellUnsafeRegexp = []string{
		"[^@a-z0-9 /:.,_'\"-]+", // Any chars not matching safe alpha-num etc
		".*[$;`!|].*",           // Explicitly shell sigils/flow control chars
	}
)

// Matches anchored regular expressions against any slice indexes.
func matchInSlice(a string, list []string) bool {
	for _, b := range list {
		match, err := regexp.MatchString("^"+b+"$", a)
		if err != nil {
			log.Fatalf("MatchString error str \"%s\"; %#v",
				b, err)
		} else if match {
			return true
		}
	}
	return false
}

// Unset all environment variables except those passed to us.
func cleanEnvironment(trustedEnvironment []string) {
	log.Debugf("Original environment: %#v", os.Environ())
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if !matchInSlice(pair[0], trustedEnvironment) {
			os.Unsetenv(pair[0])
		} else {
			log.Debugf("Saving trustedEnvironment %s=%s", pair[0], pair[1])
		}
	}
	log.Debugf("Sanitized environment: %#v", os.Environ())
}

// Determine if the cache should be synchronised or not.
func shouldCacheSync() (bool, error) {
	sync := false
	if os.Getenv("GIT_CACHESSH_SYNC") != "" {
		sync = isTrue(os.Getenv("GIT_CACHESSH_SYNC"))
	} else {
		str, err := gitconfig.Global("cache.ssh.sync")
		if err != nil {
			return sync, err
		}
		sync = isTrue(str)
	}
	return sync, nil
}

// Returns true if the given string contains unsafe shell characters that match
// any of the shellUnsafeRegexp list regexes.
func matchUnsafeCharacters(s string) bool {
	for i, r := range shellUnsafeRegexp {
		r = "^" + r + "$"
		match, err := regexp.MatchString(r, s)
		if err != nil {
			log.Fatalf("MatchString error regexp[%d], str \"%s\"; %#v",
				i, s, err)
		} else if match {
			log.Printf("String \"%s\" matched shellUnsafeRegexp %s", s, r)
			return true
		}
	}
	return false
}

// Perform a git remote update, or git clone mirror to sync our locally cached
// copy of the repository.
func syncRepository(cacheRepoPath string, remoteRepo string) error {
	cmdArgs := []string{"-C", cacheRepoPath, "remote", "update", "--prune"}
	if !pathExists(cacheRepoPath) {
		cmdArgs = []string{"clone", "--mirror", remoteRepo, cacheRepoPath}
	}

	log.WithFields(log.Fields{
		"os.Environ": os.Environ(),
	}).Debugf("Executing %s \"%s\"", gitBinary, strings.Join(cmdArgs, "\" \""))

	cmd := exec.Command(gitBinary, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Debugf("exec.StdoutPipe failure syncing cache repository with "+"remote; %#v", err)
		return err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		// No point in continuing if we can't start the command.
		log.Debugf("exec.Start failure syncing cache repository with "+
			"remote; %#v", err)
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	for scanner.Scan() {
		line := scanner.Text()
		log.Info(line)
	}
	if err := scanner.Err(); err != nil {
		// Warn and try to soldier on if we couldn't read the command's output.
		log.Errorf("bufio.Scan failure syncing cache repository with "+
			"remote while reading cmd.StdoutPipe; %#v", err)
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0.
			// This works on both Unix and Windows. Although package syscall is
			// generally platform dependent, WaitStatus is defined for both Unix
			// and Windows and in both cases has an ExitStatus() method with
			// the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				log.Debugf("exit status: %d", status.ExitStatus())
			}
		} else {
			log.Debugf("exec.Wait failure syncing cache repository with "+
				"remote; %#v", err)
		}
		return err
	}

	return nil
}

// Parse the first argument in SSH_ORIGINAL_COMMAND (the second index)
// containing the remote Git repository, returning the fully qualified version
// (use the default remote if it did not contain a remote host), and the local
// path where we intend to store the cached copy of this repository.
func parseSSHOrigCmdRepo(repo string) (string, string) {
	exp := regexp.MustCompile(`(?i)^/?((?:[a-z]+://|[\[\]a-z0-9_\.-@]+:).+)`)
	match := exp.FindStringSubmatch(repo)

	var remoteRepo string
	if len(match) >= 2 {
		// Regular URL or Git@/scp style remoteRepo.
		remoteRepo = match[1]
	} else {
		// Unqualified path only repo requires qualification with defaultRemote.
		remoteRepo = "git@" + defaultRemoteHostname + ":" + repo
	}

	// Parse out the remoteRepo as a regular URL object so we can extract things
	// like the hostname, username (if set), port, path etc.
	u, err := giturl.Parse(remoteRepo)
	if err != nil {
		log.WithFields(log.Fields{
			"err":        err,
			"repo":       repo,
			"remoteRepo": remoteRepo,
		}).Fatalf("Failed to giturl.Parse('%s'); %s", remoteRepo, err.Error())
	} else if u.Host == "" {
		log.WithFields(log.Fields{
			"err":        err,
			"repo":       repo,
			"remoteRepo": remoteRepo,
		}).Fatalf("Failed to giturl.Parse('%s'); unable to extract hostname",
			remoteRepo)
	}

	// Build a string representing where we will store a local cached version of
	// this repository.
	remoteHostDir, path := u.Host, u.Path
	if u.User != nil && u.User.Username() != "" {
		remoteHostDir = u.User.Username() + "@" + remoteHostDir
	}
	if u.Host == defaultRemoteHostname &&
		!strings.HasSuffix(strings.TrimRight(path, "/"), ".git") {
		// Horrific little hack for our GitLab server, where we appear to have
		// people pulling from repo URLs both with and without the .git suffix.
		// We try and normalise in this situation and force the cacheRepoPath
		// to always include a .git suffix, so we don't end up with two versions
		// of every repo.
		path = strings.TrimRight(path, "/") + ".git"
	}
	cacheRepoPath := filepath.Join(cacheRootDir, remoteHostDir, path)

	log.WithFields(log.Fields{
		"remoteRepo":    remoteRepo,
		"cacheRepoPath": cacheRepoPath,
	}).Debugf("Parsed repository %s okay.", repo)

	return remoteRepo, cacheRepoPath
}

// Tests is a filesystem path exists or not.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// Tests the given string for "truthiness".
// https://en.wikipedia.org/wiki/Truthiness
func isTrue(s string) bool {
	b, err := strconv.ParseBool(s)
	if err == nil && b {
		return true
	}
	i, err := strconv.ParseInt(s, 10, 8)
	if err == nil && i > 0 {
		return true
	}
	return false
}

func init() {
	// FIXME: Optimise logging, see TODO comments in header.
	logOutput := os.Stderr
	logLevel := log.InfoLevel

	// GIT_TRACE  and GIT_CACHESSH_TRACE envinronment variables can be set to
	// enable more verbose debug logging.
	for _, v := range []string{"GIT_TRACE", "GIT_CACHE_SSH_TRACE"} {
		if isTrue(os.Getenv(v)) {
			logLevel = log.DebugLevel
		}
	}

	// Add a syslog hook to log to syslog as well.
	hook, err := lSyslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
	if err != nil {
		log.Error("Unable to connect to local syslog daemon")
	} else {
		log.AddHook(hook)
	}

	// Simplify logging format if we're not running on an interactive terminal.
  if terminal.IsTerminal(int(logOutput.Fd())) {
		log.SetFormatter(&log.TextFormatter{
			DisableColors:    true,
			DisableTimestamp: true,
		})
	}

	log.SetOutput(logOutput)
	log.SetLevel(logLevel)
}

func main() {
	log.WithFields(log.Fields{
		"version": Version,
		"build":   Build,
	}).Debugf("Starting %s", logIdentity)

	// Sanitize the environment as we're probably running SUID as root.
	cleanEnvironment(trustedEnvironment)
	os.Setenv("PATH", envPath)

	// Usage contextual fatal logger (help point the confused user in the
	// right direction).
	logUsage := log.WithFields(log.Fields{
		"source":  thisSourceURL,
		"program": logIdentity,
		"version": Version,
		"build":   Build,
	})

	// We rely on SSH_ORIGINAL_COMMAND being set for us to function.
	sshOrigCmd, err := shellwords.Parse(os.Getenv("SSH_ORIGINAL_COMMAND"))
	if err != nil {
		logUsage.Fatalf("Unable to parse SSH_ORIGINAL_COMMAND (\"%s\"); %v",
			os.Getenv("SSH_ORIGINAL_COMMAND"), err)
	}

	// We expect (1) the Git command, (2) the Git command arguments (repo) to be
	// parsed from SSH_ORIGINAL_COMMAND.
	if len(sshOrigCmd) != 2 {
		logUsage.Fatalf("Too few arguments found in SSH_ORIGINAL_COMMAND; "+
			"found %d when 2 were expected!", len(sshOrigCmd))
	}

	// Check for potentially unsafe nastiness in the strings we're going to use.
	for _, s := range sshOrigCmd {
		if matchUnsafeCharacters(s) ||
			strings.Contains(s, "../") || strings.Contains(s, "..\\") {
			log.Fatalf("You were trying to do something naughty; aborting!")
		}
	}

	// Parse out the local cache repo path and the remote upstream repo from
	// the parsed slice version of SSH_ORIGINAL_COMMAND envinronment variable.
	gitAction := sshOrigCmd[0]
	if !matchInSlice(gitAction, trustedGitCommands) {
		logUsage.Fatalf("Unrecognized or unauthorized command \"%s\".",
			gitAction)
	}
	remoteRepo, cacheRepoPath := parseSSHOrigCmdRepo(sshOrigCmd[1])

	// Determine if we should sync before changing effective user and
	// environment, so that we can read the user's ~/.gitconfig if necessary.
	shouldSyncCache, _ := shouldCacheSync()

	// If we are running SUID root then do the right thing and make both our
	// effective and reported UIDs match, and update HOME and USER environment
	// variables.
	euid := os.Geteuid()
	if euid != os.Getuid() {
		user, err := user.LookupId(strconv.Itoa(euid))
		// user.LookupId appears to only just being implemented in golang on
		// Linux as we speak (March 2017). URGH!
		if err != nil {
			log.Debug(err)
			if euid == 0 {
				os.Setenv("HOME", "/root")
				os.Setenv("USER", "root")
			}
		} else {
			os.Setenv("HOME", user.HomeDir)
			os.Setenv("USER", user.Username)
		}
		syscall.Setreuid(os.Geteuid(), os.Geteuid())
	}

	// Sync a copy of the remote repository to our local cache if necessary.
	log.Debugf("shouldCacheSync=%v", shouldSyncCache)
	if shouldSyncCache || !pathExists(cacheRepoPath) {
		log.Infof("Updating %s from %s", cacheRepoPath, remoteRepo)
		err := syncRepository(cacheRepoPath, remoteRepo)
		if err == nil {
			log.Infof("Finished update of %s from %s",
				cacheRepoPath, remoteRepo)
		} else {
			log.Errorf("Failed update of %s from %s; %s",
				cacheRepoPath, remoteRepo, err.Error())
		}
	}

	gitActionArgv := []string{gitAction, cacheRepoPath}
	// Fully qualify the gitAction binary command with a path if necessary.
	if gitAction[:1] != "/" || !pathExists(gitAction) {
		gitAction = filepath.Join(gitBinaryPath, gitAction)
	}

	// Perform a passthrough to the original SSH command by exec()'ing our
	// modified and sanitized version, pointing to our own local cache repo.
	os.Unsetenv("SSH_ORIGINAL_COMMAND")
	log.Debugf("Environment: %#v", os.Environ())
	log.Debugf("Executing passthrough of %s \"%s\"",
		gitAction, strings.Join(gitActionArgv, "\" \""))
	err = syscall.Exec(gitAction, gitActionArgv, os.Environ())

	// We should never get this far because we just replaced ourself with an
	// Exec() syscall. Something has gone HORRIBLY wrong!!!
	if err != nil {
		log.Panicf("syscall.Exec passthrough to %s failed; %#v", gitAction, err)
	}
	log.Panic("This can never happen!")
}
