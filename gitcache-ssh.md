# gitcache-ssh

## NAME

gitcache-ssh - Simple SSH based Git cache.

## SYNOPSIS

`gitcache-ssh` <`git-upload-pack`|`git-upload-archive`> ...

## DESCRIPTION

gitcache-ssh is a simple SSH based Git cache.

This tool was conceived and written to dramatically reduce the bandwidth
consumed by developers and CI/CD build and test systems that constantly
communicate with geographically remote Git repositories.

It is designed to be (for the most part, after initial setup), transparrent to
the user or automation systems that are using it.

It provides an asymetric (pull/read-only) cache function to shim between the Git
client and the originating remote upstream Git repository.

## INSTALLATION

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

    # cat << 'EOF' >> /etc/ssh/sshd_config
    AcceptEnv LANG LC_* GIT_CACHESSH_*
    Match Group developers
      AllowTCPForwarding no
      X11Forwarding no
      ForceCommand /usr/bin/gitcache-ssh
    EOF
    # systemctl restart ssh

## CLIENT CONFIGURATION

Git clients need to be configured to pull from the cache for specific
repositories or hosts. This is done using the `insteadOf` and `pushInsteadOf`
configuration directives.

The following example will enable the cache for SSH-based access to all
GitHub.com repositories:

    $ git config --global --add url."ssh://gitcache.host/git@github.com:".insteadOf "git@github.com:"
    $ git config --global --add url."ssh://gitcache.host/git@github.com:".insteadOf "ssh://git@github.com/"
    $ git config --global --add url."git@github.com:".pushInsteadOf "git@github.com:"
    $ git config --global --add url."ssh://git@github.com/".pushInsteadOf "ssh://git@github.com/"

Enabling for HTTP and HTTPS-based access can be achieved by simply modifying the
URLs that are being re-written:

    $ git config --global --add url."ssh://gitcache.host/git@github.com:".insteadOf "https://github.com/"
    $ git config --global --add url."ssh://gitcache.host/git@github.com:".insteadOf "http://github.com/"
    $ git config --global --add url."https://github.com/".pushInsteadOf "https://github.com/"
    $ git config --global --add url."http://github.com/".pushInsteadOf "http://github.com/"

The `git config -l | grep insteadof` command will display your current URL
re-writing configuration.

### FORCED CACHE REFRESH

The `gitcache-ssh` wrapper can be forced to refresh the local cache by setting
the environment variable `GIT_CACHESSH_SYNC=true`. The OpenSSH client will not
forward this environment variable to the remote SSH Git cache host by default,
so a small change is necessary either to the local client's `~/.ssh/ssh_config`
file, or to the SSH command line that Git will invoke.

Ad-hoc command line based solution:

    $ GIT_CACHESSH_SYNC=true GIT_SSH_COMMAND="ssh -o SendEnv=GIT_CACHESSH_*" git pull

Permanent configuration based solution:

    $ cat >> ~/.ssh/config <<EOF
    Host gitcache.host
      SendEnv GIT_CACHESSH_*
    EOF
    $ GIT_CACHESSH_SYNC=true git pull

## SECURITY CONSIDERATIONS

Great care should be taken when using `gitcache-ssh` in any environment that
contains private or senstivie projects, as the tool effectively operates as a
man-in-the-middle.

By default, Git repositories that have been cached by one user are automatically
world-readable to all other users of the same Git cache host.

This mode of operation could be modified through judicious application of
more restrictive user and group permissions under the `/var/cache/git/`
directory. However, doing so would partially (but not entirely) defeat the
purpose of using the cache to begin with.

## ENVIRONMENT

Setting `GIT_CACHESSH_SYNC` to boolean true will force the local cache to be
refreshed.

## FILES

`/usr/bin/gitcache-ssh`, `/var/cache/git/`, `/etc/cron.d/gitcache-refresh.cron`,
`/usr/share/doc/gitcache-ssh/`

## SEE ALSO

`git config --global http.proxy`, `git clone --reference`

## TODO

* Replace `fpm` packaging with native RPM spec file and Debian package
  configuration directory.
* Complete documentation.
* Tidy up logging.
* Address `TODO` and `FIXME` mentioned in the source.
* Vendor dependencies for easier reproducable builds.
* Add unit tests.
* Complete this man page.
* Move as many constants as possible out of the Go binary, into a configuration
  file, or to be read from environment variables.
* Add support to pass-through to the users' default shell if it is detected that
  the SSH client is not a `git` command.
* Add installation helper script to configure OpenSSH for you.
* Setup CI pipeline in Travis CI.
* Setup code analysis with CodeClimate.

## AUTHOR

Nicola Worthington <nicolaw@tfb.net>.

## COPYRIGHT

BSD 3-Clause License

Copyright (c) 2018, Nokia
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

Redistributions of source code must retain the above copyright notice, this
list of conditions and the following disclaimer.

Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

Neither the name of the copyright holder nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
