BINARY := gitcache-ssh
VET_REPORT := vet.report
TEST_REPORT := tests.xml
GOARCH ?= amd64

include version.mk
NAME := gitcache-ssh
DEBVER := $(shell grep -Eom1 '^[a-z0-9\+\.\-]+ \(.+\) ' debian/changelog | cut -d'(' -f2 |  cut -d')' -f1)
LDFLAGS ?= -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"
BUILD := $(shell git rev-parse HEAD)

name := gitcache-ssh

builddir := $(name)-$(VERSION)
disttar := $(name)-$(VERSION).tar.gz
distrpm := $(name)-$(VERSION)-$(REVISION).noarch.rpm
distdebtar := $(name)_$(VERSION).orig.tar.gz
distdeb := $(name)_$(VERSION)_all.deb

prefix := /usr
bindir := $(prefix)/bin
sharedir = $(prefix)/share
docsdir = $(sharedir)/doc/$(name)
mandir = $(sharedir)/man
man1dir = $(mandir)/man1
confdir := /etc
crondir := $(confdir)/cron.d

DISTFILES := $(BINARY) $(name).1 gitcache-refresh.cron LICENSE

.PHONY: clean install dist distclean veryclean rpm deb

.DEFAULT_GOAL: all
all: clean $(BINARY)

$(name).1: $(name).md
	go-md2man -in $< -out $@

$(BINARY):
	go get ./...
	go build $(LDFLAGS) -o $(@) .

linux:
	GOOS=linux GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BINARY)-linux-$(GOARCH) .

darwin:
	GOOS=darwin GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BINARY)-darwin-$(GOARCH) .

windows:
	GOOS=windows GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BINARY)-windows-$(GOARCH).exe .

test:
	test "$(VERSION)-$(REVISION)" = "$(DEBVER)" # debian/changelog matches version
	if ! command -v go2xunit 2>/dev/null; then \
		go get github.com/tebeka/go2xunit; \
		go install github.com/tebeka/go2xunit; \
	fi
	godep go test -v ./... 2>&1 | go2xunit -output $(TEST_REPORT)

vet:
	godep go vet ./... > $(VET_REPORT) 2>&1

fmt:
	go fmt $$(go list ./... | grep -v /vendor/)

install:
	go install $(LDFLAGS) ./...

clean:
	$(RM) $(BINARY) $(BINARY)-*-$(GOARCH)* $(TEST_REPORT) $(VET_REPORT)
	$(RM) -r $(builddir)
	go clean

distclean:
	$(RM) $(distrpm) $(distdeb) $(disttar) $(distdebtar)

veryclean: clean distclean

deb: PKGTYPE=deb
deb: dist

rpm: PKGTYPE=rpm
rpm: dist

$(builddir): $(DISTFILES)
	mkdir $@
	cp -r $^ $@/

$(distdebtar): $(disttar)
	ln -f $< $@

$(disttar): $(builddir)
	tar -zchf $@ $(builddir)

dist: $(DISTFILES)
	install -m 0755 -d $(builddir)$(bindir)
	install -m 0755 $(BINARY) $(builddir)$(bindir)
	install -m 0644 gitcache-refresh.cron $(builddir)$(crondir)
	install -m 0644 $(name).1 $(builddir)$(man1dir)
	install -m 0644 LICENSE $(builddir)$(docsdir)
	test -s *.$(PKGTYPE) || fpm -s dir -t $(PKGTYPE) -C $(builddir) \
		--name $(BINARY) \
		--version $(VERSION) \
		--iteration $(REVISION) \
		--maintainer "Nicola Worthington <nicolaw@tfb.net>" \
		--vendor Nokia \
		--url https://github.com/nokia/gitcache-ssh \
		--category Development \
		--description "Caching Git Wrapper" \
		--deb-no-default-config-files \
		.
