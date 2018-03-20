# gitcache-ssh

`gitcache-ssh` is a simple Git cache

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
Match Group developers
  AllowTCPForwarding no
  X11Forwarding no
  ForceCommand /usr/local/bin/gitcache-ssh
EOF
# systemctl restart ssh
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

