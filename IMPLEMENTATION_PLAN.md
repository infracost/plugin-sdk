# Implementation Plan — Ensure parser SDK documentation matches the live proto contract

## STATUS ADDENDUM (2026-07-31, post-rebase — read first)
The branch has been **rebased onto `update-plugin-sdk-refactor`**, which already rewrote
the docs/examples for the `infracost.plugin` contract. That resolves most of the
"Verified current-state gaps" below (they described the old spike base) and most of
Tasks 1, 4, 5, 6 in their original form. **Residual work on the new base:**

1. `raw_options_format` no longer exists in the proto (`ParseRequest` field 4 reserved;
   `raw_options` is always JSON). Remove it from: `parser/SPEC.md` (2×),
   `parser/template/README.md`, `parser/template/options/options.go`,
   `parser/template/server/parse.go` (+ `provider/SPEC.md` 2×, see scope note).
2. `IdentifyEnvironments` is missing everywhere — add to `parser/SPEC.md` (full RPC
   section incl. Unimplemented-vs-empty semantics and `attributed_files` caveat),
   `parser/README.md` RPC table, root README parser RPC row, template docs.
3. `IdentifyProjectsResponse.raw_options` (seed blob) and the raw_options lifecycle
   (identify → environments → config YAML → ParseRequest, always JSON) are undocumented.
4. Both examples' `go.mod` still pin proto `v1.34.0` with the broken
   `replace ../../../proto` — bump to **v1.160.0**, drop replace (Open decision 1: RESOLVED).
5. Example tests per Task 2 still needed; template has CFN copy-paste defects
   (`identify_projects.go` sniffer, CFN test assertions, `parse_test.go` reading
   `cloudformation.options.json`) and `t.Skip`-guarded tests.
