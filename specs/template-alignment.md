# Template Alignment

## Overview
The `../parser` repo ships a starter template at `plugin/template/` (main.go + options package + server package with one file per RPC, table-driven tests, testdata fixtures). The SDK's parser materials must align with that latest template so authors copying from either place end up with the same, current structure.

## Requirements
- The SDK's parser starting point mirrors `../parser/plugin/template/` in layout and idiom: `main.go`, `options/`, `server/` with `get_plugin_info.go`, `get_parser_config.go`, `identify_projects.go`, `parse.go`, per-file `_test.go`, and `testdata/` fixtures.
- Decide (and record) the relationship between the SDK's `example/` and a possible SDK `template/` copy: the unmerged SDK refactor branch added a `parser/template/` snapshot of the ../parser template — either carry that forward refreshed, or link to the ../parser template as canonical and keep only the minimal `example/`. Avoid shipping two divergent starters without a stated purpose.
- Any template copy in the SDK must be refreshed from ../parser HEAD and cleaned of the known copy-paste defects in the upstream template, or those defects fixed upstream first: `parse.go` importing `plugin/cloudformation/options` instead of its own options package, `identify_projects.go` being verbatim CloudFormation sniffing logic, and skipped tests asserting CloudFormation values (`infracost/cloudformation` name, CloudFormation project type).
- The template/README must document `IdentifyEnvironments` as an optional RPC (the upstream template omits it; only terraform/terragrunt/kubernetes implement it) so authors know it exists.
- SDK docs must point to the template location(s) that actually exist — `plugin/template/` in ../parser — not `cmd/...` paths.
- Note for scope: ../providers has no template; the provider example (see provider-example-accuracy) fills that role and should say so.

## Acceptance Criteria
- [ ] SDK parser starter layout matches `../parser/plugin/template/` file-for-file in structure (allowing SDK-specific module paths).
- [ ] No CloudFormation copy-paste remnants (imports, test assertions, sniffer logic presented as neutral) in whichever starter the SDK ships.
- [ ] Docs contain a single, unambiguous "start here" pointer; if both example and template exist, each states when to use it.
- [ ] `IdentifyEnvironments` optionality is mentioned in the starter's README or code comments.

## Edge Cases
- Upstream template drift: if `../parser/plugin/template/` changes after the SDK copy is refreshed, the SDK docs should carry a "sourced from parser@<ref>" note so future drift is detectable.
- The upstream template's tests run only under `t.Skip`; an SDK copy should ship passing (non-skipped) tests for its own placeholder behavior.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [parser-example-accuracy](parser-example-accuracy.md)
