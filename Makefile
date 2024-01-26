.PHONY: clean gui gui-install gui-symlinks install install-gui install-symlinks ko ko-install og og-install symlinks symlinks-install vg-symlink

PROJECT ?= orbiton

GOFLAGS := -mod=vendor -trimpath -v
GOBUILD := go build

SRCFILES := $(wildcard go.* v2/*.go v2/go.*)


# macOS and FreeBSD detection
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
  PREFIX ?= /usr/local
  MAKE ?= make
else ifeq ($(UNAME_S),FreeBSD)
  PREFIX ?= /usr/local
  MAKE ?= gmake
else
  PREFIX ?= /usr
  MAKE ?= make
endif

MANDIR ?= $(PREFIX)/share/man/man1

UNAME_R ?= $(shell uname -r)
ifneq (,$(findstring arch,$(UNAME_R)))
# Arch Linux
LDFLAGS ?= -Wl,-O2,--sort-common,--as-needed,-z,relro,-z,now
GOFLAGS += -buildmode=pie
BUILDFLAGS ?= -ldflags "-s -w -linkmode=external -extldflags $(LDFLAGS)"
endif

CXX ?= g++

CXXFLAGS ?= -O2 -pipe -fPIC -fno-plt -fstack-protector-strong -Wall -Wshadow -Wpedantic -Wno-parentheses -Wfatal-errors -Wvla -Wignored-qualifiers -pthread $(LDFLAGS)
CXXFLAGS += -DGDK_DISABLE_DEPRECATED -DGTK_DISABLE_DEPRECATED

ifeq ($(UNAME_S),Darwin)
  CXXFLAGS += -std=c++20
else
  CXXFLAGS += -Wl,--as-needed
endif

o: $(SRCFILES)
	cd v2 && $(GOBUILD) $(GOFLAGS) $(BUILDFLAGS) -o ../o

trace: clean $(SRCFILES)
	cd v2 && $(GOBUILD) $(GOFLAGS) $(BUILDFLAGS) -tags=trace -o ../o

nopdf: clean $(SRCFILES)
	cd v2 && $(GOBUILD) $(GOFLAGS) $(BUILDFLAGS) -tags=nopdf -o ../o

pgo: v2/default.pgo

v2/default.pgo: clean $(SRCFILES)
	cd v2 && $(GOBUILD) $(GOFLAGS) $(BUILDFLAGS) -tags=trace -o ../o
	-rm v2/default.pgo
	@# v2/main.go could be any filename, it's just for collecting the CPU profile info
	./o --cpuprofile v2/default.pgo v2/main.go

bench:
	cd v2 && go test -bench=. -benchmem

gui: og
ko: og
og: gtk3/gtk3
gtk3: gtk3/gtk3

gtk3/gtk3: gtk3/main.cpp
	$(CXX) "$<" -o "$@" $(CXXFLAGS) $(shell pkg-config --cflags --libs vte-2.91) $(LDFLAGS)

o.1.gz: o.1
	gzip -f -k o.1

install: o o.1.gz
	mkdir -p "$(DESTDIR)$(PREFIX)/bin"
	install -m755 o "$(DESTDIR)$(PREFIX)/bin/o"
	mkdir -p "$(DESTDIR)$(MANDIR)"
	install -m644 o.1.gz "$(DESTDIR)$(MANDIR)/o.1.gz"

gui-install: install-gtk3
install-gui: install-gtk3
ko-install: install-gtk3
install-ko: install-gtk3
og-install: install-gtk3
install-og: install-gtk3

install-gtk3: gtk3/gtk3 vg-symlink
	mkdir -p "$(DESTDIR)$(PREFIX)/bin"
	install -m755 gtk3/gtk3 "$(DESTDIR)$(PREFIX)/bin/og"
	mkdir -p "$(DESTDIR)$(PREFIX)/share/pixmaps"
	install -m644 img/og.png "$(DESTDIR)$(PREFIX)/share/pixmaps/og.png"
	install -m644 img/lig.png "$(DESTDIR)$(PREFIX)/share/pixmaps/lig.png"
	mkdir -p "$(DESTDIR)$(PREFIX)/share/applications"
	install -m644 gtk3/og.desktop "$(DESTDIR)$(PREFIX)/share/applications/og.desktop"
	install -m644 gtk3/lig.desktop "$(DESTDIR)$(PREFIX)/share/applications/lig.desktop"

install-symlinks: symlinks
symlinks-install: symlinks

symlinks:
	mkdir -p "$(DESTDIR)$(PREFIX)/bin"
	ln -s -f "$(PREFIX)/bin/o" "$(DESTDIR)$(PREFIX)/bin/li"
	ln -s -f "$(PREFIX)/bin/o" "$(DESTDIR)$(PREFIX)/bin/redblack"
	ln -s -f "$(PREFIX)/bin/o" "$(DESTDIR)$(PREFIX)/bin/sw"
	ln -s -f "$(PREFIX)/bin/o" "$(DESTDIR)$(PREFIX)/bin/edi"
	ln -s -f "$(PREFIX)/bin/o" "$(DESTDIR)$(PREFIX)/bin/vs"
	ln -s -f "$(PREFIX)/bin/o" "$(DESTDIR)$(PREFIX)/bin/teal"

# For pico/nano style editing
nano-symlink:
	ln -s -f "$(PREFIX)/bin/o" "$(DESTDIR)$(PREFIX)/bin/nano"

# For pico/nano style editing, but "nan" does not conflict with "nano".
nan-symlink:
	ln -s -f "$(PREFIX)/bin/o" "$(DESTDIR)$(PREFIX)/bin/nan"

symlinks-gui: gui-symlinks
symlinks-gui-install: gui-symlinks
symlinks-install-gui: gui-symlinks
install-symlinks-gui: gui-symlinks

vg-symlink:
	mkdir -p "$(DESTDIR)$(PREFIX)/bin"
	ln -s -f "$(PREFIX)/bin/og" "$(DESTDIR)$(PREFIX)/bin/vg"

gui-symlinks: vg-symlink
	mkdir -p "$(DESTDIR)$(PREFIX)/bin"
	ln -s -f "$(PREFIX)/bin/og" "$(DESTDIR)$(PREFIX)/bin/lig"
	ln -s -f "$(PREFIX)/bin/og" "$(DESTDIR)$(PREFIX)/bin/redblackg"
	ln -s -f "$(PREFIX)/bin/og" "$(DESTDIR)$(PREFIX)/bin/swg"
	ln -s -f "$(PREFIX)/bin/og" "$(DESTDIR)$(PREFIX)/bin/edg"
	ln -s -f "$(PREFIX)/bin/og" "$(DESTDIR)$(PREFIX)/bin/tealg"

easteregg:
	mkdir -p "$(DESTDIR)$(PREFIX)/bin"
	ln -s -f "$(PREFIX)/bin/o" "$(DESTDIR)$(PREFIX)/bin/feedgame"

gui-easteregg: easteregg-gui

easteregg-gui:
	mkdir -p "$(DESTDIR)$(PREFIX)/bin"
	ln -s -f "$(PREFIX)/bin/og" "$(DESTDIR)$(PREFIX)/bin/feedgameg"

license:
	mkdir -p "$(DESTDIR)$(PREFIX)/share/licenses/$(PROJECT)"
	install -m644 LICENSE "$(DESTDIR)$(PREFIX)/share/licenses/$(PROJECT)/LICENSE"

clean:
	-rm -f o o.1.gz gtk3/gtk3 v2/o v2/orbiton
