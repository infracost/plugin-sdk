.PHONY: build test install clean

BINARY = infracost-parser-example

# Where the Infracost CLI discovers plugins. This mirrors Go's os.UserCacheDir():
# macOS uses ~/Library/Caches, everything else uses $XDG_CACHE_HOME or ~/.cache.
# Override PLUGIN_DIR if needed.
ifeq ($(shell uname),Darwin)
CACHE_DIR ?= $(HOME)/Library/Caches
else
CACHE_DIR ?= $(if $(XDG_CACHE_HOME),$(XDG_CACHE_HOME),$(HOME)/.cache)
endif
PLUGIN_DIR ?= $(CACHE_DIR)/infracost/plugins

build:
	go build -o $(BINARY) .

test:
	go test ./...

# Install the plugin where the CLI can find it. The CLI launches every binary
# in the plugin directory and calls GetPluginInfo to determine its type, so the
# executable name is up to you.
install: build
	mkdir -p "$(PLUGIN_DIR)"
	cp $(BINARY) "$(PLUGIN_DIR)/"

clean:
	rm -f $(BINARY)
