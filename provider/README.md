# Infracost Provider Plugin SDK

Build custom provider plugins to teach [Infracost](https://infracost.io) how to price resources from any cloud or service.

## What is a provider plugin?

Infracost uses provider plugins to turn parsed IaC resources into cost estimates. Each plugin is a standalone binary that communicates with the Infracost CLI over gRPC. A provider is completely free to decide how it obtains prices — hardcode them, query a cloud pricing API, use a local database, or call the Infracost Cloud API.

## Quick start

1. Copy the `example/` directory as your starting point
2. Update `ListSupportedResources()` with the resource types you handle
3. Implement `Process()` to return cost components for each resource
4. Validate with `infracost plugin validate ./your-binary`

### Build and validate the example

```bash
cd example
go build -o infracost-provider-plugin-example .
infracost plugin validate ./infracost-provider-plugin-example
```

## Interface contract

Your plugin must implement five gRPC RPCs:

| RPC | Purpose |
|-----|---------|
| **Describe** | Return plugin metadata (canonical name, display name) |
| **ListSupportedResources** | Declare which resource types this provider can price |
| **Process** | Price a flat list of parsed resources |
| **ProcessTree** | Price a provider-agnostic resource tree |
| **ListFinopsPolicies** | Return available FinOps policy definitions |

See [SPEC.md](SPEC.md) for the full specification including message formats and the CostComponent contract.

## Validation

The Infracost CLI includes a built-in conformance test suite:

```bash
infracost plugin validate ./infracost-provider-plugin-myprovider
```

Checks:
- Binary starts and handshakes correctly
- `Describe` returns a valid canonical name and display name
- `ListSupportedResources` returns at least one resource type
- `Process` accepts an empty request without crashing
- `ListFinopsPolicies` accepts the call without error

## Reference

- [SPEC.md](SPEC.md) — Full plugin interface specification
- [example/](example/) — Minimal working plugin with hardcoded pricing
- [infracost/providers](https://github.com/infracost/providers) — Production provider implementation
