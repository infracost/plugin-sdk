# Infracost Parser Plugin Specification

This document defines the interface contract for Infracost parser plugins. A parser plugin teaches Infracost how to read a new Infrastructure-as-Code format (e.g., Pulumi, Crossplane, CDK for Terraform) and translate it into the cost-estimation tree that the Infracost providers consume.

## Architecture Overview

```
infracost scan
    |
    v
Plugin Manager (discovers per-IaC binaries in plugin directory)
    |
    v
For each project path:
    1. Describe() — get plugin metadata (name, priority, extensions)
    2. Detect()   — ask each plugin (by priority) if it handles this path
    3. Initialize() — pass supported resource types
    4. Parse() or ParseToTree() — extract resources from the IaC files
    |
    v
Provider plugins (AWS, Azure, GCP) price the extracted resources
```

Plugins are standalone binaries that communicate with the Infracost CLI over gRPC using the [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin) framework. The CLI spawns the plugin process, performs a handshake, and issues gRPC calls. When parsing is complete the plugin process exits.

## Plugin Naming

Every plugin has a **canonical name** in `registry/namespace/name` format:

```
plugins.infracost.io/infracost/terraform
plugins.infracost.io/acme/crossplane
plugins.example.com/acme/pulumi
```

Resolution rules (short forms expand to the full canonical name):
- `terraform` → `plugins.infracost.io/infracost/terraform` (official plugin, default registry)
- `acme/crossplane` → `plugins.infracost.io/acme/crossplane` (community plugin, default registry)
- `plugins.example.com/acme/pulumi` → as-is (custom registry, first segment contains a dot)

The `infracost/` namespace is reserved for official plugins.

The canonical name is returned by the `Describe` RPC and is the source of truth for plugin identity. It is used in dependency declarations (`requires`) and internal tracking.

## Binary Naming Convention

Plugin binaries **must** be named:

```
infracost-parser-plugin-<identifier>
```

For official plugins (in the `infracost/` namespace), the identifier is the short name:
- `infracost-parser-plugin-terraform` (canonical: `plugins.infracost.io/infracost/terraform`)

For community plugins, slashes in the namespace are replaced with double-dashes:
- `infracost-parser-plugin-acme--crossplane` (canonical: `plugins.infracost.io/acme/crossplane`)

Rules:
- On Windows, the `.exe` extension is added automatically
- Binaries ending in `-debug` are ignored (use this for debug builds)
- The CLI discovers plugins by scanning the plugin directory for files matching this pattern
- After discovery, the CLI calls `Describe` to get the canonical name — the binary filename is only for discovery

## go-plugin Handshake

Every plugin must use this exact handshake configuration:

```go
plugin.Serve(&plugin.ServeConfig{
    HandshakeConfig: plugin.HandshakeConfig{
        ProtocolVersion:  1,
        MagicCookieKey:   "INFRACOST_PARSER_PLUGIN_MAGIC_COOKIE",
        MagicCookieValue: "ac92b06c592f",
    },
    Plugins: map[string]plugin.Plugin{
        "parser": yourPluginStruct,
    },
    GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
        opts = append(opts,
            grpc.MaxRecvMsgSize(64 * 1024 * 1024),
            grpc.MaxSendMsgSize(64 * 1024 * 1024),
        )
        return grpc.NewServer(opts...)
    },
})
```

The plugin struct must implement `plugin.GRPCPlugin`:

```go
type myPlugin struct {
    plugin.NetRPCUnsupportedPlugin
}

func (p *myPlugin) GRPCServer(_ *plugin.GRPCBroker, g *grpc.Server) error {
    api.RegisterParserServiceServer(g, myServiceImplementation)
    return nil
}

func (p *myPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, _ *grpc.ClientConn) (interface{}, error) {
    return nil, fmt.Errorf("not implemented")
}
```

## gRPC Service Contract

