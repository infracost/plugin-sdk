# Infracost Provider Plugin SDK

Build custom provider plugins to teach [Infracost](https://infracost.io) how to price resources from any cloud or service.

## What is a provider plugin?

Infracost uses provider plugins to turn the IaC-agnostic cost tree (produced by a parser plugin) into cost estimates and FinOps policy results. Each plugin is a standalone binary that communicates with the Infracost CLI over gRPC. A provider is completely free to decide how it obtains prices — hardcode them, query a cloud pricing API, use a local database, or call the Infracost Cloud API.

The CLI discovers plugins by launching every binary in its plugin directory and calling `GetPluginInfo`. Plugins that report `type: PROVIDER` are handed the cost tree to price.

## Quick start

1. Copy the [`example/`](example) directory as your starting point
2. Update `GetPluginInfo()` with your plugin's name and metadata
3. Implement `Process()` to walk the cost tree and return cost components
4. (Optional) Implement `ListFinopsPolicies()` to expose FinOps policies
5. Build and install the binary into the plugin directory

### Build the example

```bash
cd example
go build -o infracost-provider-plugin-example .
```

### Install it where the CLI can find it

The CLI scans `os.UserCacheDir()/infracost/plugins` (Linux `~/.cache/...`, macOS `~/Library/Caches/...`). The `Makefile` has an `install` target that copies the binary there:

```bash
make install
```

Then run `infracost` against a project.

## Interface contract

Your plugin implements two gRPC services:

**PluginService**

| RPC | Purpose |
|-----|---------|
| **GetPluginInfo** | Report the plugin type (`PROVIDER`) and metadata (name, version, description) |

**ProviderService**

| RPC | Purpose |
|-----|---------|
| **Process** | Walk the cost tree and return priced resources + policy results |
| **ListFinopsPolicies** | Return available FinOps policy definitions |

Both services are registered on the same gRPC server using a shared handshake. See [SPEC.md](SPEC.md) for the full specification, including the handshake, the `TreeInput`/`Output` message formats, and the `CostComponent` contract.

## Testing

The plugin contract is plain Go gRPC, so the most reliable way to test is with Go unit tests that build a `TreeInput` and call `Process` / `ListFinopsPolicies` directly. The reference plugins in the [infracost/providers](https://github.com/infracost/providers) repo follow this pattern.

To try it end to end, install the binary in the plugin directory (`make install`) and run `infracost` against a project.

## Reference

- [SPEC.md](SPEC.md) — Full plugin interface specification
- [example/](example) — Minimal working plugin with hardcoded pricing
- [infracost/providers](https://github.com/infracost/providers) — Production provider implementations
