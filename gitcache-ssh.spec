# FIXME: Follow https://fedoraproject.org/wiki/PackagingDrafts/Go

Summary: Simple SSH based Git cache
Name: gitcache-ssh
Version: %{version}
Release: %{release}%{?dist}
License: BSD
Prefix: /usr
Prefix: /usr/bin
Prefix: /usr/share/man
Source0: gitcache-ssh-%{version}.tar.gz
Requires: parallel
BuildRequires: go-md2man
#BuildRequires: compiler(go-compiler)
#BuildRequires: golang(github.com/sirupsen/logrus)
#BuildRequires: golang(github.com/sirupsen/logrus/hooks/syslog)
#BuildRequires: golang(github.com/mattn/go-shellwords)
#BuildRequires: golang(github.com/neechbear/gogiturl)
#BuildRequires: golang(github.com/tcnksm/go-gitconfig)
URL: https://github.com/nokia/gitcache-ssh
Packager: Nicola Worthington <nicolaw@tfb.net>

%description
gitcache-ssh is a simple SSH based Git cache.

This tool was conceived and written to dramatically reduce the bandwidth
consumed by developers and CI/CD build and test systems that constantly
communicate with geographically remote Git repositories.

It is designed to be (for the most part, after initial setup), transparrent to
the user or automation systems that are using it.

It provides an asymetric (pull/read-only) cache function to shim between the Git
client and the originating remote upstream Git repository.

%clean
rm -rf --one-file-system "%{buildroot}"

%prep
%setup -q

%install
make install DESTDIR="%{buildroot}" prefix=/usr

%files
/usr/bin/gitcache-ssh
/etc/cron.d/gitcache-refresh.cron
/usr/share/man/man1/gitcache-ssh.1.gz
%docdir /usr/share/doc/gitcache-ssh
/usr/share/doc/gitcache-ssh
%dir %attr(2770, root, gitcache-ssh) /var/cache/git

%changelog
