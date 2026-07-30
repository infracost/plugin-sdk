# Implementation Baseline

## Overview
Defines what "the actual implementation" means for all documentation-accuracy work in this repo. The SDK docs must be verified against a single, explicit baseline, because the sibling repos contain multiple generations of the plugin contract and several unmerged branches that describe abandoned designs.

## Requirements
- **Repository scope (decided):** review and correct both plugin types. In-scope surfaces are the root README; parser README, specification, example, and template; provider README, specification, and example; and shared discovery, naming, handshake, command, and testing guidance.
- **Doc base (decided):** this branch is rebased onto `update-plugin-sdk-refactor`, whose rewritten docs/examples/template are the working baseline. The remaining work is drift correction on top of it, not a rewrite: removing `raw_options_format` (dropped from the proto; `raw_options` is always JSON), documenting `IdentifyEnvironments`, dropping the nonexistent `TreeInput.raw_options` provider-options channel unless the tagged proto has it, refreshing go.mod pins, and template cleanup.
- **Sibling-repo authority (decided):** `origin/main` (or `master` where that is the default) of each sibling repo — ../cli, ../proto, ../parser, ../providers, ../config — is assumed to be the correct, working implementation. Do not build sibling repos to verify behavior; verify doc claims by reading their code. Local checkouts may sit on feature branches (e.g. ../proto on `feature/arm-plugin-architecture`); claims must be checked against `origin/main`, not the local branch state.
- The wire-contract reference is the tagged `github.com/infracost/proto` release that ../parser pins (v1.160.0), whose `gen/go/infracost/plugin` package is what examples build against.
- The authoritative contract is the `infracost.plugin` proto package (`PluginService.GetPluginInfo`, `ParserService.{GetParserConfig, IdentifyProjects, IdentifyEnvironments (optional), Parse}`, `ProviderService.{Process, ListFinopsPolicies}`).
- The legacy `infracost.parser.api.ParserService` (`Initialize`/`Parse`/`ParseToTree`) and legacy `infracost.provider.ProviderService` (`Process`/`ProcessTree`/`ListFinopsPolicies`) still exist in the proto repo but are NOT served by any plugin or dialed by the CLI; docs must not present them as the plugin contract.
- The `Describe`/`Detect`/`ListSupportedResources` RPCs, `DetectConfidence` enum, registry-host naming (`plugins.infracost.io/...`), per-type magic cookies, and priority values 10/25/30 exist only on unmerged feature branches. None of these may appear in the docs as current behavior.
- **Planned CLI tooling (decided):** `infracost plugin validate` and `infracost plugin add` are wanted and now tracked in the CLI project as issues `cli-b10` and `cli-6f8`. Until they ship, SDK docs must not present them as existing; a clearly-labeled future-work note referencing the issues is acceptable.

## Acceptance Criteria
- [ ] Parser and provider documentation, examples, and shared root guidance are all included in the implementation review.
- [ ] A written baseline note (in the implementation plan or a doc appendix) records which repo/branch/version each doc claim was verified against, with sibling repos checked at `origin/main`/`master`.
- [ ] No SDK doc references `Describe`, `Detect`, `Initialize`, `ParseToTree`, `ProcessTree`, `ListSupportedResources`, `DetectConfidence`, `plugins.infracost.io`, `INFRACOST_PARSER_PLUGIN_MAGIC_COOKIE`, or `INFRACOST_PROVIDER_PLUGIN_MAGIC_COOKIE` as current behavior.
- [ ] Docs describe only RPCs, messages, and fields present in the tagged `github.com/infracost/proto` v1.160.0 release, verified against that module's `gen/go` output rather than the local ../proto working tree.
- [ ] `plugin validate`/`plugin add` appear only as future work referencing `cli-b10`/`cli-6f8`, if at all.

## Edge Cases
- The legacy `infracost.parser.api` package is still imported by ../parser for dependency extraction only; if docs mention it at all, they must scope it to that use.
- Proto messages that exist but are unused by the CLI (e.g. `identification_priority` is defined but never read by ../cli in production; identification is driven via the ../config library) — docs must not overstate runtime behavior.
- Where the local ../proto feature branch and the tagged release disagree (e.g. a field present locally but not in v1.160.0, or vice versa), the tagged release wins for anything the examples compile against; flag the difference rather than silently picking one.

## Dependencies
- All other specs in this directory inherit this baseline.
