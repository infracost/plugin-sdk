# Infracost Parser Plugin SDK

Build custom parser plugins to teach [Infracost](https://infracost.io) how to read new Infrastructure-as-Code formats.

## What is a parser plugin?

Infracost uses parser plugins to extract resources from IaC files (Terraform, CloudFormation, etc.) into an IaC-agnostic cost tree. Each plugin is a standalone binary that communicates with the Infracost CLI over gRPC.

The CLI discovers plugins by launching every binary in its plugin directory and calling `GetPluginInfo`. Plugins that report `type: PARSER` are then asked, in identification-priority order, which paths in each scanned directory they can parse. The plugin that claims a path parses it into a cost tree for the provider plugins to price.

## Quick start

This SDK ships two starting points — pick one:

- **[`example/`](example)** — a single self-contained `main.go`. Read this
  first to see the whole contract (handshake, both services, all four RPCs)
  in one file with no package split.
- **[`template/`](template)** — the production-shaped starting point, split
  one file per RPC (`server/get_plugin_info.go`, `server/parse.go`, etc.) with
  its own tests and `testdata/` fixtures. Copy this when you're ready to build
  a real plugin.

1. Copy [`template/`](template) (or [`example/`](example) for something
   smaller) as your starting point
2. Update `GetPluginInfo()` with your plugin's name and metadata
3. Implement `IdentifyProjects()` to recognise your IaC files in a directory
4. Implement `Parse()` to extract resources into a `tree.Tree`
5. Build and install the binary into the plugin directory

### Build the example

```bash
cd example
go build -o infracost-parser-example .
```

### Install it where the CLI can find it

The CLI scans `os.UserCacheDir()/infracost/plugins` (Linux `~/.cache/...`, macOS `~/Library/Caches/...`). The `Makefile` has an `install` target that copies the binary there:

```bash
make install
```

Then run `infracost` against a project containing your format.

## Interface contract

Your plugin implements two gRPC services:

**PluginService**

| RPC | Purpose |
|-----|---------|
| **GetPluginInfo** | Report the plugin type (`PARSER`) and metadata (name, version, description) |

**ParserService**

| RPC | Purpose |
|-----|---------|
| **GetParserConfig** | Report identification priority and project-type mapping |
| **IdentifyProjects** | Report which paths in a directory this plugin can parse (no recursion) |
| **IdentifyEnvironments** *(optional)* | Refine a project into named environments; return `codes.Unimplemented` if not supported |
| **Parse** | Parse a path into an IaC-agnostic `tree.Tree` |

Both services are registered on the same gRPC server using a shared handshake. See [SPEC.md](SPEC.md) for the full specification, including the handshake, message formats, priority semantics, and the cost-tree structure.

## Testing

The plugin contract is plain Go gRPC, so the most reliable way to test is with Go unit tests that call your service methods directly with `testdata/` fixtures. [`template/`](template) shows the pattern (`server/*_test.go`, in-process via go-plugin's `TestPluginGRPCConn`); the reference plugins in the [infracost/parser](https://github.com/infracost/parser) repo follow the same pattern.

To try it end to end, install the binary in the plugin directory (`make install`) and run `infracost` against a project.

## Adding a new IaC format

Most formats can pass their format-specific options as JSON via `ParseRequest.raw_options`, so no proto changes are needed. See the "Adding a New Format" section in [SPEC.md](SPEC.md).

## Reference

- [SPEC.md](SPEC.md) — Full plugin interface specification
- [example/](example) — Minimal working plugin (single file)
- [template/](template) — Production-shaped starting point (one file per RPC, with tests)
- [infracost/parser](https://github.com/infracost/parser) — Production plugin implementations
