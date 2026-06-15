.PHONY: build install clean

BINARY = infracost-parser-plugin-example

# Where the Infracost CLI discovers plugins. The default is os.UserCacheDir()/
# infracost/plugins — on Linux ~/.cache/infracost/plugins, on macOS
# ~/Library/Caches/infracost/plugins. Override if needed.
PLUGIN_DIR ?= $(HOME)/.cache/infracost/plugins

build:
	go build -o $(BINARY) .

# Install the plugin where the CLI can find it. The CLI launches every binary
# in the plugin directory and calls GetPluginInfo to determine its type, so the
# executable name is up to you.
install: build
	mkdir -p "$(PLUGIN_DIR)"
	cp $(BINARY) "$(PLUGIN_DIR)/"

clean:
	rm -f $(BINARY)
