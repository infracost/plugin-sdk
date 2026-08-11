# Infracost Parser Plugin Specification

This document defines the interface contract for Infracost parser plugins. A parser plugin teaches Infracost how to read a new Infrastructure-as-Code format (e.g., Pulumi, Crossplane, CDK for Terraform) and translate it into the IaC-agnostic cost tree that the Infracost providers consume.

## Architecture Overview

```
infracost scan
    |
    v
Plugin Manager (launches every binary in the plugin directory)
    |
    v
GetPluginInfo()  — identify the plugin and its type (PARSER / PROVIDER)
    |
    v
For each parser plugin (ordered by identification priority):
    1. GetParserConfig()  — priority + project-type mapping
    2. IdentifyProjects() — which paths in a directory this plugin parses
    3. Parse()            — extract resources into an IaC-agnostic tree
    |
    v
Provider plugins (AWS, Azure, GCP) price the tree
```

Plugins are standalone binaries that communicate with the Infracost CLI over gRPC using the [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin) framework. The CLI spawns the plugin process, performs a handshake, and issues gRPC calls. When work is complete the plugin process exits.

Every plugin implements two gRPC services:

- **`PluginService`** — a single `GetPluginInfo` RPC that reports the plugin's type and metadata. This is how the CLI tells parser plugins apart from provider plugins.
- **`ParserService`** — the parsing RPCs (`GetParserConfig`, `IdentifyProjects`, `IdentifyEnvironments`, `Parse`).

