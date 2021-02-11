.PHONY: all install clean FORCE

PREFIX := /usr/local
SYSCONFDIR := $(PREFIX)/etc
GMIKITCONFDIR := $(SYSCONFDIR)/gmikit
BINDIR := $(PREFIX)/bin
SBINDIR := $(PREFIX)/sbin
BINPREFIX := gmikit-
GOFLAGS :=

all: convert gateway get

convert: FORCE
	go build $(GOFLAGS) -o $@ anachronauts.club/repos/gmikit/cmd/convert

gateway: FORCE
	go generate anachronauts.club/repos/gmikit/cmd/gateway/templates
	go build $(GOFLAGS) -ldflags "-X main.confDir=$(GMIKITCONFDIR)" -o $@ anachronauts.club/repos/gmikit/cmd/gateway

get: FORCE
	go build $(GOFLAGS) -o $@ anachronauts.club/repos/gmikit/cmd/get

install: all
	install -d $(BINDIR) $(SBINDIR) $(GMIKITCONFDIR)
	install -m 755 convert $(BINDIR)/$(BINPREFIX)convert
	install -m 755 gateway $(SBINDIR)/$(BINPREFIX)gateway
	install -m 755 get $(BINDIR)/$(BINPREFIX)get
	install -m 644 example/gateway.conf $(GMIKITCONFDIR)

clean:
	-rm convert gateway get