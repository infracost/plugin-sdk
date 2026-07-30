# Infracost Plugin SDK

Build custom plugins to extend [Infracost](https://infracost.io) with new IaC formats and cloud providers.

Infracost has two types of plugins:

| | Parser Plugin | Provider Plugin |
|---|---|---|
| **Purpose** | Extract resources from IaC files into a cost tree | Price the cost tree and apply policies |
| **Plugin type** | `GetPluginInfo` reports `PARSER` | `GetPluginInfo` reports `PROVIDER` |
| **Services** | `PluginService` + `ParserService` | `PluginService` + `ProviderService` |
| **RPCs** | GetPluginInfo · GetParserConfig · IdentifyProjects · IdentifyEnvironments *(optional)* · Parse | GetPluginInfo · Process · ListFinopsPolicies |
| **Examples** | Terraform, Terragrunt, CloudFormation | AWS, Azure, GCP |
| **Use case** | "I have a new IaC format" | "I have a new cloud to price" |

Every plugin implements `PluginService` (a single `GetPluginInfo` RPC that reports its type and metadata) plus one of `ParserService` / `ProviderService`. Both services share one gRPC handshake.

## Getting started

- **Parser plugins** — See [parser/README.md](parser/README.md) and [parser/SPEC.md](parser/SPEC.md)
- **Provider plugins** — See [provider/README.md](provider/README.md) and [provider/SPEC.md](provider/SPEC.md)

Each directory contains a working example you can copy as a starting point.

## How plugins are discovered

The CLI does **not** infer a plugin's type or identity from its binary filename. It launches every executable in the plugin directory, calls `GetPluginInfo`, and uses the returned `type` (`PARSER` / `PROVIDER`) to decide how to use it. Binaries that fail to launch or handshake are skipped.

The plugin directory defaults to `os.UserCacheDir()/infracost/plugins`:

- Linux: `~/.cache/infracost/plugins`
- macOS: `~/Library/Caches/infracost/plugins`
- Windows: `%LocalAppData%\infracost\plugins`

A descriptive binary name such as `infracost-parser-<name>` or `infracost-provider-<name>` is conventional and recommended for clarity, but it is not required for discovery.

## Plugin naming

The `name` returned by `GetPluginInfo` is the plugin's identity, by convention `<namespace>/<name>`:

| Example | Meaning |
|---|---|
| `infracost/terraform` | Official plugin (the `infracost/` namespace is reserved for official plugins) |
| `acme/crossplane` | Community plugin |

Names must be unique across all installed plugins.

## The handshake

All plugins — parser and provider — use the same go-plugin handshake. Because the type is resolved at runtime via `GetPluginInfo`, there is a single magic cookie:

```go
goplugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "INFRACOST_PLUGIN",
    MagicCookieValue: "de8c7e96-497c-4168-80c4-fc875c8ce764",
}
```

The dispense key is always `"plugin"`, and both `PluginService` and the relevant parser/provider service are registered on the same gRPC server. See the per-type SPECs for the full serving code.

## Building and installing a plugin

```bash
# From an example directory
go build -o infracost-parser-myformat .

# Install it where the CLI discovers plugins (see "How plugins are discovered")
make install
```

Then run `infracost` against a project — the CLI launches the binary, calls `GetPluginInfo`, and routes work to it based on the reported type.