Plugins implement the `ParserService` defined in `infracost/parser/api/service.proto`:

```protobuf
service ParserService {
    rpc Describe(DescribeRequest) returns (DescribeResponse);
    rpc Detect(DetectRequest) returns (DetectResponse);
    rpc Initialize(InitializeRequest) returns (InitializeResponse);
    rpc Parse(ParseRequest) returns (ParseResponse);
    rpc ParseToTree(ParseToTreeRequest) returns (ParseToTreeResponse);
}
```

### Describe

Returns static metadata about the plugin. Called once at startup when the CLI discovers the binary.

**Request:** Empty.

**Response:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Canonical plugin name in `registry/namespace/name` format, e.g. `"plugins.infracost.io/infracost/pulumi"`. Must be unique across all plugins. |
| `display_name` | string | yes | Human-readable name, e.g. `"Pulumi"`. Shown in CLI output. |
| `priority` | int32 | yes | Detection order. Lower = checked first. See [Priority](#priority). |
| `file_extensions` | string[] | yes | Extensions this plugin may handle, e.g. `[".yaml", ".json"]`. |
| `supports_directories` | bool | yes | Whether the plugin can detect/parse whole directories (not just single files). |

#### Priority

When multiple plugins claim overlapping file extensions (e.g., `.json` is used by CloudFormation, ARM, and Terraform JSON), priority determines which plugin's `Detect` is called first. The first plugin to return `detected: true` wins.

Guidelines for choosing a priority value:
- **1-19**: Formats with unambiguous extensions (e.g., `.tf` is always Terraform)
- **20-39**: Formats that require content sniffing but have strong signals
- **40+**: Formats with weak or generic signals (e.g., any `.yaml` file)

Existing plugins use: Terraform=10, ARM=25, CloudFormation=30.

### Detect

Determines whether this plugin can handle a given file or directory path. Called for each project path, in plugin priority order.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Absolute path to a file or directory. |
| `content` | bytes | Optional: pre-read file content (for LSP virtual documents). |
| `content_provided` | bool | `true` if the `content` field is populated. |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `detected` | bool | `true` if this plugin claims the path. |
| `project_type` | string | Identifier for the detected format, e.g. `"pulumi_python"`. Passed back in routing. |
| `confidence` | DetectConfidence | How strong the detection signal is. |

**Detection confidence levels:**
- `LOW` — Extension-only heuristic (e.g., "it's a .yaml file")
- `MEDIUM` — Content sniffing (e.g., "it contains apiVersion + kind fields")
- `HIGH` — Definitive match (e.g., "the `$schema` URL is an ARM deployment template")

**Contract:**
- Return `detected: false` quickly for paths you don't handle. Don't error on unknown paths.
- If `content_provided` is true, use the provided content instead of reading from disk.
- For directory detection, scan only the top-level entries (don't recurse deeply).
- Detection must be fast (< 100ms). Avoid network calls.

### Initialize

Called once before parsing begins. The CLI passes the set of resource types that the provider plugins can cost.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `terraform_supported_resources` | SupportedResources | Resource types the Terraform providers can cost. |
| `cloudformation_supported_resources` | SupportedResources | Resource types the CloudFormation providers can cost. |
| `kubernetes_supported_resources` | SupportedResources | Resource types the Kubernetes providers can cost. |
| `disable_graph_cache` | bool | Set by the LSP for fast re-parses. |

Most plugins can accept this call and return an empty response. Use the supported resources list if you want to mark resources as `supported: true/false` in your parse output.

**Response:** Empty.

### Parse

Parses the IaC files and returns format-specific resource data.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `repo_directory` | string | Absolute root of the repository. |
| `working_directory` | string | Absolute working directory of the project. |
| `target` | ParseRequestTarget | Format-specific target (oneof). |

The `target` is a protobuf `oneof` — your plugin receives the variant matching your format. For new formats, you'll need to define a new target message in the proto repo and add it to the oneof.

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `diagnostics` | Diagnostic[] | Warnings and errors encountered during parsing. |
| `result` | ParseResponseResult | Format-specific result (oneof). |
| `dependencies` | Dependency[] | Source-code dependencies (optional, for IDE features). |

The `result` is a protobuf `oneof` matching the format. For new formats, define a new result message.

**Contract:**
- Return partial results with diagnostics on recoverable errors (e.g., one file in a directory fails to parse).
- Return an error only for unrecoverable failures (e.g., the target is nil).
- Resources should have `supported: true` if the provider plugins can cost them.

### ParseToTree

Like Parse, but returns a provider-agnostic tree structure instead of a format-specific result. This is used by the provider plugins to generate cost estimates. Most plugins implement this by calling their Parse logic internally and converting to the tree format.

The tree structure (`tree.Tree`) is a hierarchy of services, resources, and cost components. See `infracost/tree/tree.proto` for the full definition.

## Adding a New Format

To add support for a completely new IaC format:

1. **Define proto messages** in the `infracost/proto` repo:
   - `proto/infracost/parser/<format>/target.proto` — your Target and Options messages
   - `proto/infracost/parser/<format>/result.proto` — your Result and Resource messages
   - Add your target to `ParseRequestTarget.oneof` in `service.proto`
   - Add your result to `ParseResponseResult.oneof` in `service.proto`
   - Run `make generate` to regenerate Go bindings

2. **Build the plugin binary**:
   - Create `cmd/infracost-parser-plugin-<format>/main.go`
   - Implement the `ParserService` gRPC server
   - Follow the handshake and naming conventions above

3. **Validate** using `infracost plugin validate ./your-binary` (see below)

4. **Register with the CLI** (optional, for official plugins):
   - Add to `ensurePerIaCPlugins` in `cli/pkg/plugins/config.go`
   - Add routing in `tryPerIaCPlugins` and `legacyRoute` in `cli/pkg/plugins/parser/parse.go`

For community plugins, step 4 is not needed — the CLI discovers any binary matching the naming pattern in the plugin directory.

## Validation

Use the CLI's built-in validation command to verify your plugin:

```bash
infracost plugin validate ./infracost-parser-plugin-myplugin
```

This runs a conformance test suite that checks:

1. **Connectivity** — The binary starts, handshakes, and responds to gRPC
2. **Describe** — Returns valid metadata (non-empty name, valid priority, extensions)
3. **Detect** — Returns `detected: false` for paths it doesn't handle (no crashes on unknown input)
4. **Detect** — Returns `detected: true` for provided test fixtures (if `--fixtures` flag is used)
5. **Initialize** — Accepts the call without error
6. **Parse** — Returns a valid ParseResponse for provided test fixtures

You can pass test fixtures to validate detection and parsing:

```bash
infracost plugin validate ./infracost-parser-plugin-myplugin \
    --fixture ./testdata/example-project
```

The validator reports pass/fail for each check with diagnostic output on failure.

## Constraints and Limits

- **Max gRPC message size**: 64 MB (both send and receive)
- **Plugin startup timeout**: 10 seconds
- **Detection time budget**: < 100ms per path
- **Binary size**: No hard limit, but aim for < 50 MB for fast downloads
- **Concurrency**: The CLI may spawn multiple plugin instances in parallel; your plugin should be safe to run concurrently (separate processes, not threads)

## Reference Implementations

See the existing per-IaC plugins in the `infracost/parser` repo for production examples:
- `cmd/infracost-parser-plugin-terraform/` — Directory-based, high-priority, unambiguous extensions
- `cmd/infracost-parser-plugin-cloudformation/` — File-based, content-sniffing, ambiguous extensions
- `cmd/infracost-parser-plugin-arm/` — File-based, schema-URL sniffing, JSON-only

For a minimal starting point, see the `example/` directory in this repo.
