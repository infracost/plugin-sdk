# Parser Example Accuracy

## Overview
`parser/example/` must be a compiling, minimal parser plugin implementing the live `infracost.plugin` contract. The rebased base already implements `GetPluginInfo`/`GetParserConfig`/`IdentifyProjects`/`Parse` with the shared handshake; the remaining gaps are the go.mod pin (`v1.34.0` + broken `replace`), the missing `IdentifyEnvironments` mention, and test coverage.

## Requirements
- The example implements `PluginService.GetPluginInfo` (returning `type: PARSER`) and `ParserService.GetParserConfig`, `IdentifyProjects`, and `Parse`, registered together on one gRPC server behind the shared handshake.
- `Parse` returns a small but real `tree.Tree` (at least one provider → service → resource) so authors see the actual output shape instead of a format-specific oneof.
- `IdentifyEnvironments` is either implemented trivially or explicitly omitted with a comment stating that returning `codes.Unimplemented` is valid (mirroring the proto contract).
- `go.mod` builds against the tagged `github.com/infracost/proto` release ../parser pins (v1.160.0 — verified to publish the `infracost.plugin` package), with the `replace` directive removed entirely.
- The Makefile's targets must all succeed: `build`, a `test` target, and (replacing the dead `validate` target) an install-to-plugin-dir target or documented manual verification step.
- Example code style should be consistent with this repo's canonical `parser/template/` (see template-alignment spec) so authors graduating from example to template aren't relearning a layout; the example may stay single-file for readability, but names and idioms should match.

## Acceptance Criteria
- [ ] `go build ./...` succeeds in `parser/example/` from a fresh clone with no sibling checkouts.
- [ ] `go vet` and `go test ./...` pass; at least one test exercises each implemented RPC in-process.
- [ ] The example's handshake constants and dispense key match the CLI's (`INFRACOST_PLUGIN` / `de8c7e96-…` / `"plugin"`).
- [ ] No references to `Describe`, `Detect`, `DetectConfidence`, `Initialize`, `ParseToTree`, or cloudformation-result shims remain.
- [ ] The in-process gRPC test proves the binary's service contract is dispensable under key `"plugin"`, and the README documents how to verify it with `infracost plugin list`.
- [ ] If a compatible `infracost` binary is already installed, the example is smoke-tested through `plugin list`; absence of such a binary does not fail acceptance.

## Edge Cases
- Proto tag availability is resolved: v1.160.0 publishes `infracost.plugin` (it's what ../parser builds against). If a newer field the docs need is absent from the tag, flag it rather than reintroducing a replace directive.
- `IdentifyProjects` mutual-exclusion rule (`directory: true` forbids `files`) should be honored in the example logic since it's the most common contract mistake.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [handshake-documentation](handshake-documentation.md)
- [template-alignment](template-alignment.md)
