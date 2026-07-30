# Infracost Plugin SDK

Build custom plugins to extend [Infracost](https://infracost.io) with new IaC formats and cloud providers.

Infracost has two types of plugins:

| | Parser Plugin | Provider Plugin |
|---|---|---|
| **Purpose** | Extract resources from IaC files into a cost tree | Price the cost tree and apply policies |
| **Plugin type** | `GetPluginInfo` reports `PARSER` | `GetPluginInfo` reports `PROVIDER` |
| **Services** | `PluginService` + `ParserService` | `PluginService` + `ProviderService` |
| **RPCs** | GetPluginInfo · GetParserConfig · IdentifyProjects · IdentifyEnvironments *(optional)* · Parse | GetPluginInfo · Process · ListFinopsPolicies |
| **Examples** | Terraform, Terragrunt, CloudFormation, Kubernetes, CiscoStacks, Terraform-plan | AWS, Azure, Google, Kubernetes |
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

Binaries that fail to launch, fail the handshake, or fail to serve `GetPluginInfo` are **skipped** — logged at debug level, not fatal. A duplicate `(name, type)` pair (two parser plugins both reporting `infracost/kubernetes`, for example) is **fatal**: the CLI kills all already-loaded plugins and exits with an error. A parser and a provider that share the same name (e.g. `infracost/kubernetes`) are different identities and are both allowed.

### Environment variables

Override defaults without touching config files:

| Variable | Default | Purpose |
|---|---|---|
| `INFRACOST_CLI_PLUGIN_DIR` | *(cache directory)* | Load plugins from this flat directory instead of the cache. Downloads are skipped; useful for local development. |
| `INFRACOST_CLI_PLUGIN_CACHE_DIRECTORY` | `os.UserCacheDir()/infracost/plugins` | Where managed (required) plugins are downloaded to. |
| `INFRACOST_CLI_PLUGIN_BASE_URL` | `https://releases.infracost.io` | Root URL for plugin archive downloads. |
| `INFRACOST_CLI_PLUGIN_AUTO_UPDATE` | `true` | Set to `false` to disable automatic updates of required plugins. |

## Plugin naming

The `name` returned by `GetPluginInfo` is the plugin's identity, by convention `<namespace>/<name>`:

| Example | Meaning |
|---|---|
| `infracost/terraform` | Official plugin (the `infracost/` namespace is reserved for official plugins) |
| `acme/crossplane` | Community plugin |

Names must be unique **within a type** — a parser and a provider may report the same name (e.g. `infracost/kubernetes`) without conflict.

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

## Managing plugins

```bash
infracost plugin list    # show installed plugins and their versions
infracost plugin update  # download the latest versions of all required plugins
```

Use `infracost plugin list` to confirm a newly installed binary is visible to the CLI and showing the expected name/version. If a binary does not appear in the list, it failed discovery — see [Diagnosing discovery failures](#diagnosing-discovery-failures) below.

## Diagnosing discovery failures

If a plugin binary is not listed by `infracost plugin list`, one of the following is likely:

- **Not executable** — run `chmod +x <binary>` (Linux/macOS).
- **Handshake failed** — the binary started but used wrong cookie/protocol values. Check the handshake constants in your `main.go`.
- **Timed out at startup** — plugins have 60 s (Linux/macOS) or 180 s (Windows) to become ready. A slow binary or missing dependency can exceed this limit.

Re-run with `--log-level debug` (or set `LOG_LEVEL=debug`) to see per-binary skip reasons. The CLI propagates `LOG_LEVEL` to the plugin subprocess, so your plugin's own `LOG_LEVEL`-aware logger will emit debug output at the same level.
