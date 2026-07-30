# Implementation Plan — Ensure parser docs match the actual proto contract

## Context
The parser SDK docs and example describe an invented 5-RPC interface
(`Describe, Detect, Initialize, Parse, ParseToTree` with `DetectConfidence`,
`file_extensions`, Kubernetes supported resources, format-specific oneofs) that
matches **neither** proto in `github.com/infracost/proto`. The example does not
compile (`undefined: api.DescribeRequest`, `api.DetectRequest`,
`api.DetectConfidence_*`). Source of truth is the generated `infracost.plugin`
package:

- `PluginService.GetPluginInfo` → `{ type(PARSER|PROVIDER), name, version, description, url, author }`
- `ParserService`: `GetParserConfig`, `IdentifyProjects`, `IdentifyEnvironments` (optional, may return `codes.Unimplemented`), `Parse`
  - `GetParserConfigResponse{ identification_priority(uint32), config_file_project_type(optional string) }`
  - `IdentifyProjectsResponse{ directory(bool), files[], dependency_paths[], raw_options(bytes JSON) }`
  - `ParseRequest{ path, generic_options(GenericOptions), raw_options(bytes JSON) }` → `ParseResponse{ diagnostics[], tree(tree.Tree), requested_dependencies[] }`
- `ProviderService`: `Process`, `ListFinopsPolicies`

Goal: rewrite docs + example to match the real `infracost.plugin` contract. Do NOT
add the fabricated RPCs to the proto.

## Tasks

### 1. Rewrite the example plugin
- [ ] Rewrite `parser/example/main.go` to import `github.com/infracost/proto/gen/go/infracost/plugin` instead of `.../parser/api`
- [ ] Register `PluginService` with `GetPluginInfo` returning `type = PLUGIN_TYPE... PARSER`, plus name/version/description/url/author
- [ ] Register `ParserService` implementing `GetParserConfig` (return an `identification_priority` and optional `config_file_project_type`)
- [ ] Implement `IdentifyProjects` for the example (recognize `.example` files; non-recursive; populate `files`/`directory`)
- [ ] Implement `IdentifyEnvironments` (return `codes.Unimplemented` or a single default environment) with an explanatory comment
- [ ] Implement `Parse` to return a minimal valid `tree.Tree` in `ParseResponse` (drop the CloudFormation oneof result)
- [ ] Remove references to `Describe`, `Detect`, `DetectConfidence`, `Initialize`, `ParseToTree`, and the cloudformation/options imports that no longer apply

### 2. Fix the example module
- [ ] Remove the stale "remove once the Describe/Detect RPCs are in a tagged proto release" comment in `parser/example/go.mod`
- [ ] Run `go mod tidy` and confirm `go build -o infracost-parser-plugin-example .` succeeds (against the real proto checkout)
- [ ] Keep the `replace github.com/infracost/proto => ../../../proto` directive (correct for the non-worktree checkout)

### 3. Rewrite parser/SPEC.md
- [ ] Replace the "gRPC Service Contract" section with the real `PluginService` + `ParserService` RPC list
- [ ] Remove the `Describe`, `Detect`, `Initialize`, and `ParseToTree` subsections
- [ ] Add `GetPluginInfo` section (metadata + PARSER/PROVIDER type as the discovery mechanism)
- [ ] Add `GetParserConfig` section documenting `identification_priority` (uint32) and `config_file_project_type`
- [ ] Add `IdentifyProjects` section (directory-scoped, non-recursive; `directory`/`files`/`dependency_paths`/`raw_options`)
- [ ] Add `IdentifyEnvironments` section (optional RPC, `codes.Unimplemented` semantics, dev/staging/prod variants)
- [ ] Rewrite `Parse` section: `path` + JSON `raw_options` in, `tree.Tree` out (remove `ParseRequestTarget`/`ParseResponseResult` oneofs)
- [ ] Remove the invented `kubernetes_supported_resources` / `SupportedResources` content
- [ ] Rewrite the priority guidance to use `identification_priority` semantics (higher = prioritised) instead of the fabricated 1-19/20-39/40+ bands
- [ ] Rewrite "Adding a New Format" to reflect JSON `raw_options` (no new oneof target/result messages required)
- [ ] Verify or remove the go-plugin handshake / magic-cookie block against the CLI's actual handshake config; mark unverifiable values clearly

### 4. Rewrite parser/README.md
- [ ] Fix the "Interface contract" RPC table to list the real RPCs (`GetPluginInfo`, `GetParserConfig`, `IdentifyProjects`, `IdentifyEnvironments`, `Parse`)
- [ ] Update the quick-start steps (replace "Implement Detect()/Parse()" with IdentifyProjects/Parse)
- [ ] Update the validation checklist to reference the real RPCs (no Describe/Detect/Initialize)

