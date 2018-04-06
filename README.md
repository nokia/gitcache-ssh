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

See [gitcache-ssh.md](gitcache-ssh.md) or `gitcache-ssh(1)` man page for more
comprehensive documentation.

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

## Design Overview

To be completed...

       _    /||
      ( }   \||D
    | /\__,=[_]           Git client rewrites pull
    |_\_  |---|  +----->  URL as specified in the
    |  |/ |   |           url.insteadOf configuration.
    |  /_ |   |
                                  +
    git clone ...                 |
                                  |                        +--------------+
                                  v                        |              |
                                                           |  +--------+  |
    +---[Git cache host]-------------------------+         |  | Repo A |  |
    |                                            |         |  +--------+  |
    |                                            |         |              |
    |                                            |         |  +--------+  |
    |                                            |         |  | Repo B |  |
    |                                            |         |  +--------+  |
    |                                            |  <---+  |              |
    |                                            |         |  +--------+  |
    |                                            |         |  | Repo C |  |
    |                                            |         |  +--------+  |
    |                                            |         |              |
    +--------------------------------------------+         +--------------+

## To-do

See [gitcache-ssh.md](gitcache-ssh.md) or `gitcache-ssh(1)` man page.

## License

BSD 3-Clause "New" or "Revised" License.

A permissive license similar to the BSD 2-Clause License, but with a 3rd clause
that prohibits others from using the name of the project or its contributors to
promote derived products without written consent.

See [LICENSE](LICENSE) file for full details.

## Author

Nicola Worthington <nicola.worthington@nokia.com>, <nicolaw@tfb.net>.

