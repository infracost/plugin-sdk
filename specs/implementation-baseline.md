# Implementation Baseline

## Overview
Defines what "the actual implementation" means for all documentation-accuracy work in this repo. The SDK docs must be verified against a single, explicit baseline, because the sibling repos contain multiple generations of the plugin contract and several unmerged branches that describe abandoned designs.

## Requirements
- The authoritative contract is the `infracost.plugin` proto package (`../proto/proto/infracost/plugin/{plugin,parser,provider}.proto`) as consumed by the CLI (`../cli/pkg/plugins/`), the config library (`../config/plugin/`), the parser plugins (`../parser/plugin/`), and the provider plugins (`../providers/plugin/`).
- The legacy `infracost.parser.api.ParserService` (`Initialize`/`Parse`/`ParseToTree`) and legacy `infracost.provider.ProviderService` (`Process`/`ProcessTree`/`ListFinopsPolicies`) still exist in the proto repo but are NOT served by any plugin or dialed by the CLI; docs must not present them as the plugin contract.
- The `Describe`/`Detect`/`ListSupportedResources` RPCs, `DetectConfidence` enum, registry-host naming (`plugins.infracost.io/...`), per-type magic cookies, and priority values 10/25/30 exist only on unmerged feature branches (`feature/kubernetes-plugin-architecture` in ../proto, dead commits in ../parser). None of these may appear in the docs as current behavior.
- The unmerged SDK branch `update-plugin-sdk-refactor` already rewrote the docs, examples, and a parser template against the `infracost.plugin` contract. Its content should be treated as the starting point and brought up to date, not discarded — but every claim in it must be re-verified because it predates newer proto changes (e.g. `IdentifyEnvironments`, `raw_options` becoming always-JSON with `raw_options_format` removed).
- Where sibling repos are checked out on feature branches (e.g. ../proto on `feature/arm-plugin-architecture`), claims must hold on the repo's main branch too, or be flagged; the tagged proto release consumed by ../cli and ../parser (v1.159.0+) is the wire-contract reference.

## Acceptance Criteria
- [ ] A written baseline note (in the implementation plan or a doc appendix) records which repo/branch/version each doc claim was verified against.
- [ ] No SDK doc references `Describe`, `Detect`, `Initialize`, `ParseToTree`, `ProcessTree`, `ListSupportedResources`, `DetectConfidence`, `plugins.infracost.io`, `INFRACOST_PARSER_PLUGIN_MAGIC_COOKIE`, or `INFRACOST_PROVIDER_PLUGIN_MAGIC_COOKIE` as current behavior.
- [ ] Docs describe only RPCs, messages, and fields present in the tagged `github.com/infracost/proto` release the examples build against.

## Edge Cases
- The legacy `infracost.parser.api` package is still imported by ../parser for dependency extraction only; if docs mention it at all, they must scope it to that use.
- Proto messages that exist but are unused by the CLI (e.g. `identification_priority` is defined but never read by ../cli in production; identification is driven via the ../config library) — docs must not overstate runtime behavior.
- If a doc claim is intentionally forward-looking (a planned feature), it must be clearly labeled as not yet implemented, or removed.

## Dependencies
- All other specs in this directory inherit this baseline.
