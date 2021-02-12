.PHONY: all install clean FORCE

PREFIX := /usr/local
SYSCONFDIR := $(PREFIX)/etc
GMIKITCONFDIR := $(SYSCONFDIR)/gmikit
BINDIR := $(PREFIX)/bin
SBINDIR := $(PREFIX)/sbin
BINPREFIX := gmikit-
EXTRAGOLDFLAGS :=

all: convert gateway get

FORCE:

convert: FORCE
	go build -ldflags="$(EXTRAGOLDFLAGS)" -o $@ anachronauts.club/repos/gmikit/cmd/convert

gateway: FORCE
	go generate anachronauts.club/repos/gmikit/cmd/gateway/templates
	go build -ldflags="$(EXTRAGOLDFLAGS) -X main.confDir=$(GMIKITCONFDIR)" -o $@ anachronauts.club/repos/gmikit/cmd/gateway

get: FORCE
	go build -ldflags="$(EXTRAGOLDFLAGS)" -o $@ anachronauts.club/repos/gmikit/cmd/get

install: all
	install -d $(BINDIR) $(SBINDIR) $(GMIKITCONFDIR)
	install -m 755 convert $(BINDIR)/$(BINPREFIX)convert
	install -m 755 gateway $(SBINDIR)/$(BINPREFIX)gateway
	install -m 755 get $(BINDIR)/$(BINPREFIX)get
	install -m 644 example/gateway.conf $(GMIKITCONFDIR)/gateway.conf.sample

clean:
	-rm convert gateway get
