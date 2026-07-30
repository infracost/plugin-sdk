# Implementation Plan — Ensure parser SDK documentation matches the live proto contract

## Context
The parser SDK docs and example describe an invented 5-RPC interface
(`Describe, Detect, Initialize, Parse, ParseToTree` with `DetectConfidence`,
`file_extensions`, Kubernetes supported resources, format-specific oneofs) plus a
registry/canonical-name scheme, per-type magic cookies, and `plugin validate`/`add`
tooling — **none of which exist**. The example imports the legacy
`infracost/parser/api` package, returns a CloudFormation-shaped oneof, and does not
compile against the current proto module.

The authoritative contract is the `infracost.plugin` proto package. This plan
rewrites the parser docs + example to match it. **Do NOT** add the fabricated RPCs
to the proto.

### Scope (aa2 = parser worktree)
In scope: `parser/SPEC.md`, `parser/README.md`, `parser/example/*`, and the
`README.md` (root) sections that concern parsers or are shared (naming, handshake,
CLI commands, comparison table). Out of scope: `provider/SPEC.md`,
`provider/README.md`, `provider/example/*` — but the root README's provider row and
shared handshake **are** corrected here, and a provider-worktree follow-up is noted
where a provider-only file would need the same fix.

## Source of truth (verified 2026-07-31)
All claims verified against the local sibling checkouts. Proto ground truth lives at
`/Users/gg/Development/infracost/proto` (branch `feature/arm-plugin-architecture`;
must also hold on the tagged release ../cli and ../parser consume, v1.159.0+).

- `plugin/plugin.proto` — **`PluginService.GetPluginInfo`** →
  `GetPluginInfoResponse{ type(PluginType: PLUGIN_TYPE_UNKNOWN=0, PARSER=1, PROVIDER=2), name, version, description, url, author }`.
- `plugin/parser.proto` — **`ParserService`**: `GetParserConfig`, `IdentifyProjects`,
  `IdentifyEnvironments` (optional; `codes.Unimplemented` = no env support, distinct
  from empty list), `Parse`.
  - `GetParserConfigResponse{ identification_priority(uint32, higher=prioritised, recommend 0), config_file_project_type(optional string, defaults to plugin name) }`
  - `IdentifyProjectsRequest{ directory }` → `IdentifyProjectsResponse{ directory(bool), files[] (mutually exclusive with directory:true), dependency_paths[], raw_options(bytes JSON) }`. Non-recursive.
  - `IdentifyEnvironmentsRequest{ directory, attributed_files[] (AttributedVarFile; Terraform/Terragrunt migration aid, others ignore), raw_options(bytes JSON) }` → `IdentifyEnvironmentsResponse{ environments[] (Environment{ name, path, files[], dependency_paths[], raw_options }) }`
  - `ParseRequest{ path, generic_options(GenericOptions), raw_options(bytes, **always JSON**; field 4 `raw_options_format` reserved/dropped) }` → `ParseResponse{ diagnostics[], tree(tree.Tree), requested_dependencies[] (Dependency) }`
- `plugin/provider.proto` — **`ProviderService`**: `Process`, `ListFinopsPolicies` (+ `GetPluginInfo`).
- **Handshake** (shared by ALL plugins): `MagicCookieKey="INFRACOST_PLUGIN"`,
  `MagicCookieValue="de8c7e96-497c-4168-80c4-fc875c8ce764"`, `ProtocolVersion=1`,
  dispense/plugin-map key `"plugin"`, gRPC-only (`NetRPCUnsupportedPlugin`), 64 MiB
  send/recv. Source: `../cli/pkg/plugins/manager.go`, `../config/plugin/list.go`,
  `../parser/internal/plugin/harness.go`.