Both services are registered on the same gRPC server (see [Handshake](#go-plugin-handshake)).

## Plugin Identity and Naming

The CLI **does not** infer a plugin's type or identity from its binary filename. It launches every executable in the plugin directory, calls `GetPluginInfo`, and uses the returned `type` to decide whether the binary is a parser or a provider. Binaries that fail to launch or handshake are **skipped** (logged at debug level, not fatal). A duplicate `(name, type)` pair is **fatal** — the CLI kills all already-loaded plugins and exits with an error.

A descriptive binary name such as `infracost-parser-<format>` is conventional and recommended for clarity, but it is not required for discovery.

The `name` returned by `GetPluginInfo` is the plugin's identity. By convention it is `<namespace>/<name>`:

- Official plugins use the `infracost/` namespace, e.g. `infracost/terraform`, `infracost/cloudformation`.
- Community plugins should use their own namespace, e.g. `acme/crossplane`.

Names must be unique **within a type** — a parser and a provider may report the same name (e.g. `infracost/kubernetes`) without conflict.

## go-plugin Handshake

All Infracost plugins — parser and provider alike — share one handshake. The type is resolved at runtime via `GetPluginInfo`, so there is a single magic cookie rather than a per-type one.

```go
const maxMessageSize = 64 * 1024 * 1024

goplugin.Serve(&goplugin.ServeConfig{
    HandshakeConfig: goplugin.HandshakeConfig{
        ProtocolVersion:  1,
        MagicCookieKey:   "INFRACOST_PLUGIN",
        MagicCookieValue: "de8c7e96-497c-4168-80c4-fc875c8ce764",
    },
    Plugins: map[string]goplugin.Plugin{
        // The dispense key is always "plugin".
        "plugin": yourPluginStruct,
    },
    GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
        opts = append(opts,
            grpc.MaxRecvMsgSize(maxMessageSize),
            grpc.MaxSendMsgSize(maxMessageSize),
        )
        return grpc.NewServer(opts...)
    },
})
```

The plugin struct must implement `plugin.GRPCPlugin` and register **both** services on the server:

```go
type myPlugin struct {
    goplugin.NetRPCUnsupportedPlugin
}

func (p *myPlugin) GRPCServer(_ *goplugin.GRPCBroker, g *grpc.Server) error {
    pluginpb.RegisterPluginServiceServer(g, myServiceImplementation)
    pluginpb.RegisterParserServiceServer(g, myServiceImplementation)
    return nil
}

func (p *myPlugin) GRPCClient(_ context.Context, _ *goplugin.GRPCBroker, _ *grpc.ClientConn) (interface{}, error) {
    return nil, fmt.Errorf("not implemented")
}
```

The generated Go bindings live in `github.com/infracost/proto/gen/go/infracost/plugin` (aliased `pluginpb` above). Embedding `pluginpb.UnimplementedPluginServiceServer` and `pluginpb.UnimplementedParserServiceServer` in your service struct keeps it forward-compatible if new RPCs are added.

## gRPC Service Contract

Plugins implement `PluginService` and `ParserService`, both defined in the `infracost.plugin` package:

```protobuf
// infracost/plugin/plugin.proto
service PluginService {
    rpc GetPluginInfo(GetPluginInfoRequest) returns (GetPluginInfoResponse);
}

// infracost/plugin/parser.proto
service ParserService {
    rpc GetParserConfig(GetParserConfigRequest) returns (GetParserConfigResponse);
    rpc IdentifyProjects(IdentifyProjectsRequest) returns (IdentifyProjectsResponse);
    rpc IdentifyEnvironments(IdentifyEnvironmentsRequest) returns (IdentifyEnvironmentsResponse); // optional
    rpc Parse(ParseRequest) returns (ParseResponse);
}
```

### GetPluginInfo

Returns the plugin's type and static metadata. Called once at startup when the CLI discovers the binary.

**Request:** Empty.

**Response:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | PluginType | yes | `PARSER` for parser plugins. (`PROVIDER` and `PLUGIN_TYPE_UNKNOWN` are the other values.) |
| `name` | string | yes | Plugin identity, e.g. `"infracost/terraform"` or `"acme/crossplane"`. Must be unique. |
| `version` | string | no | Plugin version. Semver recommended but not required. |
| `description` | string | no | Human-readable description shown in CLI output. |
| `url` | string | no | Where the plugin is documented/downloaded. |
| `author` | string | no | Company or individual that authored the plugin. |

### GetParserConfig

Returns settings the CLI uses to decide if and how this plugin participates in project identification. Called once after `GetPluginInfo`.

**Request:** Empty.

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `identification_priority` | uint32 | Order in which plugins are offered a directory. **Higher = checked first.** Recommended default is `0`. |
| `config_file_project_type` | string (optional) | Maps to an `infracost/config` project type, e.g. `"terraform"`, `"cloudformation"`, `"terragrunt"`. Custom strings are allowed. If unset, defaults to the plugin name. |

#### Priority

When several plugins could claim the same directory, `identification_priority` determines the order in which they are asked. Plugins with a **higher** priority are offered the directory first; the first plugin to identify a project for a given path wins it.

Leave this at `0` unless your format must take precedence over another. For example, the Terragrunt plugin uses `1` so it is offered a directory before the Terraform plugin (`0`), because a Terragrunt project may contain Terraform files that should not be parsed as standalone Terraform.

### IdentifyProjects

Inspects a single directory and reports which paths this plugin can parse. Called for each directory the CLI scans, in identification-priority order.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `directory` | string | Absolute path to a directory to inspect. |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `directory` | bool | `true` if the whole directory is a single project of this format. Mutually exclusive with `files`. |
| `files` | string[] | Individual files in the directory that are each a project in their own right (paths relative to `directory`). Must be empty if `directory` is `true`. |
| `dependency_paths` | string[] | Paths (relative to `directory`) that this project depends on. Optional. |
| `raw_options` | bytes (JSON) | Seed options blob for this project, owned by the plugin. The CLI persists this in the config file and passes it back verbatim in subsequent `ParseRequest.raw_options` calls. |

**Contract:**
- **Do not recurse.** Inspect only the entries directly inside `directory`; the CLI walks the tree and calls `IdentifyProjects` per directory.
- Return an empty response (not an error) for directories you don't handle or can't read.
- Identification must be fast. Avoid network calls.
- Use `directory: true` for directory-oriented formats (Terraform, Terragrunt). Use `files` for file-oriented formats where each file is an independent project (CloudFormation, Kubernetes).

### IdentifyEnvironments (optional)

Refines a project identified by `IdentifyProjects` into one or more named environments (e.g., Terraform workspaces or variable-file sets). This RPC is **optional** — if your plugin does not support environments, return `codes.Unimplemented` and the CLI will treat the project as a single default environment. This is distinct from returning an empty list, which means the project has zero environments and will not be parsed.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `directory` | string | Absolute path to the project directory. |
| `attributed_files` | AttributedVarFile[] | Variable files attributed to this project by the Terraform/Terragrunt autodetect flow. Provided as a migration aid — non-Terraform/Terragrunt plugins should ignore this field. |
| `raw_options` | bytes (JSON) | Options blob seeded by `IdentifyProjects`. |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `environments` | Environment[] | List of environments for this project. Each `Environment` has a `name`, `path`, `files[]`, `dependency_paths[]`, and `raw_options` (refined per-environment). |

### Parse

Parses the IaC at the given path and returns an IaC-agnostic cost tree.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Absolute path to the project file or directory to parse. |
| `generic_options` | GenericOptions | IaC-agnostic options: working/repo directories, cache settings, credential sets, an optional `dependency_request`, etc. See `infracost/parser/options/options.proto`. |
| `raw_options` | bytes (JSON) | Plugin-specific options. **Always JSON** (proto field 4 is reserved and dropped). The schema is owned by the plugin; document what your plugin expects. |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `tree` | tree.Tree | The IaC-agnostic cost tree (providers → services → resources). See `infracost/tree/tree.proto`. |
| `diagnostics` | Diagnostic[] | Warnings and errors encountered during parsing. |
| `requested_dependencies` | Dependency[] | Dependencies extracted when `generic_options.dependency_request` was set (optional, for IDE features). |

**Contract:**
- Build the `tree.Tree` from your parsed resources. The tree is the single output format — there is no separate format-specific result.
- Return partial results with diagnostics on recoverable errors (e.g., one file in a directory fails to parse). Add a `critical` diagnostic for failures that prevented parsing.
- Return an error only for unrecoverable failures (e.g., `path` is empty).
- Set `is_supported` on a tree resource if a provider plugin can price it.

#### The cost tree

`tree.Tree` is a hierarchy of providers → services → resources:

```
Tree
 └── providers: map<string, Provider>   // e.g. "aws", "azure", "google"
      └── services: map<string, Service>
           └── resources: []Resource    // id, type, region, is_supported, attributes, tags, ...
```

The wire format is generated from `infracost/tree/tree.proto`. The proto comments recommend building it with the `tree` package's `ToProto()` / `FromProto()` helpers rather than constructing the protobuf messages by hand where those helpers are available to you.

## `raw_options` lifecycle

`raw_options` is the channel through which a plugin seeds, refines, and receives its own format-specific configuration:

1. **`IdentifyProjects`** — the plugin returns a `raw_options` blob (JSON) alongside the identified paths. This is the initial seed, e.g. `{"vars_file": "prod.tfvars"}`.
2. **`IdentifyEnvironments`** (optional) — the plugin refines the seed per environment; each returned `Environment` carries its own `raw_options`.
3. **Config file** — the CLI persists the blob as a readable YAML map in the project's `infracost.yml` (or equivalent). Users can edit it; the CLI reads it back on subsequent runs.
4. **`Parse`** — the CLI passes the current blob verbatim as `ParseRequest.raw_options` (always JSON). The plugin parses it and uses it to configure the parse.

`raw_options` is always JSON — proto field 4 is reserved and dropped. The schema is entirely owned by the plugin; document it in your plugin's README.

## Adding a New Format

To add support for a completely new IaC format:

1. **Decide how options are passed.** Pass format-specific options as JSON in `ParseRequest.raw_options` — no proto changes needed. Document the JSON schema your plugin expects.

2. **Build the plugin binary:**
   - Create a `main.go` that serves `PluginService` + `ParserService` over the handshake above.
   - Implement `GetPluginInfo` (returning `type: PARSER`), `GetParserConfig`, `IdentifyProjects`, and `Parse`.
   - Have `Parse` produce a `tree.Tree`.

3. **Install** the binary in the plugin directory (see [Installing and testing](#installing-and-testing)).

No registration with Infracost is required — a plugin only needs to be discoverable in the plugin directory.

## Installing and testing

The CLI discovers plugins by scanning a plugin directory, which defaults to `os.UserCacheDir()/infracost/plugins`:

- Linux: `~/.cache/infracost/plugins`
- macOS: `~/Library/Caches/infracost/plugins`
- Windows: `%LocalAppData%\infracost\plugins`

Drop your built binary in that directory and run `infracost` against a project containing your format. The CLI will launch the binary, call `GetPluginInfo`, and route matching directories to it.

Run `infracost plugin list` to verify the binary is visible and reporting the expected name and version.

### Environment variable overrides

| Variable | Default | Purpose |
|---|---|---|
| `INFRACOST_CLI_PLUGIN_DIR` | *(cache dir)* | Load plugins from this directory instead of the cache (downloads skipped; for local development). |
| `INFRACOST_CLI_PLUGIN_CACHE_DIRECTORY` | `os.UserCacheDir()/infracost/plugins` | Download location for managed plugins. |
| `INFRACOST_CLI_PLUGIN_AUTO_UPDATE` | `true` | Set to `false` to disable automatic updates. |

Because the gRPC contract is plain Go, the most reliable way to test a plugin is with Go unit tests that call your service methods directly — see [`template/server/*_test.go`](template/server) in this repo for the pattern (table-driven tests with `testdata/` fixtures).

## Constraints and Limits

- **Max gRPC message size**: 64 MB (both send and receive)
- **Plugin start timeout**: 60 s (Linux/macOS), 180 s (Windows). The plugin must serve the handshake within this window or the CLI kills the process and skips it.
- **`GetPluginInfo` query timeout**: 30 s. Called during install/update to read the installed version; also called at discovery on every startup.
- **Identification**: must be fast and side-effect free; avoid network calls
- **Binary size**: no hard limit, but aim for fast downloads
- **Concurrency**: the CLI may run plugins in parallel; plugins are separate processes, so keep any shared on-disk state (caches) concurrency-safe

### Diagnosing discovery failures

If `infracost plugin list` doesn't show your binary, re-run infracost with `--log-level debug` (or set `LOG_LEVEL=debug`). The CLI logs a skip reason for each binary it rejects. Common causes: binary not executable (`chmod +x`), wrong handshake constants, or startup timeout exceeded. The CLI propagates `LOG_LEVEL` to the plugin subprocess, so your plugin's own logger will emit debug output at the same level.

## Reference Implementations

For a minimal, single-file starting point, see [`example/`](example) in this repo. For a production-shaped starting point (one file per RPC, with tests), see [`template/`](template) — copy it directly to start a real plugin.

The official Infracost parser plugins follow the same shape as `template/` (a `main.go` plus a `server/` package with one file per RPC). Their identification behaviour, observable via `infracost plugin list` and a `--log-level debug` run, is a useful calibration point for your own:
- **Terraform** and **Terragrunt** are directory-based: they claim whole directories.
- **CloudFormation**, **Kubernetes**, **CiscoStacks**, and **ARM** are file-based: they content-sniff individual files.
- **Terragrunt** registers identification priority 1 (above Terraform's 0); **Terraform-plan** registers 10, so a plan file always wins over the Terraform directory containing it. All others use the default 0.
