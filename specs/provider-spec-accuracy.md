# Provider Spec Accuracy

## Overview
`provider/SPEC.md` must document the live provider plugin contract: `infracost.plugin.ProviderService` (`Process`, `ListFinopsPolicies`) plus the mandatory `PluginService`, with the real input/output message shapes from `../proto/proto/infracost/provider/{tree,input,output}.proto`.

## Requirements
- Document exactly two `ProviderService` RPCs — `Process` and `ListFinopsPolicies` — from `../proto/proto/infracost/plugin/provider.proto`. `Describe`, `ProcessTree`, and `ListSupportedResources` do not exist and must be removed.
- `Process` takes `provider.TreeInput` (the tree-shaped input; there is no flat parse-result path). Document `TreeInput`'s actual fields: `tree`, `absolute_path`, `project_info`, `previous_resource_addresses`, `usage`, `finops_policy_config`, `features`, `settings`, `infracost` (credentials — community providers ignore). Do **not** document `raw_options`/`raw_options_format` on `TreeInput` — no provider-options channel exists in the current proto (the unmerged SDK refactor branch documents one; that content must be dropped or explicitly re-verified against a proto release that adds it).
- Document `Output` accurately: `resources[]` and `finops_results[]`. The `Resource` message has 14 fields — including `metadata`, `provider_link`, `is_provider_supported`, `action`, `tagging`, `call_stack` — and `costs` is a `ResourceCosts` wrapper containing `components[]`, not a bare list.
- `CostComponent` includes `environmental_metrics` in addition to the eight commonly known fields. `PeriodPrice.period` enum includes `PERIOD_UNSPECIFIED`. `Rat` stays as documented (`infracost.rational`).
- `FinopsPolicy` has exactly five fields: `slug`, `name`, `group`, `description`, `only_new_resources` (singular flag — not "applicability flags").
- Document `GetPluginInfo` for providers (`type: PROVIDER`; name convention `infracost/<cloud>`).
- Reference implementations: `../providers` has one binary **per cloud** — `plugin/{aws,azure,google,kubernetes}/main.go` — built on a shared internal server that prunes the incoming tree to its own provider key. There is no `cmd/infracost-provider-plugin/` and no combined AWS+Azure+GCP binary; Kubernetes is a fourth provider the docs currently omit.
- Since `ListSupportedResources` doesn't exist, remove the claim that the CLI feeds provider-supported resource types back into parsers; nothing populates `SupportedResources` anywhere.
- Keep and verify the pricing-agnosticism section (hardcode / cloud APIs / Infracost Cloud API) — it is a design statement, but its references to message fields (`settings.currency`, `infracost` credentials) must match the proto.

## Acceptance Criteria
- [ ] RPC list in `provider/SPEC.md` is exactly `GetPluginInfo` + `Process` + `ListFinopsPolicies`.
- [ ] `TreeInput`, `Output`, `Resource`, `ResourceCosts`, `CostComponent`, `PeriodPrice`, `Rat`, `FinopsPolicy` field tables match the proto files field-for-field (including the fields previously omitted).
- [ ] No `raw_options` documented on `TreeInput` unless it exists in the tagged proto the example builds against.
- [ ] Reference-implementation paths resolve in ../providers HEAD and name all four cloud plugins.
- [ ] No mention of supported-resources feedback between providers and parsers.

## Edge Cases
- `ProcessResponse.output` is field 2 in the legacy proto but field 1 in messages actually used — docs shouldn't cite field numbers at all unless verified.
- Multiple provider plugins all receive the tree (the CLI runs every loaded provider); tree pruning by provider key happens plugin-side. Docs advising "return empty Output when nothing to price" should reflect this.
- `FinopsPolicyResult` (output) has its own applicability booleans that are easy to confuse with `FinopsPolicy.only_new_resources`; keep the two clearly separated.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [handshake-documentation](handshake-documentation.md)
- [discovery-and-naming-documentation](discovery-and-naming-documentation.md)
