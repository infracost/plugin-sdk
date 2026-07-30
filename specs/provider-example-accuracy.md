# Provider Example Accuracy

## Overview
`provider/example/` must be a compiling, minimal provider plugin implementing the live `infracost.plugin` contract. The rebased base already implements `GetPluginInfo`/`Process`/`ListFinopsPolicies` with the shared handshake; the remaining gaps are the go.mod pin (`v1.34.0` + broken `replace`) and test coverage.

## Requirements
- The example implements `PluginService.GetPluginInfo` (returning `type: PROVIDER`) and `ProviderService.Process` + `ListFinopsPolicies`, registered together on one gRPC server behind the shared handshake.
- `Process` walks `input.tree` (providers → services → resources) and returns a hardcoded-price `Output` for a recognizable resource type, demonstrating: `ResourceCosts.components`, `PeriodPrice` with `Rat` math, `quantity`, and `price_was_hardcoded`.
- `ListFinopsPolicies` returns an empty list with a comment showing the five `FinopsPolicy` fields for authors who want policies.
- No `ListSupportedResources`, no `SupportedResources` imports from `infracost/parser/api`, no `ProcessTree`.
- `go.mod` builds against the tagged `github.com/infracost/proto` v1.160.0 release with the `replace` directive removed (same resolution as the parser example spec).
- Makefile targets all succeed; the dead `validate` target is replaced per the CLI-commands spec.
- Since ../providers has no template/skeleton, this example is the de-facto template for community providers — it should include at least one unit test demonstrating how to build a `TreeInput` and assert on `Output`, mirroring the test style used in ../providers.

## Acceptance Criteria
- [ ] `go build ./...` succeeds in `provider/example/` from a fresh clone with no sibling checkouts.
- [ ] `go test ./...` passes with at least one in-process test per implemented RPC.
- [ ] Handshake constants and dispense key match the CLI's.
- [ ] No references to `Describe`, `ListSupportedResources`, `ProcessTree`, or `plugins.infracost.io` remain.
- [ ] A binary built from the example is loadable by the CLI (appears in `infracost plugin list`), or the blocker is documented.

## Edge Cases
- The example receives the full tree including other clouds' resources (the CLI sends the tree to every provider); it should show returning only what it prices and ignoring the rest, like the official plugins' prune-by-provider-key behavior.
- `Rat` encoding of negative values and zero denominators — the helper should guard or comment on them since authors will copy it.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [handshake-documentation](handshake-documentation.md)
- [provider-spec-accuracy](provider-spec-accuracy.md)
