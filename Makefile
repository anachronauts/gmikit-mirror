.PHONY: all check install clean FORCE

PREFIX := /usr/local
SYSCONFDIR := $(PREFIX)/etc
DATADIR := $(PREFIX)/share
GMIKITCONFDIR := $(SYSCONFDIR)/gmikit
GMIKITDATADIR := $(DATADIR)/gmikit
BINDIR := $(PREFIX)/bin
SBINDIR := $(PREFIX)/sbin
BINPREFIX := gmikit-
INSTALLFLAGS :=
EXTRAGOLDFLAGS :=
STAGE := stage
PKGDIR := out

all: convert gateway get

clean:
	-rm convert gateway get

check:
	go test

FORCE:

convert: FORCE
	go build -ldflags="$(EXTRAGOLDFLAGS)" -o $@ anachronauts.club/repos/gmikit/cmd/convert

gateway: FORCE
	go build -ldflags="$(EXTRAGOLDFLAGS) -X main.confDir=$(GMIKITCONFDIR) -X main.dataDir=$(GMIKITDATADIR)" -o $@ anachronauts.club/repos/gmikit/cmd/gateway

get: FORCE
	go build -ldflags="$(EXTRAGOLDFLAGS)" -o $@ anachronauts.club/repos/gmikit/cmd/get

install: all
	install $(INSTALLFLAGS) -d $(BINDIR) $(SBINDIR) $(GMIKITCONFDIR) $(GMIKITDATADIR)/templates
	install $(INSTALLFLAGS) -m 755 convert $(BINDIR)/$(BINPREFIX)convert
	install $(INSTALLFLAGS) -m 755 gateway $(SBINDIR)/$(BINPREFIX)gateway
	install $(INSTALLFLAGS) -m 755 get $(BINDIR)/$(BINPREFIX)get
	install $(INSTALLFLAGS) -m 644 example/gateway.conf $(GMIKITCONFDIR)/gateway.conf.sample
	install $(INSTALLFLAGS) -m 644 example/templates/1x.html $(GMIKITDATADIR)/templates
	install $(INSTALLFLAGS) -m 644 example/templates/2x.html $(GMIKITDATADIR)/templates
	install $(INSTALLFLAGS) -m 644 example/templates/3x.html $(GMIKITDATADIR)/templates
	install $(INSTALLFLAGS) -m 644 example/templates/4x.html $(GMIKITDATADIR)/templates
	install $(INSTALLFLAGS) -m 644 example/templates/5x.html $(GMIKITDATADIR)/templates
	install $(INSTALLFLAGS) -m 644 example/templates/error.html $(GMIKITDATADIR)/templates
	install $(INSTALLFLAGS) -m 644 example/templates/response.html $(GMIKITDATADIR)/templates
	install $(INSTALLFLAGS) -m 644 example/templates/style.css $(GMIKITDATADIR)/templates
	install $(INSTALLFLAGS) -m 644 example/templates/unknown.html $(GMIKITDATADIR)/templates

package:
	-rm install.log
	$(MAKE) PREFIX=$(STAGE)$(PREFIX) INSTALLFLAGS=-Minstall.log install
	cat MANIFEST > $(STAGE)/+MANIFEST
	echo prefix $(PREFIX) >> $(STAGE)/+MANIFEST
	awk '/type=file/ { print substr($$1, index($$1, "$(PREFIX)")) }' install.log > $(STAGE)/plist
	mkdir -p $(PKGDIR)
	pkg create -o "$(PKGDIR)" -r "$(STAGE)" -M "$(STAGE)/+MANIFEST" -p "$(STAGE)/plist"
	pkg repo "$(PKGDIR)"

