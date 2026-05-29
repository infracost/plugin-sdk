# Infracost Parser Plugin SDK

Build custom parser plugins to teach [Infracost](https://infracost.io) how to read new Infrastructure-as-Code formats.

## What is a parser plugin?

Infracost uses parser plugins to extract resources from IaC files (Terraform, CloudFormation, ARM, etc.). Each plugin is a standalone binary that communicates with the Infracost CLI over gRPC. The CLI discovers plugins by scanning a directory for binaries named `infracost-parser-plugin-<name>`.

When Infracost encounters a project, it asks each plugin (in priority order) whether it can handle the path. The first plugin to claim the path parses it and returns the extracted resources for cost estimation.

## Quick start

1. Copy the `example/` directory as your starting point
2. Rename the binary and update `Describe()` with your format's metadata
3. Implement `Detect()` to recognize your IaC files
4. Implement `Parse()` to extract resources
5. Validate with `infracost plugin validate ./your-binary`

### Build and validate the example

```bash
cd example
go build -o infracost-parser-plugin-example .
infracost plugin validate ./infracost-parser-plugin-example
```

### Test with a fixture

```bash
echo "hello" > test.example
infracost plugin validate ./infracost-parser-plugin-example --fixture test.example
```

## Interface contract

Your plugin must implement five gRPC RPCs:

| RPC | Purpose |
|-----|---------|
| **Describe** | Return plugin metadata (name, priority, file extensions) |
| **Detect** | Check if the plugin can handle a given path |
| **Initialize** | Accept supported resource types from the CLI |
| **Parse** | Parse IaC files and return resources |
| **ParseToTree** | Parse into a provider-agnostic tree (for cost estimation) |

See [SPEC.md](SPEC.md) for the full specification including message formats, detection contracts, and priority guidelines.

## Validation

The Infracost CLI includes a built-in conformance test suite:

```bash
infracost plugin validate ./infracost-parser-plugin-myformat
```

Checks:
- Binary starts and handshakes correctly
- `Describe` returns valid metadata
- `Detect` handles empty and nonexistent paths gracefully
- `Initialize` accepts the call without error
- With `--fixture`: `Detect` claims the fixture, `Parse` returns a response

## Adding a new IaC format

For a completely new format (not just a variant of an existing one), you'll also need to define proto messages for your target and result types. See the "Adding a New Format" section in [SPEC.md](SPEC.md).

## Reference

- [SPEC.md](SPEC.md) — Full plugin interface specification
- [example/](example/) — Minimal working plugin
- [infracost/parser](https://github.com/infracost/parser) — Production plugin implementations
