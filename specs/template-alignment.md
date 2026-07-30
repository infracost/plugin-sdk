# Template Alignment

## Overview
This repo is the **canonical home of the parser plugin template** (decided). The rebased branch already contains `parser/template/` (a snapshot of `../parser/plugin/template/`); it must be brought up to date, cleaned of copy-paste defects, and positioned so that other repos treat it as the source of truth rather than the other way around.

## Requirements
- `parser/template/` in this repo is the canonical starter: `main.go`, `options/`, `server/` with one file per RPC (`server.go`, `get_plugin_info.go`, `get_parser_config.go`, `identify_projects.go`, `parse.go`), per-RPC `_test.go`, and `testdata/` fixtures.
- Make `parser/template/` a standalone Go module that builds from a fresh clone without sibling repositories. Its `go.mod` must pin the same tagged proto release as the examples (v1.160.0), contain no local `replace`, and include only importable public dependencies.
- Replace dependencies on `../parser` internal packages with self-contained template code: inline placeholder metadata/version values and use public generated proto types instead of parser-internal diagnostic, tree, harness, or version helpers.
- Fix the copy-paste defects inherited from the upstream snapshot (verified present in this repo's copy):
  - `server/identify_projects.go` is verbatim CloudFormation sniffing logic (`identifyCloudFormationFile/JSON/YAML`) — replace with neutral placeholder identification that demonstrates the `directory` XOR `files` contract.
  - `server/get_parser_config_test.go` asserts `config.ProjectTypeCloudFormation`; `server/get_plugin_info_test.go` asserts CloudFormation identity; `server/parse_test.go` reads `cloudformation.options.json` — all must assert the template's own placeholder values.
  - Template tests must pass un-skipped for the template's own placeholder behavior (no `t.Skip` guards).
  - `options/options.go` and `server/parse.go` comments reference `raw_options_format`, which no longer exists in the proto — `raw_options` is always JSON.
- Document `IdentifyEnvironments` in the template (README section and/or a commented-out `identify_environments.go` example) as an optional RPC where `codes.Unimplemented` means "no environment support".
- Run `go mod tidy`; the template must pass `go build ./...`, `go vet ./...`, and `go test ./...` in its own module.
- SDK docs give a single "start here" pointer: `example/` = minimal single-file walkthrough, `template/` = production-shaped starter to copy; each states when to use it.
- Downstream relationship: `../parser/plugin/template/` should be treated as a downstream copy of this repo's template. Actually updating ../parser (pointing its template here or syncing it) is a follow-up in that repo, out of scope here — but this repo's template README should state that this copy is canonical, and note the upstream ref it was last synced against so drift is detectable.

## Acceptance Criteria
- [ ] `parser/template/` contains no CloudFormation remnants (imports, sniffer logic, test assertions, fixture names).
- [ ] The standalone template passes `go build ./...`, `go vet ./...`, and `go test ./...` without sibling repositories, local replacements, or `t.Skip`.
- [ ] Template source imports no `internal/` package from another module.
- [ ] No `raw_options_format` references anywhere in the template.
- [ ] `IdentifyEnvironments` optionality is documented in the template.
- [ ] Template README states this repo is the canonical template source and records the `../parser` ref it was last reconciled with.
- [ ] Docs contain a single, unambiguous "start here" pointer distinguishing example vs template.

## Edge Cases
- Upstream divergence: if `../parser/plugin/template/` has changed on origin/main since the snapshot, reconcile those changes into this copy first, then declare canonical — don't declare canonical over a stale copy.
- The downstream parser repository may need to adapt public template imports to its own internal helpers when syncing; that follow-up must not make this canonical copy depend on those helpers.
- ../providers has no template; the provider `example/` fills that role and its README should say so.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [parser-example-accuracy](parser-example-accuracy.md)
