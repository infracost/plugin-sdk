# Infracost Provider Plugin Specification

This document defines the interface contract for Infracost provider plugins. A provider plugin takes the IaC-agnostic cost tree extracted by a parser plugin and returns cost estimates and FinOps policy results. The SDK makes no assumption about how prices are obtained — implementations may hardcode prices, query cloud APIs directly, use the Infracost Cloud API, or any other mechanism.

## Architecture Overview

```
infracost scan
    |
    v
Parser plugins extract resources into an IaC-agnostic cost tree
    |
    v
Provider plugins price the tree and evaluate FinOps policies
    |
    v
CLI aggregates costs, evaluates output, and renders results
```

Plugins are standalone binaries that communicate with the Infracost CLI over gRPC using the [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin) framework. The CLI spawns the plugin process, performs a handshake, and issues gRPC calls. When processing is complete the plugin process exits.

Every plugin implements two gRPC services:

- **`PluginService`** — a single `GetPluginInfo` RPC that reports the plugin's type and metadata. This is how the CLI tells provider plugins apart from parser plugins.
- **`ProviderService`** — the pricing RPCs (`Process`, `ListFinopsPolicies`).

Both services are registered on the same gRPC server (see [Handshake](#go-plugin-handshake)).

## Plugin Identity and Naming

The CLI **does not** infer a plugin's type or identity from its binary filename. It launches every executable in the plugin directory, calls `GetPluginInfo`, and uses the returned `type` to decide whether the binary is a parser or a provider. Binaries that fail to launch or handshake are skipped.

A descriptive binary name such as `infracost-provider-plugin-<name>` is conventional and recommended for clarity, but it is not required for discovery.

The `name` returned by `GetPluginInfo` is the plugin's identity. By convention it is `<namespace>/<name>`:

- Official plugins use the `infracost/` namespace, e.g. `infracost/aws`, `infracost/azure`, `infracost/google`.
- Community plugins should use their own namespace, e.g. `acme/oracle`.

Names must be unique across all installed plugins.

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
    pluginpb.RegisterProviderServiceServer(g, myServiceImplementation)
    return nil
}

func (p *myPlugin) GRPCClient(_ context.Context, _ *goplugin.GRPCBroker, _ *grpc.ClientConn) (interface{}, error) {
    return nil, fmt.Errorf("not implemented")
}
```

The generated Go bindings live in `github.com/infracost/proto/gen/go/infracost/plugin` (aliased `pluginpb` above). Embedding `pluginpb.UnimplementedPluginServiceServer` and `pluginpb.UnimplementedProviderServiceServer` in your service struct keeps it forward-compatible if new RPCs are added.

## gRPC Service Contract

Plugins implement `PluginService` and `ProviderService`, both defined in the `infracost.plugin` package:

```protobuf
// infracost/plugin/plugin.proto
service PluginService {
    rpc GetPluginInfo(GetPluginInfoRequest) returns (GetPluginInfoResponse);
}

// infracost/plugin/provider.proto
service ProviderService {
    rpc Process(ProcessRequest) returns (ProcessResponse);
    rpc ListFinopsPolicies(ListFinopsPoliciesRequest) returns (ListFinopsPoliciesResponse);
}
```

### GetPluginInfo

Returns the plugin's type and static metadata. Called once at startup when the CLI discovers the binary.

**Request:** Empty.

**Response:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | PluginType | yes | `PROVIDER` for provider plugins. (`PARSER` and `PLUGIN_TYPE_UNKNOWN` are the other values.) |
| `name` | string | yes | Plugin identity, e.g. `"infracost/aws"` or `"acme/oracle"`. Must be unique. |
| `version` | string | no | Plugin version. Semver recommended but not required. |
| `description` | string | no | Human-readable description shown in CLI output. |
| `url` | string | no | Where the plugin is documented/downloaded. |
| `author` | string | no | Company or individual that authored the plugin. |

### Process

Receives the IaC-agnostic cost tree and returns priced resources and FinOps policy results.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `input` | TreeInput | The cost tree plus usage data, project info, feature flags, and settings. |

The `TreeInput` message (`infracost/provider/tree.proto`) includes:

| Field | Type | Description |
|-------|------|-------------|
| `tree` | tree.Tree | The parser output as an IaC-agnostic tree (providers → services → resources). |
| `absolute_path` | string | Path to the project being scanned (a directory for Terraform/Terragrunt, a file for CloudFormation). |
| `project_info` | ProjectInfo | Project name, branch, workspace, production flag. |
| `previous_resource_addresses` | string[] | Resource addresses seen on a previous run, used to decide which resources are "new". |
| `usage` | Usage | Usage data for usage-based resources. |
| `finops_policy_config` | FinopsPolicyConfiguration | Which policies to run and their settings (only relevant if `features.enable_finops_policies`). |
| `features` | Features | Feature flags (price lookups, recommendations, policies, environmental metrics). |
| `settings` | Settings | Currency code, disk cache preferences. |
| `infracost` | Infracost | Infracost-specific credentials (API key, pricing endpoint, trace/org IDs). Community providers should ignore these. |

Unlike parser plugins, `TreeInput` has no generic options channel — a provider
plugin that needs its own configuration must use an out-of-band mechanism
(env vars, files, or flags).

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `output` | Output | Priced resources and FinOps policy results. See [Output Format](#output-format). |

**Contract:**
- Walk `input.tree` (providers → services → resources), price the resources you support, and return them in `output.resources`.
- Return an empty `Output` rather than an error when there is nothing to price.
- Honour `input.settings.currency`; convert prices to the requested currency where you can.

### ListFinopsPolicies

Returns the set of FinOps policies this provider can evaluate. Policies are rules like "use GP3 instead of GP2 for EBS volumes" or "avoid oversized instances".

**Request:** Empty.

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `policies` | FinopsPolicy[] | Available policies. |

Each `FinopsPolicy` has:

| Field | Type | Description |
|-------|------|-------------|
| `slug` | string | Unique policy identifier. |
| `name` | string | Human-readable name. |
| `group` | string | Grouping/category. |
| `description` | string | Human-readable description. |
| `only_new_resources` | bool | Whether the policy applies only to newly added resources. |

**Contract:**
- Return an empty list if the provider does not evaluate any policies.
- Each policy must have a unique `slug`.

## Output Format

The `Output` message (`infracost/provider/output.proto`) contains:

| Field | Type | Description |
|-------|------|-------------|
| `resources` | Resource[] | Normalised, priced, IaC-agnostic resources. |
| `finops_results` | FinopsPolicyResult[] | Per-policy pass/fail results. |

### Resource

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique resource ID from the parser. |
| `type` | string | Resource type (e.g. `aws_instance`, `AWS::EC2::Instance`). |
| `name` | string | Full resource address. |
| `region` | string | Cloud region (e.g. `us-east-1`). |
| `metadata` | ResourceMetadata | Filename, line numbers, checksums. |
| `is_supported` | bool | Whether this provider can price the resource. |
| `is_free` | bool | Whether the resource has no associated costs. |
| `is_provider_supported` | bool | Whether the resource's provider is supported at all. |
| `action` | ResourceAction | Whether the resource was added/modified/deleted/unchanged this run. |
| `costs` | ResourceCosts | Cost components for this resource. |
| `tagging` | Tagging | Resource-level tag information. |
| `child_resources` | Resource[] | Nested resources (e.g. a disk belonging to a VM). |

### CostComponent

Each resource's `costs.components` is a list of `CostComponent` messages:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Display name (e.g. `"Compute (t3.small, on-demand)"`). |
| `unit` | string | Unit of measurement (e.g. `"hours"`, `"GB"`, `"requests"`). |
| `usage_based` | bool | Whether the price depends on usage data. |
| `price_not_found` | bool | True if no price was found (still return the component). |
| `price_was_hardcoded` | bool | True if the price was hardcoded rather than looked up dynamically. |
| `period_price` | PeriodPrice | Price per unit per period. |
| `quantity` | Rat | Number of units. |
| `discount_rate` | Rat | Discount rate (0.0 = none, 0.3 = 30% off). |
| `environmental_metrics` | EnvironmentalMetrics | Optional carbon/water metrics. |

### PeriodPrice

| Field | Type | Description |
|-------|------|-------------|
| `price` | Rat | The price as a rational number. |
| `period` | Period | `MONTH` or `HOUR`. |

### Rat (Rational Number)

Prices and quantities use `infracost.rational.Rat` to avoid floating-point precision issues:

| Field | Type | Description |
|-------|------|-------------|
| `numerator` | bytes | Big-endian unsigned integer bytes. |
| `denominator` | bytes | Big-endian unsigned integer bytes. |
| `negative` | bool | Whether the value is negative. |

For example, `$0.0116/hr` can be represented as `numerator=116, denominator=10000`.

## Pricing Agnosticism

The provider plugin interface is deliberately agnostic about how prices are obtained. A provider may:

- **Hardcode prices** — Simplest approach, suitable for stable or internal pricing
- **Query cloud APIs** — Call AWS Pricing API, Azure Retail Prices API, etc. directly
- **Use a pricing database** — Query a local or remote database of prices
- **Use the Infracost Cloud API** — The official Infracost providers use this approach
- **Combine approaches** — e.g., hardcode some prices and look up others dynamically

`TreeInput.settings` includes a currency code so the provider can return prices in the user's preferred currency. `TreeInput.infracost` contains Infracost-specific credentials that only apply to providers using the Infracost Cloud API — community providers should ignore these fields.

## Installing and testing

The CLI discovers plugins by scanning a plugin directory, which defaults to `os.UserCacheDir()/infracost/plugins`:

- Linux: `~/.cache/infracost/plugins`
- macOS: `~/Library/Caches/infracost/plugins`
- Windows: `%LocalAppData%\infracost\plugins`

Drop your built binary in that directory and run `infracost` against a project. The CLI will launch the binary, call `GetPluginInfo`, and route the cost tree to it if it reports `type: PROVIDER`.

Because the gRPC contract is plain Go, the most reliable way to test a provider is with Go unit tests that build a `TreeInput` and call `Process` / `ListFinopsPolicies` directly.

## Constraints and Limits

- **Max gRPC message size**: 64 MB (both send and receive)
- **Binary size**: no hard limit, but aim for fast downloads
- **Concurrency**: the CLI may run plugins in parallel; plugins are separate processes, so keep any shared on-disk state (caches) concurrency-safe

## Reference Implementations

For a minimal starting point, see the [`example/`](example) directory in this repo.

The official Infracost provider plugins (AWS, Azure, Google, Kubernetes) follow this same contract as separate per-cloud binaries, each scoped to one cloud's pricing and policies.
