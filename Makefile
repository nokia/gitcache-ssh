BINARY := gitcache-ssh
VET_REPORT := vet.report
TEST_REPORT := tests.xml
GOARCH ?= amd64

include version.mk
BUILD := $(shell git rev-parse HEAD)
LDFLAGS ?= -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"

builddir := build
prefix := /usr/local
bindir := $(prefix)/bin

.PHONY: clean install dist rpm deb

.DEFAULT_GOAL: all
all: clean $(BINARY)

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
	$(RM) $(BINARY) $(BINARY)-*-$(GOARCH)* $(BINARY)*.rpm $(BINARY)*.deb $(TEST_REPORT) $(VET_REPORT)
	$(RM) -r $(builddir)
	go clean

deb: PKGTYPE=deb
deb: dist

rpm: PKGTYPE=rpm
rpm: dist

dist: $(BINARY)
	install -m 0755 -d $(builddir)$(bindir)
	install -m 0755 $(BINARY) $(builddir)$(bindir)
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
