.PHONY: all check install clean FORCE

PREFIX := /usr/local
SYSCONFDIR := $(PREFIX)/etc
GMIKITCONFDIR := $(SYSCONFDIR)/gmikit
BINDIR := $(PREFIX)/bin
SBINDIR := $(PREFIX)/sbin
BINPREFIX := gmikit-
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
	go generate anachronauts.club/repos/gmikit/cmd/gateway/templates
	go build -ldflags="$(EXTRAGOLDFLAGS) -X main.confDir=$(GMIKITCONFDIR)" -o $@ anachronauts.club/repos/gmikit/cmd/gateway

get: FORCE
	go build -ldflags="$(EXTRAGOLDFLAGS)" -o $@ anachronauts.club/repos/gmikit/cmd/get

install: all
	install -Minstall.log -d $(BINDIR) $(SBINDIR) $(GMIKITCONFDIR)
	install -Minstall.log -m 755 convert $(BINDIR)/$(BINPREFIX)convert
	install -Minstall.log -m 755 gateway $(SBINDIR)/$(BINPREFIX)gateway
	install -Minstall.log -m 755 get $(BINDIR)/$(BINPREFIX)get
	install -Minstall.log -m 644 example/gateway.conf $(GMIKITCONFDIR)/gateway.conf.sample

package:
	-rm install.log
	$(MAKE) PREFIX=$(STAGE)$(PREFIX) install
	cat MANIFEST > $(STAGE)/+MANIFEST
	echo prefix $(PREFIX) >> $(STAGE)/+MANIFEST
	awk '/type=file/ { print substr($$1, index($$1, "$(PREFIX)")) }' install.log > $(STAGE)/plist
	mkdir -p $(PKGDIR)
	pkg create -o "$(PKGDIR)" -r "$(STAGE)" -M "$(STAGE)/+MANIFEST" -p "$(STAGE)/plist"
	pkg repo "$(PKGDIR)"