- **Real priority values** (`../parser/plugin/*/server/get_parser_config.go`):
  terraform/cloudformation/kubernetes/ciscostacks = 0, terragrunt = 1 ("higher than
  terraform"), terraform-plan = 10.
- **Template layout** (`../parser/plugin/template/`): `main.go`, `options/options.go`,
  `server/{server,get_plugin_info,get_parser_config,identify_projects,parse}.go` +
  per-RPC `_test.go` + `testdata/`. Note: template has **no** `identify_environments.go`.

### Legacy interfaces — do NOT present as the plugin contract
- `infracost/parser/api` (`infracost.parser.api`): old in-process parser
  (`Initialize`/`Parse`/`ParseToTree`, format oneofs). Still imported by ../parser
  **for dependency extraction only** — scope any mention to that use.
- `infracost.provider` legacy `ProcessTree`: not served/dialed.

## Verified current-state gaps (what's wrong today)
- **README.md (root):** wrong RPC rows; per-type magic-cookie row; binary naming
  `infracost-parser-plugin-<name>` (wrong — real is `infracost-parser-<key>`, no
  `-plugin-`); `plugins.infracost.io` registry/canonical-name table; ARM listed /
  Kubernetes+Terragrunt+CiscoStacks+Terraform-plan missing; `plugin validate` +
  `plugin add` examples (neither command exists).
- **parser/SPEC.md:** entirely the abandoned design — Describe/Detect/Initialize/
  ParseToTree, DetectConfidence, `file_extensions`/`supports_directories`,
  registry/canonical names, `-debug` suffix, per-type cookie
  `INFRACOST_PARSER_PLUGIN_MAGIC_COOKIE`/`ac92b06c592f`, priority bands 1-19/20-39/40+
  and values 10/25/30, `kubernetes_supported_resources`, `<100ms` detection & `10s`
  startup limits, `cmd/…` + ARM reference paths, `plugin validate`/`--fixture`.
- **parser/README.md:** five-RPC table, Describe/Detect quick-start, `plugin validate`
  + `--fixture` validation section, ARM mention, filename-pattern discovery claim.
- **parser/example/main.go:** imports `infracost/parser/api` +
  `parser/cloudformation` + `parser/options`; implements Describe/Detect/Initialize/
  Parse/ParseToTree; returns CloudFormation oneof; uses per-type cookie. Does not
  compile against the current proto (undefined `api.DescribeRequest` etc.).
- **parser/example/go.mod:** `infracost/proto v1.34.0` + stale "remove once the
  Describe/Detect RPCs are in a tagged proto release" comment + `replace … => ../../../proto`
  (path does **not** resolve from this worktree — confirmed `ls ../../../proto` fails;
  the real proto is `/Users/gg/Development/infracost/proto`).
- **parser/example/Makefile:** `validate` target invokes the nonexistent
  `infracost plugin validate`; no `test` target.

## Tasks

### 1. Rewrite `parser/example/main.go`
- [ ] Import `github.com/infracost/proto/gen/go/infracost/plugin` (drop `parser/api`,
      `parser/cloudformation`, `parser/options`).
- [ ] Register **two** services on one gRPC server: `PluginService` (GetPluginInfo →
      `type = plugin.PluginType_PARSER`, name/version/description/url/author) **and**
      `ParserService`.
- [ ] `GetParserConfig` → `identification_priority: 0`, optional
      `config_file_project_type` (comment that empty defaults to plugin name; use
      kubernetes' hardcoded-type comment as the model).
- [ ] `IdentifyProjects` → recognise `.example` files; non-recursive; honour the
      `directory:true` XOR `files[]` mutual-exclusion rule.
- [ ] `IdentifyEnvironments` → either return `codes.Unimplemented` with a comment that
      this is valid ("no environment support"), or a single default environment.
- [ ] `Parse` → return a minimal **real** `tree.Tree` (≥1 provider→service→resource)
      in `ParseResponse`; drop the CloudFormation oneof.
- [ ] Use the shared handshake constants (cookie `INFRACOST_PLUGIN` /
      `de8c7e96-497c-4168-80c4-fc875c8ce764`, dispense key `"plugin"`), inline
      `plugin.Serve` (the `Expose()` helpers are `internal/` — not importable).
- [ ] No `Describe`/`Detect`/`DetectConfidence`/`Initialize`/`ParseToTree`/cloudformation
      remnants.

### 2. Align the example with the parser template + add tests
- [ ] Mirror `../parser/plugin/template/` layout/idiom: `main.go` + `server/` with one
      file per RPC (`server.go`, `get_plugin_info.go`, `get_parser_config.go`,
      `identify_projects.go`, `parse.go`), plus `options/` and `testdata/` as needed.
- [ ] Add per-RPC `_test.go` (table-driven), in-process via go-plugin's
      `TestPluginGRPCConn` dispensing `"plugin"`. Tests must **pass** (not `t.Skip`) and
      cover each implemented RPC. No CloudFormation copy-paste (imports, sniffer logic,
      `infracost/cloudformation` assertions).
- [ ] **Decision to record:** SDK `example/` vs a `template/` copy — link to
      `../parser/plugin/template/` as canonical and keep only a minimal `example/`
      (avoid shipping two divergent starters). Docs get a single "start here" pointer;
      carry a "sourced from parser@<ref>" note if any template snapshot is shipped.

### 3. Fix `parser/example/go.mod` + `Makefile`
- [ ] **go.mod decision (record explicitly):** prefer bumping to the tagged
      `github.com/infracost/proto` release that contains `infracost.plugin` (the version
      ../parser pins, v1.159.0+) with **no** `replace`, so `go build ./...` succeeds from
      a fresh clone with no siblings (example-accuracy acceptance criterion). If that tag
      does not yet publish the `infracost.plugin` package, degrade to a `replace` with an
      **accurate** comment and a path that works from a fresh clone — this must be a
      stated, recorded decision, not the current broken `../../../proto`.
- [ ] `go mod tidy`; drop the stale Describe/Detect comment.
- [ ] Makefile: remove the `validate` target; keep `build`; add `test` (`go test ./...`)
      and an `install` target that copies the binary into the plugin dir (or a documented
      manual verification step). All targets must succeed end-to-end.
- [ ] Build-check caveat: from this worktree a relative `replace` to `../../../proto`
      does not resolve; run the build from the canonical checkout
      (`/Users/gg/Development/infracost/parser-plugin-sdk/parser/example`) or temporarily
      point `replace` at the absolute proto path — do not commit that absolute path.

### 4. Rewrite `parser/SPEC.md`
- [ ] **Architecture overview:** replace the Describe/Detect/Initialize/ParseToTree flow
      diagram. Attribute identification to the `infracost/config` autodetect flow (config
      drives `IdentifyProjects`/`IdentifyEnvironments`; the CLI scan path calls `Parse`) —
      don't imply the CLI calls every RPC inline.
- [ ] **gRPC service contract:** `PluginService.GetPluginInfo` +
      `ParserService.{GetParserConfig, IdentifyProjects, IdentifyEnvironments, Parse}`.
      Remove Describe/Detect/Initialize/ParseToTree/SupportedResources subsections.
- [ ] **GetPluginInfo** section: metadata + `type` (`PARSER`) as the discovery/type
      mechanism. No `display_name`/`file_extensions`/`supports_directories`.
- [ ] **GetParserConfig** section: `identification_priority` (uint32, higher-wins,
      recommend 0) with **real** values (terraform/cfn/k8s/ciscostacks 0, terragrunt 1,
      terraform-plan 10 — use terragrunt-over-terraform as the example); note it is a
      plugin-provided hint enforced via the ../config flow, **not** read by ../cli
      production code. `config_file_project_type` (optional; defaults to plugin name;
      kubernetes hardcodes `"kubernetes"` — cite it).
- [ ] **IdentifyProjects** section: directory-scoped, non-recursive; `directory`/`files`
      (mutually exclusive)/`dependency_paths`/`raw_options`.
- [ ] **IdentifyEnvironments** section: optional RPC; `codes.Unimplemented` vs empty-list
      semantics; `attributed_files` = Terraform/Terragrunt-only migration aid (others
      ignore); `environments[]` shape.
- [ ] **Parse** section: `path` + `generic_options` + JSON `raw_options` in; `tree.Tree`
      + `diagnostics` + `requested_dependencies` out. Remove `ParseRequestTarget`/
      `ParseResponseResult` oneofs.
- [ ] **`raw_options` lifecycle** section (new): seeded by `IdentifyProjects` → refined
      per-environment by `IdentifyEnvironments` → persisted as a readable YAML map in the
      config file → passed verbatim into `ParseRequest.raw_options`. Always JSON; no
      `raw_options_format`; schema owned/documented by the plugin.
- [ ] **Plugin naming:** remove registry/canonical/`plugins.infracost.io`/double-dash.
      Identity = `GetPluginInfo.name` + `type`; bare `<namespace>/<name>` by convention
      (`infracost/` reserved-by-convention for official); names unique (duplicate =
      fatal at load).
- [ ] **Binary naming + discovery:** `infracost-parser-<key>` (no `-plugin-`; `.exe` on
      Windows; `infracost-plugin-<key>` is legacy the CLI removes; `-debug` handling
      gone). Discovery launches **every** executable in the plugin dir and calls
      `GetPluginInfo` (no filename filter; dirs/`.sha256`/`.version` skipped; failed
      handshakes silently skipped). Default dir `os.UserCacheDir()/infracost/plugins`;
      env overrides `INFRACOST_CLI_PLUGIN_DIR` (implies skip auto-install),
      `INFRACOST_CLI_PLUGIN_BASE_URL`, `_CACHE_DIRECTORY`, `_AUTO_UPDATE`, `_<KEY>_VERSION`.
- [ ] **Handshake** section: shared cookie/protocol/dispense-key/64 MiB (per
      handshake-documentation.md); state that registering **both** `PluginService` and
      `ParserService` is mandatory and why (runtime type discovery); type never from
      cookie/filename; `NetRPCUnsupportedPlugin` gRPC-only; inline `plugin.Serve`
      (Expose helpers internal); `LOG_LEVEL` env passthrough. Samples must compile
      against `gen/go/infracost/plugin`.
- [ ] **Constraints:** keep only code-verifiable limits — 64 MiB message size; plugin
      start timeout 60s (180s Windows); `GetPluginInfo` query timeout 30s; non-recursive
      identification. Remove `<100ms` detection and `10s` startup.
- [ ] **Testing your plugin** section (replaces Validation): Go unit tests
      (`server/*_test.go` + `testdata/`, `TestPluginGRPCConn` dispensing `"plugin"`);
      drop-in-and-run manual check; enumerate load-time checks (exec bit, handshake,
      non-nil `GetPluginInfo`, `GetParserConfig` for parsers, unique name+type); note
      load failures are silent skips — check CLI logs via `LOG_LEVEL`.
- [ ] **Adding a New Format:** JSON `raw_options` (no new oneof target/result messages);
      point at `../parser/plugin/template/`.
- [ ] **Reference implementations:** `../parser/plugin/{terraform,terragrunt,terraform-plan,
      cloudformation,kubernetes,ciscostacks}/` (main.go + server/ one file per RPC).
      Remove `cmd/…` and ARM.

### 5. Rewrite `parser/README.md`
- [ ] "What is a parser plugin": remove ARM; discovery via `GetPluginInfo` (not filename
      pattern); identification via the config autodetect flow.
- [ ] Quick start: copy the example / point at the template; implement GetPluginInfo/
      GetParserConfig/IdentifyProjects/Parse (+ optional IdentifyEnvironments). Remove
      Describe/Detect/`plugin validate`.
- [ ] Interface-contract table → the real RPCs.
- [ ] Replace the "Validation" section with "Testing your plugin" (unit tests +
      drop-in-and-run), consistent with SPEC.md.
- [ ] Binary naming `infracost-parser-<key>`; single unambiguous "start here" pointer
      (per template-alignment); note `IdentifyEnvironments` optionality.

### 6. Fix `README.md` (root)
- [ ] Comparison table:
  - RPCs — parser: `GetPluginInfo, GetParserConfig, IdentifyProjects, IdentifyEnvironments (optional), Parse`;
    provider: `GetPluginInfo, Process, ListFinopsPolicies`.
  - Replace the per-type magic-cookie row with one shared `INFRACOST_PLUGIN`.
  - Binary naming: `infracost-parser-<key>` / `infracost-provider-<key>` (convention;
    type resolved via `GetPluginInfo`).
  - Examples: parsers Terraform/Terragrunt/CloudFormation/Kubernetes/CiscoStacks/
    Terraform-plan; providers AWS/Azure/Google/Kubernetes (ARM out, Kubernetes in).
- [ ] Remove the "Plugin naming" canonical-form/registry table → `<namespace>/<name>`
      convention (`infracost/` reserved by convention).
- [ ] Remove the "Validation" `plugin validate` section and the `plugin add` example;
      "Managing plugins" shows `plugin list` + `plugin update` only (note `plugin update`
      errors when `INFRACOST_CLI_PLUGIN_DIR` is set).
- [ ] Keep two-plugin framing + links to `parser/` and `provider/` docs (each must exist).
- [ ] Provider-specific SPEC/README/example fixes remain a **provider-worktree
      follow-up**; the root README provider row and shared handshake are corrected here.

### 7. Baseline note (implementation-baseline acceptance)
- [ ] Keep/refresh the "Verification Baseline" table (below) recording repo/branch/
      version each claim was verified against.

### 8. Validation
- [ ] `go build ./...` + `go vet ./...` + `go test ./...` pass in `parser/example/`
      (from the canonical checkout, or the recorded fresh-clone path).
- [ ] Cross-check every RPC/message/field name cited in SPEC.md + README.md +
      parser/README.md against `gen/go/infracost/plugin/*` for 1:1 correspondence.
- [ ] Grep all in-scope docs/Makefile for forbidden strings and confirm **zero**:
      `Describe`, `Detect`, `DetectConfidence`, `Initialize`, `ParseToTree`, `ProcessTree`,
      `ListSupportedResources`, `SupportedResources`, `kubernetes_supported_resources`,
      `file_extensions`, `supports_directories`, `plugins.infracost.io`,
      `INFRACOST_PARSER_PLUGIN_MAGIC_COOKIE`, `INFRACOST_PROVIDER_PLUGIN_MAGIC_COOKIE`,
      `ac92b06c592f`, `04d179d767fc`, `plugin validate`, `plugin add`, `--fixture`,
      `-debug`, `raw_options_format`, `cmd/infracost-parser-plugin`.
- [ ] Confirm no README/SPEC contradictions (naming, cookie, RPC set, priority values).

## Verification Baseline (verified 2026-07-31 against local sibling checkouts)
| Claim | Source of truth | Branch / value |
|-------|-----------------|----------------|
| Parser RPCs = `GetParserConfig, IdentifyProjects, IdentifyEnvironments, Parse` | `proto/infracost/plugin/parser.proto` | proto `feature/arm-plugin-architecture` |
| `PluginService.GetPluginInfo` + `PluginType{UNKNOWN=0,PARSER=1,PROVIDER=2}` | `plugin/plugin.proto` | same |
| Provider RPCs = `Process, ListFinopsPolicies` (+ GetPluginInfo) | `plugin/provider.proto` | same |
| `GetParserConfigResponse{identification_priority uint32, config_file_project_type optional string}` | parser.proto | same |
| `ParseRequest{path, generic_options, raw_options bytes}`; field 4 `raw_options_format` reserved; always JSON | parser.proto | same |
| Priority values 0 (tf/cfn/k8s/ciscostacks), 1 (terragrunt), 10 (terraform-plan) | `../parser/plugin/*/server/get_parser_config.go` | parser HEAD |
| Handshake `INFRACOST_PLUGIN` / `de8c7e96-497c-4168-80c4-fc875c8ce764` / proto v1 / key `"plugin"` | `../cli/pkg/plugins/manager.go`, `../config/plugin/list.go`, `../parser/internal/plugin/harness.go` | cli main / config feature/arm / parser main |
| 64 MiB message size; gRPC-only (`NetRPCUnsupportedPlugin`) | `../cli/pkg/plugins/consts/consts.go`, `manager.go` (`AllowedProtocols`), parser `harness.go` | main |
| Binary naming `infracost-parser-<key>` (no `-plugin-`); discovery launches all binaries + GetPluginInfo | `../cli/pkg/plugins/required.go`, `manager.go`, `../config/plugin/list.go` | main / feature/arm |
| Template layout `../parser/plugin/template/` (main.go + options/ + server/ one file per RPC; no identify_environments.go) | `../parser/plugin/template/` | parser HEAD |

## Open decisions to confirm during build
1. **Proto dependency form** (Task 3): **RESOLVED (verified 2026-07-31)** — use the
   tagged `github.com/infracost/proto v1.160.0` with **no** `replace`. Evidence: this is
   the exact version `../parser/go.mod` pins, it is present in `../parser/go.sum`
   (`v1.160.0 h1:qPH7Afh1…`), and `../parser/plugin/template/server/*.go` import
   `github.com/infracost/proto/gen/go/infracost/plugin` against it — so the tagged
   release publishes the `infracost.plugin` package and resolves from a fresh clone with
   no siblings. The relative-`replace` fallback is therefore NOT needed; drop the current
   broken `../../../proto` replace entirely.
2. **example vs template/** (Task 2): recommend link-to-canonical-template + minimal
   example; confirm before shipping any snapshot.
