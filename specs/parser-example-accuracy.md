# Parser Example Accuracy

## Overview
`parser/example/` must be a compiling, minimal parser plugin implementing the live `infracost.plugin` contract. Today it implements the abandoned `Describe`/`Detect` API and fails to build against the current proto module.

## Requirements
- The example implements `PluginService.GetPluginInfo` (returning `type: PARSER`) and `ParserService.GetParserConfig`, `IdentifyProjects`, and `Parse`, registered together on one gRPC server behind the shared handshake.
- `Parse` returns a small but real `tree.Tree` (at least one provider → service → resource) so authors see the actual output shape instead of a format-specific oneof.
- `IdentifyEnvironments` is either implemented trivially or explicitly omitted with a comment stating that returning `codes.Unimplemented` is valid (mirroring the proto contract).
- `go.mod` builds against a **tagged** `github.com/infracost/proto` release that contains the `infracost.plugin` package (the version ../parser pins, v1.159.0+), with no `replace` directive; if a replace is temporarily unavoidable it must carry an accurate comment and a path that works from a fresh clone (the current `../../../proto` breaks in any non-sibling checkout).
- The Makefile's targets must all succeed: `build`, a `test` target, and (replacing the dead `validate` target) an install-to-plugin-dir target or documented manual verification step.
- Example code style should mirror the structure of `../parser/plugin/template/` (main.go + server package, one file per RPC, table-driven tests) so authors graduating to the template aren't relearning a layout — see the template-alignment spec.

## Acceptance Criteria
- [ ] `go build ./...` succeeds in `parser/example/` from a fresh clone with no sibling checkouts.
- [ ] `go vet` and `go test ./...` pass; at least one test exercises each implemented RPC in-process.
- [ ] The example's handshake constants and dispense key match the CLI's (`INFRACOST_PLUGIN` / `de8c7e96-…` / `"plugin"`).
- [ ] No references to `Describe`, `Detect`, `DetectConfidence`, `Initialize`, `ParseToTree`, or cloudformation-result shims remain.
- [ ] A binary built from the example is loadable by the CLI (appears in `infracost plugin list` when dropped in the plugin dir), or the reason it cannot be (if any) is documented.

## Edge Cases
- Proto tag availability: if no tagged release exposes `infracost.plugin` publicly, the requirement degrades to a replace directive with a working relative path plus a README note — this must be an explicit, recorded decision, not an accident.
- `IdentifyProjects` mutual-exclusion rule (`directory: true` forbids `files`) should be honored in the example logic since it's the most common contract mistake.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [handshake-documentation](handshake-documentation.md)
- [template-alignment](template-alignment.md)
