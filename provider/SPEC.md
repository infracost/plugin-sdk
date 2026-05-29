# Infracost Provider Plugin Specification

This document defines the interface contract for Infracost provider plugins. A provider plugin takes resources extracted by a parser plugin and returns cost estimates. The SDK makes no assumption about how prices are obtained — implementations may hardcode prices, query cloud APIs directly, use the Infracost Cloud API, or any other mechanism.

## Architecture Overview

```
infracost scan
    |
    v
Parser plugins extract resources from IaC files
    |
    v
Provider plugins price the extracted resources
    |
    v
CLI aggregates costs, evaluates policies, and renders output
```

Plugins are standalone binaries that communicate with the Infracost CLI over gRPC using the [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin) framework. The CLI spawns the plugin process, performs a handshake, and issues gRPC calls. When processing is complete the plugin process exits.

## Plugin Naming

Every plugin has a **canonical name** in `registry/namespace/name` format:

```
plugins.infracost.io/infracost/aws
plugins.infracost.io/acme/oracle
plugins.example.com/acme/custom
```

Resolution rules (short forms expand to the full canonical name):
- `aws` → `plugins.infracost.io/infracost/aws` (official plugin, default registry)
- `acme/oracle` → `plugins.infracost.io/acme/oracle` (community plugin, default registry)
- `plugins.example.com/acme/custom` → as-is (custom registry, first segment contains a dot)

The `infracost/` namespace is reserved for official plugins.

The canonical name is returned by the `Describe` RPC and is the source of truth for plugin identity.

## Binary Naming Convention

Plugin binaries **must** be named:

```
infracost-provider-plugin-<identifier>
```

For official plugins (in the `infracost/` namespace), the identifier is the short name:
- `infracost-provider-plugin-aws` (canonical: `plugins.infracost.io/infracost/aws`)

For community plugins, slashes in the namespace are replaced with double-dashes:
- `infracost-provider-plugin-acme--oracle` (canonical: `plugins.infracost.io/acme/oracle`)

Rules:
- On Windows, the `.exe` extension is added automatically
- Binaries ending in `-debug` are ignored (use this for debug builds)
- After discovery, the CLI calls `Describe` to get the canonical name — the binary filename is only for discovery

## go-plugin Handshake

Every plugin must use this exact handshake configuration:

```go
plugin.Serve(&plugin.ServeConfig{
    HandshakeConfig: plugin.HandshakeConfig{
        ProtocolVersion:  1,
        MagicCookieKey:   "INFRACOST_PROVIDER_PLUGIN_MAGIC_COOKIE",
        MagicCookieValue: "04d179d767fc",
    },
    Plugins: map[string]plugin.Plugin{
        "provider": yourPluginStruct,
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
    provider.RegisterProviderServiceServer(g, myServiceImplementation)
    return nil
}

func (p *myPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, _ *grpc.ClientConn) (interface{}, error) {
    return nil, fmt.Errorf("not implemented")
}
```

## gRPC Service Contract

Plugins implement the `ProviderService` defined in `infracost/provider/service.proto`:

```protobuf
service ProviderService {
    rpc Describe(DescribeRequest) returns (DescribeResponse);
    rpc Process(ProcessRequest) returns (ProcessResponse);
    rpc ProcessTree(ProcessTreeRequest) returns (ProcessTreeResponse);
    rpc ListFinopsPolicies(ListFinopsPoliciesRequest) returns (ListFinopsPoliciesResponse);
    rpc ListSupportedResources(ListSupportedResourcesRequest) returns (ListSupportedResourcesResponse);
}
```

### Describe

Returns static metadata about the plugin. Called once at startup when the CLI discovers the binary.

**Request:** Empty.

**Response:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Canonical plugin name in `registry/namespace/name` format, e.g. `"plugins.infracost.io/infracost/aws"`. |
| `display_name` | string | yes | Human-readable name, e.g. `"AWS"`. Shown in CLI output. |

### ListSupportedResources

Declares which resource types this provider can price. The CLI uses this to tell parser plugins which resources are "supported" (have cost data available).

**Request:** Empty.

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `terraform` | SupportedResources | Terraform resource types this provider can price (e.g. `aws_instance`, `aws_s3_bucket`). |
| `cloudformation` | SupportedResources | CloudFormation resource types this provider can price (e.g. `AWS::EC2::Instance`). |
| `kubernetes` | SupportedResources | Kubernetes resource types this provider can price. |

