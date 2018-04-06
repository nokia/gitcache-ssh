[![Build Status](https://travis-ci.org/nokia/gitcache-ssh.svg?branch=master)](https://travis-ci.org/nokia/gitcache-ssh)
[![Maintainability](https://api.codeclimate.com/v1/badges/ad5d773a12b517ed5735/maintainability)](https://codeclimate.com/github/nokia/gitcache-ssh/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/ad5d773a12b517ed5735/test_coverage)](https://codeclimate.com/github/nokia/gitcache-ssh/test_coverage)

# gitcache-ssh

`gitcache-ssh` is a simple SSH based Git cache.

This tool was conceived and written to dramatically reduce the bandwidth
consumed by developers and CI/CD build and test systems that constantly
communicate with geographically remote Git repositories.

It is designed to be (for the most part, after initial setup), transparrent to
the user or automation systems that are using it.

It provides an asymetric (pull/read-only) cache function to shim between the Git
client and the originating remote upstream Git repository.

## Building

```
$ make
$ make deb
$ make rpm
```

### Prerequisites

* A valid working Golang build environment. Don't forget to define your
  `$GOROOT`, `$GOPATH` and `$GOBIN` environment variables.
* The FPM packaging tool.
* An OpenSSH server.

## Installation

Create a suitable `/var/cache/git/` directory to store the local cached copies
of remote Git repositories.

```
# mkdir -p /var/cache/git/
```

Add a suitable `Match Group` or `Match User` block to your OpenSSH
`/etc/ssh/sshd_config` configuration file to force the `gitcache-ssh` wrapper
command to be executed upon login instead of the user's shell.

For example, to enable the `gitcache-ssh` command for all users who are a member
of the `developers` group:

```
# cat << 'EOF' >> /etc/ssh/sshd_config
AcceptEnv LANG LC_* GIT_CACHESSH_*
Match Group developers
  AllowTCPForwarding no
  X11Forwarding no
  ForceCommand /usr/local/bin/gitcache-ssh
EOF
# systemctl restart ssh
```

## Client Configuration

Git clients need to be configured to pull from the cache for specific
repositories or hosts. This is done using the `insteadOf` and `pushInsteadOf`
configuration directives.

The following example will enable the cache for SSH-based access to all
GitHub.com repositories:

```
$ git config --global --add url."ssh://gitcache.host/git@github.com:".insteadOf "git@github.com:"
$ git config --global --add url."ssh://gitcache.host/git@github.com:".insteadOf "ssh://git@github.com/"
$ git config --global --add url."git@github.com:".pushInsteadOf "git@github.com:"
$ git config --global --add url."ssh://git@github.com/".pushInsteadOf "ssh://git@github.com/"
```

Enabling for HTTP and HTTPS-based access can be achieved by simply modifying the
URLs that are being re-written:

```
$ git config --global --add url."ssh://gitcache.host/git@github.com:".insteadOf "https://github.com/"
$ git config --global --add url."ssh://gitcache.host/git@github.com:".insteadOf "http://github.com/"
$ git config --global --add url."https://github.com/".pushInsteadOf "https://github.com/"
$ git config --global --add url."http://github.com/".pushInsteadOf "http://github.com/"
```

The `git config -l | grep insteadof` command will display your current URL
re-writing configuration.

### Force Cache Refresh from Upstream Origin

The `gitcache-ssh` wrapper can be forced to refresh the local cache by setting
the environment variable `GIT_CACHESSH_SYNC=true`. The OpenSSH client will not
forward this environment variable to the remote SSH Git cache host by default,
so a small change is necessary either to the local client's `~/.ssh/ssh_config`
file, or to the SSH command line that Git will invoke.

Ad-hoc command line based solution:

```
$ GIT_CACHESSH_SYNC=true GIT_SSH_COMMAND="ssh -o SendEnv=GIT_CACHESSH_*" git pull
```

Permanent configuration based solution:

```
$ cat >> ~/.ssh/config <<EOF
Host gitcache.host
  SendEnv GIT_CACHESSH_*
EOF
$ GIT_CACHESSH_SYNC=true git pull
```

## To-do

* Replace `fpm` packaging with native RPM spec file and Debian package
  configuration directory.
* Complete documentation.
* Tidy up logging.
* Address `TODO` and `FIXME` mentioned in the source.
* Vendor dependencies for easier reproducable builds.
* Add unit tests.
* Add man page.
* Move as many constants as possible out of the Go binary, into a configuration
  file, or to be read from environment variables.
* Add support to pass-through to the users' default shell if it is detected
  that the SSH client is not a `git` command.
* Add installation helper script to configure OpenSSH for you.
* Setup CI pipeline in Travis CI.
* Setup code analysis with CodeClimate.

## License

BSD 3-Clause "New" or "Revised" License.

A permissive license similar to the BSD 2-Clause License, but with a 3rd clause
that prohibits others from using the name of the project or its contributors to
promote derived products without written consent.

See [LICENSE](LICENSE) file for full details.

## Author

Nicola Worthington <nicola.worthington@nokia.com>, <nicolaw@tfb.net>.

