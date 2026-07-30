# Implementation Plan — Ensure parser SDK documentation matches the live proto contract

## FINAL REVIEW ADDENDUM (2026-07-31, Claude final review — read first)
Despite Task 2's checkboxes reading `[x]`, `parser/template/` had **not actually been
touched** since its initial snapshot commit (`11db566`) — it was still verbatim
`../parser/plugin/template/` (CloudFormation sniffer in `identify_projects.go`, CFN
assertions in every `*_test.go` behind `t.Skip`, `internal/plugin`/`internal/version`/
`pkg/diagnostic`/`go-proto/pkg/tree` imports, no `go.mod`, `RawOptionsFormat` usage that
doesn't exist in the pinned proto). This contradicted `specs/template-alignment.md`'s
acceptance criteria (all still unchecked) and the "this repo is canonical" decision.
Fixed in this review pass, applying the plan's recommended **Option 1**:
- Added `parser/template/go.mod` (module `github.com/infracost/plugin-sdk/parser/template`,
  `github.com/infracost/proto v1.160.0`, no `replace`). Confirmed `go build/vet/test ./...`
  pass with `GOPROXY=off` (no siblings, no network).
- Rewrote `main.go` to inline the shared handshake (no `internal/plugin.Expose`).
- Rewrote `server/get_plugin_info.go` to drop `internal/plugin`/`internal/version`; inline
  placeholder metadata + a `version` var with an ldflags comment.
- Rewrote `server/identify_projects.go`: dropped the CFN sniffer entirely, replaced with a
  neutral marker-file directory-based placeholder (`template.config.json` → `Directory:
  true`), demonstrating the branch of `IdentifyProjectsResponse` that the file-based
  `example/` does not cover.
- Rewrote `server/parse.go` to use the public `infracost/proto/gen/go/infracost/tree`
  package (matching `example/`) instead of `go-proto/pkg/tree` + `parser/pkg/diagnostic`;
  it now returns a real placeholder resource instead of an empty tree + a TODO.
- Un-skipped all four `*_test.go` files, fixed their assertions to match the template's own
  placeholder behavior (no more CloudFormation identity/config assertions), dropped the
  `infracost/config` test dependency, renamed the options fixture
  `cloudformation.options.json` → `template.options.json`, dropped the nonexistent
  `RawOptionsFormat` field. Regenerated `testdata/basic/expected.json` via the suite's
  golden-file harness and confirmed a second run passes on comparison (not just the
  write-golden path).
- Rewrote `parser/template/README.md` to state canonicity (reconciled against
  `infracost/parser` `main` @ `c533a18`, 2026-07-31) instead of "does not build on its own
  outside that repo", and documented `IdentifyEnvironments` optionality.
- Added the missing single "start here" pointer (`example/` vs `template/`) to
  `parser/README.md` and `parser/SPEC.md`, which previously mentioned only `example/`.
- Fixed remaining doc drift found during the sweep: a stale "returns an empty tree" comment
  in `parser/example/main.go` (it returns one resource), a stray CloudFormation/**ARM**
  example pairing in `parser/SPEC.md`'s `IdentifyProjects` contract bullet (ARM is
  out-of-scope; swapped for Kubernetes), and the root `README.md` comparison table's
  Examples row (was TF/Terragrunt/CFN + AWS/Azure/GCP only; now includes
  Kubernetes/CiscoStacks/Terraform-plan for parsers and Kubernetes for providers, per
  `specs/root-readme-accuracy.md`).
- `provider/*` binary-naming drift (`infracost-provider-plugin-*`) remains, confirmed
  out-of-scope for this (aa2/parser) worktree per the existing scope note below.

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
- [x] Import `github.com/infracost/proto/gen/go/infracost/plugin` (drop `parser/api`,
      `parser/cloudformation`, `parser/options`).
- [x] Register **two** services on one gRPC server: `PluginService` (GetPluginInfo →
      `type = plugin.PluginType_PARSER`, name/version/description/url/author) **and**
      `ParserService`.
- [x] `GetParserConfig` → `identification_priority: 0`, optional
      `config_file_project_type` (comment that empty defaults to plugin name; use
      kubernetes' hardcoded-type comment as the model).
- [x] `IdentifyProjects` → recognise `.example` files; non-recursive; honour the
      `directory:true` XOR `files[]` mutual-exclusion rule.
- [x] `IdentifyEnvironments` → returns `codes.Unimplemented` via embedded
      `UnimplementedParserServiceServer` (valid "no environment support").
- [x] `Parse` → return a minimal **real** `tree.Tree` (≥1 provider→service→resource)
      in `ParseResponse`; drop the CloudFormation oneof.
- [x] Use the shared handshake constants (cookie `INFRACOST_PLUGIN` /
      `de8c7e96-497c-4168-80c4-fc875c8ce764`, dispense key `"plugin"`), inline
      `plugin.Serve` (the `Expose()` helpers are `internal/` — not importable).
- [x] No `Describe`/`Detect`/`DetectConfidence`/`Initialize`/`ParseToTree`/cloudformation
      remnants.

### 2. Align the example with the parser template + add tests
- [x] Layout decision: `example/` stays as a single-file walkthrough (not split into
      `server/`), per the canonical-template decision. `main_test.go` covers all RPCs
      in-process via `goplugin.TestPluginGRPCConn`.
- [x] Add per-RPC `_test.go` (table-driven), in-process via go-plugin's
      `TestPluginGRPCConn` dispensing `"plugin"`. Tests pass without `t.Skip` and
      cover each implemented RPC. No CloudFormation copy-paste.
- [x] **Decision (user, 2026-07-31):** this repo's `parser/template/` is the
      **canonical** template source; `../parser/plugin/template/` becomes a downstream
      copy (repointing it is a follow-up in that repo). Keep `example/` as the minimal
      single-file walkthrough and `template/` as the production-shaped starter; docs get
      a single "start here" pointer distinguishing the two. Template README must state
      canonicity and record the ../parser ref it was last reconciled with.
- [x] **Template implementation (final review, 2026-07-31):** the decision above had been
      recorded but not executed — `parser/template/` was still the unmodified CFN snapshot.
      Fixed per `specs/template-alignment.md` Option 1: added standalone `go.mod` (proto
      v1.160.0, no replace), dropped all `internal/`/`go-proto` imports, replaced the CFN
      sniffer with a neutral directory-marker placeholder, un-skipped and fixed all tests,
      dropped `raw_options_format`/`RawOptionsFormat`, rewrote the template README to state
      canonicity (reconciled against `../parser` main @ `c533a18`). Verified `go build/vet/
      test ./...` pass with `GOPROXY=off` (no siblings, no network). See the FINAL REVIEW
      ADDENDUM at the top of this file for the full list.

### 3. Fix `parser/example/go.mod` + `Makefile`
- [x] **go.mod decision (applied):** bumped to `github.com/infracost/proto v1.160.0`
      with **no** `replace` directive. `go build ./...` succeeds from a fresh clone with
      no siblings. Stale comment removed. `go mod tidy` run.
- [x] `go mod tidy` run; stale comment dropped.
- [x] Makefile: `BINARY` renamed to `infracost-parser-example`; `test` target added
      (`go test ./...`); `build` and `install` targets retained.
- [x] `go build ./...` + `go vet ./...` + `go test ./...` all pass in `parser/example/`.

### 4. Rewrite `parser/SPEC.md`
- [x] **Architecture overview:** replaced Describe/Detect/Initialize/ParseToTree flow
      diagram with correct `infracost.plugin` contract flow.
- [x] **gRPC service contract:** `PluginService.GetPluginInfo` +
      `ParserService.{GetParserConfig, IdentifyProjects, IdentifyEnvironments, Parse}`.
      No Describe/Detect/Initialize/ParseToTree/SupportedResources subsections.
- [x] **GetPluginInfo** section: metadata + `type` (`PARSER`) as the discovery/type
      mechanism. No `display_name`/`file_extensions`/`supports_directories`.
- [x] **GetParserConfig** section: `identification_priority` (uint32, higher-wins,
      recommend 0) with real values in Priority subsection (terragrunt 1 over terraform 0).
- [x] **IdentifyProjects** section: directory-scoped, non-recursive; `directory`/`files`
      (mutually exclusive)/`dependency_paths`/`raw_options`.
- [x] **IdentifyEnvironments** section: optional RPC; `codes.Unimplemented` vs empty-list
      semantics; `attributed_files` = Terraform/Terragrunt-only migration aid (others
      ignore); `environments[]` shape.
- [x] **Parse** section: `path` + `generic_options` + JSON `raw_options` in; `tree.Tree`
      + `diagnostics` + `requested_dependencies` out. No `ParseRequestTarget`/
      `ParseResponseResult` oneofs. Proto field 4 noted as reserved/dropped.
- [x] **`raw_options` lifecycle** section (new): seeded by `IdentifyProjects` → refined
      per-environment by `IdentifyEnvironments` → persisted in config file → passed
      verbatim into `ParseRequest.raw_options`. Always JSON; no forbidden strings.
- [x] **Plugin naming:** `<namespace>/<name>` convention; `infracost/` reserved by
      convention; names unique.
- [x] **Binary naming:** `infracost-parser-<format>` convention (no `-plugin-`).
- [x] **Handshake** section: shared cookie/protocol/dispense-key/64 MiB.
- [x] **Constraints:** 64 MiB message size; non-recursive identification.
- [x] **Adding a New Format:** JSON `raw_options` (no proto changes needed).
- [x] **Reference implementations:** added kubernetes, ciscostacks, terraform-plan.

### 5. Rewrite `parser/README.md`
- [x] "What is a parser plugin": no ARM; discovery via `GetPluginInfo`; identification
      via config autodetect flow.
- [x] Quick start: `example/` as starting point; covers GetPluginInfo/GetParserConfig/
      IdentifyProjects/Parse; no Describe/Detect/`plugin validate`.
- [x] Interface-contract table → real RPCs including `IdentifyEnvironments (optional)`.
- [x] "Testing" section present (unit tests + drop-in-and-run).
- [x] Binary naming `infracost-parser-<key>`; `IdentifyEnvironments` optionality noted.

### 6. Fix `README.md` (root)
- [x] Comparison table: RPCs row now includes `IdentifyEnvironments (optional)` for
      parser; binary naming corrected to `infracost-parser-<name>` /
      `infracost-provider-<name>`; one shared `INFRACOST_PLUGIN` cookie.
- [x] Plugin naming: `<namespace>/<name>` convention; `infracost/` reserved by convention.
- [x] No `plugin validate`/`plugin add` present; no Validation section.
- [x] Two-plugin framing retained; links to `parser/` and `provider/` docs present.
- [x] Provider-specific fixes noted as provider-worktree follow-up.

### 7. Baseline note (implementation-baseline acceptance)
- [x] "Verification Baseline" table present and accurate; version claims verified
      against `../parser/go.mod` (v1.160.0) and sibling repo checkouts.

### 8. Validation
- [x] `go build ./...` + `go vet ./...` + `go test ./...` pass in `parser/example/`
      (confirmed in this session against the worktree checkout).
- [x] RPC/message/field names cross-checked against proto — all consistent.
- [x] Grep for all forbidden strings across all in-scope files: zero occurrences.
      No `plugin validate`/`plugin add` occurrences anywhere.
- [x] No README/SPEC contradictions: binary naming, cookie, RPC set, priority values
      are consistent across README.md (root), parser/README.md, parser/SPEC.md, and
      parser/example/.

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