### 5. Fix root README.md
- [ ] Fix the parser RPCs cell in the comparison table to the real parser RPCs
- [ ] Note (comment or follow-up) that the provider RPCs cell (`Describe, ListSupportedResources, Process, ProcessTree, ListFinopsPolicies`) is also wrong — real `plugin.ProviderService` is `Process, ListFinopsPolicies` (+ `GetPluginInfo`); provider fix is out of scope for the aa2 (parser) worktree

### 6. Validation
- [ ] `go build` the example and confirm it compiles and starts
- [ ] Cross-check every RPC name, message name, and field name cited in SPEC.md/README.md against `gen/go/infracost/plugin/*` for 1:1 correspondence
- [ ] Grep the docs for the removed identifiers (`Describe`, `Detect`, `DetectConfidence`, `ParseToTree`, `Initialize`, `kubernetes_supported_resources`) to confirm none remain
- [ ] Also grep for the spec-forbidden strings: `ListSupportedResources`, `ProcessTree`, `plugins.infracost.io`, `INFRACOST_PARSER_PLUGIN_MAGIC_COOKIE`, `INFRACOST_PROVIDER_PLUGIN_MAGIC_COOKIE`, `ac92b06c592f`, `04d179d767fc`, `file_extensions`

## Verification Baseline (recorded per implementation-baseline.md acceptance criteria)
All contract claims below were verified on 2026-07-31 against the local sibling checkouts:

| Claim | Source of truth | Branch / value |
|-------|-----------------|----------------|
| Parser RPCs = `GetParserConfig, IdentifyProjects, IdentifyEnvironments, Parse` | `../../../proto/proto/infracost/plugin/parser.proto` | `feature/arm-plugin-architecture` |
| `PluginService.GetPluginInfo` + `PluginType{PARSER=1,PROVIDER=2}` | `.../plugin/plugin.proto` | same |
| Provider RPCs = `Process, ListFinopsPolicies` | `.../plugin/provider.proto` | same |
| `GetParserConfigResponse{identification_priority uint32, config_file_project_type optional string}` | parser.proto §GetParserConfigResponse | same |
| `ParseRequest{path, generic_options, raw_options bytes}`; `raw_options_format` reserved/dropped, always JSON | parser.proto §ParseRequest (field 4 reserved) | same |
| Handshake: `MagicCookieKey=INFRACOST_PLUGIN`, value `de8c7e96-497c-4168-80c4-fc875c8ce764`, `ProtocolVersion=1`, dispense key `"plugin"` | `../../../cli/pkg/plugins/manager.go:26-44`, `../../../config/plugin/list.go:19-28`, `../../../parser/internal/plugin/harness.go` | cli `main`, config `feature/arm-plugin-architecture`, parser `main` |
| Max gRPC message size 64 MiB | `../../../cli/pkg/plugins/consts/consts.go:6` (CLI call option) + parser harness server option | main |
| gRPC-only, `NetRPCUnsupportedPlugin` pattern | `manager.go:298` (`AllowedProtocols` gRPC), `harness.go:29` | — |

## Spec-derived refinements (not yet in tasks above)
1. **Build validation path caveat.** From inside this worktree the example's `replace github.com/infracost/proto => ../../../proto` resolves to a nonexistent dir, so `go build` fails with a *replacement-directory* error (not the `undefined: api.*` errors the gap analysis cited). Run the build from the canonical non-worktree checkout (`parser-plugin-sdk/parser/example`, where `../../../proto` = the real proto repo), or temporarily point the replace at the absolute proto path for the build check. Keep the committed relative `replace` unchanged.
2. **Handshake section must cover the shared, mandatory two-service pattern** (per handshake-documentation.md): every plugin registers `PluginService` (`GetPluginInfo`) **plus** `ParserService`; plugin type is resolved at runtime from `GetPluginInfoResponse.type`, never the cookie/binary name. State that `PluginService` registration is mandatory and why (type discovery). Show inline `plugin.Serve` — the repo's `Expose()` helpers live under `internal/` and are not importable by third-party plugins. Mention `LOG_LEVEL` env passthrough in the process contract.
3. **provider/SPEC.md handshake tension.** handshake-documentation.md wants *both* SPECs to show identical constants, but aa2 is the parser worktree and root README notes the provider fix is out of scope here. Resolution: fix parser/SPEC.md's handshake fully; for provider/SPEC.md, either apply the same handshake block if trivially safe, or flag it as a tracked follow-up for the provider worktree — do not leave the fictional per-type cookie in place uncorrected without a note.
4. **Do not overstate `identification_priority` runtime behavior** (baseline edge case): it is defined in proto but identification is driven via the `../config` library; document it as a plugin-provided hint, not as CLI-read production behavior.
5. **Legacy `infracost.parser.api` scoping** (baseline edge case): if referenced at all, scope it strictly to ../parser's dependency-extraction use; never present it as the plugin contract.