Each `SupportedResources` contains a list of `SupportedResource` messages with a `resource_type` string field.

**Contract:**
- Return at least one resource type across all categories.
- Resource type strings must match the parser's output exactly (case-sensitive).

### Process

Takes a flat list of parsed resources (from `Parse`) and returns cost estimates.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `input` | Input | Contains the parse result, usage data, project info, feature flags, and settings. |

The `Input` message includes:

| Field | Type | Description |
|-------|------|-------------|
| `parse_result` | ParseResponse | The output of a parser plugin. |
| `absolute_path` | string | Path to the project being scanned. |
| `project_info` | ProjectInfo | Project name, branch, workspace. |
| `usage` | Usage | Usage data for usage-based resources. |
| `features` | Features | Feature flags (enable price lookups, recommendations, policies, etc.). |
| `settings` | Settings | Currency code, disk cache preferences. |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `output` | Output | Contains priced resources and FinOps policy results. |

### ProcessTree

Like Process, but receives a provider-agnostic tree structure instead of a format-specific parse result. This is the modern path used by per-IaC parser plugins.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `input` | TreeInput | Contains the tree, usage data, project info, feature flags, and settings. |

The `TreeInput` message is identical to `Input` except `parse_result` is replaced by `tree` (a `tree.Tree`).

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `output` | Output | Contains priced resources and FinOps policy results. |

### ListFinopsPolicies

Returns the set of FinOps policies this provider can evaluate. Policies are rules like "use GP3 instead of GP2 for EBS volumes" or "avoid oversized instances."

**Request:** Empty.

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `policies` | FinopsPolicy[] | Available policies with slug, name, group, description, and applicability flags. |

**Contract:**
- Return an empty list if the provider does not evaluate any policies.
- Each policy must have a unique `slug`.

## Output Format

The `Output` message contains:

### Resource

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique resource ID from the parser. |
| `type` | string | Resource type (e.g. `aws_instance`, `AWS::EC2::Instance`). |
| `name` | string | Full resource address. |
| `region` | string | Cloud region (e.g. `us-east-1`). |
| `is_supported` | bool | Whether this provider can price the resource. |
| `is_free` | bool | Whether the resource has no associated costs. |
| `costs` | ResourceCosts | Cost components for this resource. |
| `child_resources` | Resource[] | Nested resources (e.g. a disk belonging to a VM). |

### CostComponent

Each resource's costs contain a list of `CostComponent` messages:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Display name (e.g. `"Compute (t3.small, on-demand)"`). |
| `unit` | string | Unit of measurement (e.g. `"hours"`, `"GB"`, `"requests"`). |
| `usage_based` | bool | Whether the price depends on usage data. |
| `price_not_found` | bool | True if no price was found (the provider should still return the component). |
| `price_was_hardcoded` | bool | True if the price was hardcoded rather than looked up dynamically. |
| `period_price` | PeriodPrice | Price per unit per period. |
| `quantity` | Rat | Number of units. |
| `discount_rate` | Rat | Discount rate (0.0 = no discount, 0.3 = 30% off). |

### PeriodPrice

| Field | Type | Description |
|-------|------|-------------|
| `price` | Rat | The price as a rational number (numerator/denominator as big-endian byte arrays). |
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

The `Input.settings` message includes a currency code so the provider can return prices in the user's preferred currency. The `Input.infracost` message contains Infracost-specific credentials that only apply to providers using the Infracost Cloud API — community providers should ignore these fields.

## Validation

Use the CLI's built-in validation command to verify your plugin:

```bash
infracost plugin validate ./infracost-provider-plugin-myprovider
```

This runs a conformance test suite that checks:

1. **Connectivity** — The binary starts, handshakes, and responds to gRPC
2. **ListSupportedResources** — Returns at least one supported resource type
3. **Process** — Accepts an empty request without crashing
4. **ListFinopsPolicies** — Accepts the call without error

## Constraints and Limits

- **Max gRPC message size**: 64 MB (both send and receive)
- **Plugin startup timeout**: 10 seconds
- **Binary size**: No hard limit, but aim for < 50 MB for fast downloads
- **Concurrency**: The CLI may spawn multiple plugin instances in parallel; your plugin should be safe to run concurrently (separate processes, not threads)

## Reference Implementations

See the official Infracost provider plugin in the `infracost/providers` repo:
- `cmd/infracost-provider-plugin/` — Production implementation supporting AWS, Azure, and GCP

For a minimal starting point, see the `example/` directory in this repo.