6. **Decisions from the user (2026-07-31):**
   - Sibling repos: assume `origin/main` (or `master`) is correct and working; verify by
     reading code, do NOT build sibling repos.
   - `plugin validate` / `plugin add` WILL be added to the CLI — tracked as issues
     **cli-b10** and **cli-6f8** in ../cli's tracker. Docs may reference them only as
     clearly-labeled future work citing those issues.
   - **This repo is the canonical template source** (reverses Task 2's recommendation):
     keep `parser/template/` here as canonical, clean it up, record the ../parser ref it
     was reconciled with; pointing ../parser's copy back here is a follow-up in that repo.

## GAP-ANALYSIS REFRESH (2026-07-31, planning phase — verified against working tree)
Re-verified every residual item in the STATUS ADDENDUM against the actual files. All
still hold. Precise, confirmed current state:

**parser/example/ (single-file, self-contained):**
- `go.mod:7,12` — still `infracost/proto v1.34.0` + `replace => ../../../proto` (unresolvable
  from this worktree → build/vet/test all fail). Fix: bump to tagged `v1.160.0`, drop replace,
  drop the stale comment, `go mod tidy`.
- `Makefile:3` — `BINARY = infracost-parser-plugin-example` (wrong `-plugin-` convention →
  `infracost-parser-example`); no `test` target. `validate` already gone; `install` OK.
- `main.go:137` — comment still says raw_options is "encoded as req.RawOptionsFormat"; the
  field no longer exists (proto field 4 reserved). Reword to "always JSON". Code itself is
  clean and on the correct contract. `Parse` returns an **empty** `tree.Tree{}` — spec wants
  ≥1 provider→service→resource. No `_test.go` yet (needs one test per RPC, in-process).

**parser/SPEC.md:**
- Line 184 — `raw_options_format` row still in the Parse request table; line 217 — "with
  raw_options_format = \"application/json\"". Remove both; state always-JSON.
- No `### IdentifyEnvironments` section (service list at 31/108 omits it). Add full section.
- No `raw_options` lifecycle section. Priority values (0/1/10) and architecture prose are OK.

**parser/README.md:** clean of forbidden strings BUT interface-contract table (42–52) omits
`IdentifyEnvironments (optional)`; build snippet (line 24) uses `infracost-parser-plugin-example`.

**README.md (root):** discovery/handshake/naming prose is correct, but three concrete drifts —
line 12 RPC row omits `IdentifyEnvironments (optional)`; line 35 uses
`infracost-parser-plugin-<name>` / `infracost-provider-plugin-<name>`; line 66 `go build -o
infracost-parser-plugin-myformat`. Examples row lists only TF/Terragrunt/CFN + AWS/Azure/GCP
(spec wants Kubernetes/CiscoStacks/Terraform-plan added, ARM stays out).

**parser/template/ — the largest residual, and it hides an unresolved design tension:**
Confirmed defects: `identify_projects.go` is verbatim CFN sniffer; `get_plugin_info.go:6-7`
imports `infracost/parser/internal/{plugin,version}`; `get_plugin_info_test.go:13,21` has
`t.Skip` + asserts `"infracost/cloudformation"`; `parse_test.go:20,85,95` has `t.Skip`, reads
`cloudformation.options.json`, sets `RawOptionsFormat`; `parse.go:7-9` imports
`infracost/go-proto/pkg/tree` + `infracost/parser/pkg/diagnostic`; README frames the copy as a
*downstream snapshot of ../parser* (opposite of the decided canonical direction) and cites
`raw_options_format` (54) + a `TreeInput.raw_options` provider channel (57-58) that the proto
lacks. There is **no `go.mod`**.

## OPEN DECISION (blocks template-alignment acceptance) — needs resolution before/at build
`template-alignment.md` requires "Template tests pass (`go test ./...`) without `t.Skip`", but
the template currently **cannot compile standalone**: it has no `go.mod` and imports three
../parser-internal packages (`internal/plugin`, `internal/version`, `pkg/diagnostic`) plus
`infracost/go-proto/pkg/tree`. These are unreachable from a fresh clone with no siblings. You
cannot both (a) keep the template a faithful in-repo mirror of ../parser's `plugin/template`
(internal deps, no module) AND (b) have its tests build+pass here. Options:
  1. **Make the template a standalone module** — add `go.mod` (proto `v1.160.0`, no replace),
     drop the internal imports: inline metadata constants for `plugin.URL/Author/version`,
     replace `go-proto/pkg/tree`+`pkg/diagnostic` usage with the public
     `infracost/proto/gen/go/infracost/tree` types the example already uses. Then un-skip and
     fix the CFN test assertions → real passing tests. (Diverges from ../parser's copy, but the
     decision says *this* repo is canonical, so ../parser becomes the downstream that adapts.)
  2. **Keep it non-buildable** and satisfy the spec via the self-contained `example/` tests
     only, explicitly recording in the template README that the template is a structural
     reference whose tests run once copied into a module. This contradicts the literal
     acceptance criterion, so it must be a stated, signed-off deviation.
Recommend **Option 1** — it is the only path that satisfies the acceptance criterion and the
"this repo is canonical" decision simultaneously; the internal helpers it drops are trivial
(two string consts + a version var + swapping to the already-used public tree package).
Whichever is chosen, the template README must be rewritten to declare canonicity and record the
../parser ref it was reconciled against (its git status is unchanged from the snapshot commit).

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
Claims below were gathered from the local sibling checkouts. **Policy (user decision):
`origin/main` (or `master`) of each sibling repo is assumed correct and working — verify
by reading code at that ref, do not build sibling repos.** For the wire contract, the
tagged release the examples build against (`github.com/infracost/proto` v1.160.0, the
version ../parser pins) is authoritative; note the local ../proto checkout sits on
`feature/arm-plugin-architecture`, so re-check any claim sourced there against the tag.

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

> **Refreshed 2026-07-31 (gap-analysis phase).** A prior session already did most of
> the destructive rewrite: the example `main.go`, the SPEC/README structure, and the
> root README's discovery/naming/handshake prose are now on the correct
> `infracost.plugin` contract (no `Describe`/`Detect`/`Initialize`/`ParseToTree`/
> `cloudformation`/`parser/api`/registry/per-type-cookie remnants). The bullets below
> record only what is **still wrong today**, verified against the working tree.

### Still broken — must fix in the implementation phase
- **parser/example/go.mod (BLOCKS build/vet/test):** still `github.com/infracost/proto
  v1.34.0` + `replace github.com/infracost/proto => ../../../proto`, which **does not
  resolve** from this worktree → `go build/vet/test ./...` all fail with "replacement
  directory ../../../proto does not exist". Apply the RESOLVED decision: bump to the
  tagged `v1.160.0` (confirmed present in `../parser/go.mod` and `../parser/go.sum`),
  drop the `replace`, drop the stale "remove once the infracost.plugin protos are in a
  tagged proto release" comment, `go mod tidy`.
- **parser/example/Makefile:** `BINARY = infracost-parser-plugin-example` uses the wrong
  `-plugin-` convention → should be `infracost-parser-example`. `validate` target is
  already gone and `install` is correct, but there is **no `test` target** — add
  `test: go test ./...`.
- **parser/SPEC.md — `raw_options_format`:** still referenced in the Parse request table
  (line ~184) and the "Adding a New Format" step (line ~217). Remove the field row and
  the `with raw_options_format = "application/json"` phrasing; state that `raw_options`
  is **always JSON** and proto field 4 (`raw_options_format`) is reserved/dropped.
- **parser/SPEC.md — missing `### IdentifyEnvironments`:** the service now documents
  GetPluginInfo/GetParserConfig/IdentifyProjects/Parse but omits the optional
  `IdentifyEnvironments` RPC. Add a section: optional; `codes.Unimplemented` vs
  empty-list semantics; `attributed_files` = Terraform/Terragrunt migration aid (others
  ignore); `environments[]` (Environment{name,path,files,dependency_paths,raw_options}).
- **README.md (root) — binary naming:** line ~35 (`infracost-parser-plugin-<name>` /
  `infracost-provider-plugin-<name>`) and line ~66 (`go build -o
  infracost-parser-plugin-myformat`) still carry the wrong `-plugin-` segment → real is
  `infracost-parser-<key>` / `infracost-provider-<key>`.
- **README.md (root) — comparison table RPC row (line ~12):** parser column reads
  `GetPluginInfo · GetParserConfig · IdentifyProjects · Parse` — **missing
  `IdentifyEnvironments (optional)`**. Add it. (Provider column `GetPluginInfo · Process
  · ListFinopsPolicies` is correct; the shared-cookie/naming/registry cleanup the plan
  called for is already done.)

### Verify during implementation (structure looks correct; confirm details)
- **parser/example/main.go:** correct proto import + shared handshake; relies on
  `UnimplementedParserServiceServer` so `IdentifyEnvironments` auto-returns
  `codes.Unimplemented` (a valid "no env support"). Optional: add a one-line comment
  making that intentional (Task 1). No test file / `server/` split yet — see Task 2
  decision (link-to-template vs minimal example).
- **parser/SPEC.md GetParserConfig:** confirm the body cites real priority values
  (tf/cfn/k8s/ciscostacks 0, terragrunt 1, terraform-plan 10) and that
  `generic_options` path `infracost/parser/options/options.proto` is correct.
- **parser/README.md:** no forbidden strings; confirm the interface-contract table
  lists `IdentifyEnvironments` as optional (Task 5) and the "start here" pointer is
  single/unambiguous.
- **README.md (root):** confirm the comparison table (if still present) has correct RPC
  rows, the single shared `INFRACOST_PLUGIN` cookie, and Kubernetes-in/ARM-out examples;
  and that `plugin validate`/`plugin add` are absent (Task 6).

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
- [x] **Decision (user, 2026-07-31):** this repo's `parser/template/` is the
      **canonical** template source; `../parser/plugin/template/` becomes a downstream
      copy (repointing it is a follow-up in that repo). Keep `example/` as the minimal
      single-file walkthrough and `template/` as the production-shaped starter; docs get
      a single "start here" pointer distinguishing the two. Template README must state
      canonicity and record the ../parser ref it was last reconciled with.

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
      errors when `INFRACOST_CLI_PLUGIN_DIR` is set). A short labeled future-work note
      may mention that `plugin validate`/`plugin add` are planned (CLI issues **cli-b10**
      and **cli-6f8**) — never present them as existing.
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
      `ac92b06c592f`, `04d179d767fc`, `--fixture`, `-debug`, `raw_options_format`,
      `cmd/infracost-parser-plugin`. For `plugin validate`/`plugin add`: no occurrences
      outside a clearly-labeled future-work note citing cli-b10/cli-6f8.
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
2. **example vs template/** (Task 2): **RESOLVED (user, 2026-07-31)** — this repo hosts
   the canonical `parser/template/`; keep the minimal `example/` alongside it. See the
   Task 2 decision bullet and the template-alignment spec.
