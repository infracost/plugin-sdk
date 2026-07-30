# Template Alignment

## Overview
This repo is the **canonical home of the parser plugin template** (decided). The rebased branch already contains `parser/template/` (a snapshot of `../parser/plugin/template/`); it must be brought up to date, cleaned of copy-paste defects, and positioned so that other repos treat it as the source of truth rather than the other way around.

## Requirements
- `parser/template/` in this repo is the canonical starter: `main.go`, `options/`, `server/` with one file per RPC (`server.go`, `get_plugin_info.go`, `get_parser_config.go`, `identify_projects.go`, `parse.go`), per-RPC `_test.go`, and `testdata/` fixtures.
- Fix the copy-paste defects inherited from the upstream snapshot (verified present in this repo's copy):
  - `server/identify_projects.go` is verbatim CloudFormation sniffing logic (`identifyCloudFormationFile/JSON/YAML`) — replace with neutral placeholder identification that demonstrates the `directory` XOR `files` contract.
  - `server/get_parser_config_test.go` asserts `config.ProjectTypeCloudFormation`; `server/get_plugin_info_test.go` asserts CloudFormation identity; `server/parse_test.go` reads `cloudformation.options.json` — all must assert the template's own placeholder values.
  - Template tests must pass un-skipped for the template's own placeholder behavior (no `t.Skip` guards).
  - `options/options.go` and `server/parse.go` comments reference `raw_options_format`, which no longer exists in the proto — `raw_options` is always JSON.
- Document `IdentifyEnvironments` in the template (README section and/or a commented-out `identify_environments.go` example) as an optional RPC where `codes.Unimplemented` means "no environment support".
- The template's `go.mod` (if the template is made a buildable module) or the example's build setup must compile against the tagged proto release, consistent with the example specs.
- SDK docs give a single "start here" pointer: `example/` = minimal single-file walkthrough, `template/` = production-shaped starter to copy; each states when to use it.
- Downstream relationship: `../parser/plugin/template/` should be treated as a downstream copy of this repo's template. Actually updating ../parser (pointing its template here or syncing it) is a follow-up in that repo, out of scope here — but this repo's template README should state that this copy is canonical, and note the upstream ref it was last synced against so drift is detectable.

## Acceptance Criteria
- [ ] `parser/template/` contains no CloudFormation remnants (imports, sniffer logic, test assertions, fixture names).
- [ ] Template tests pass (`go test ./...`) without `t.Skip`.
- [ ] No `raw_options_format` references anywhere in the template.
- [ ] `IdentifyEnvironments` optionality is documented in the template.
- [ ] Template README states this repo is the canonical template source and records the `../parser` ref it was last reconciled with.
- [ ] Docs contain a single, unambiguous "start here" pointer distinguishing example vs template.

## Edge Cases
- Upstream divergence: if `../parser/plugin/template/` has changed on origin/main since the snapshot, reconcile those changes into this copy first, then declare canonical — don't declare canonical over a stale copy.
- ../providers has no template; the provider `example/` fills that role and its README should say so.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [parser-example-accuracy](parser-example-accuracy.md)
