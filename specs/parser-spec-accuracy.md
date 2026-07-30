# Parser Spec Accuracy

## Overview
`parser/SPEC.md` must document the live parser plugin contract: the `infracost.plugin.ParserService` (plus mandatory `PluginService`), not the abandoned `Describe`/`Detect`/`Initialize`/`Parse`/`ParseToTree` design.

## Requirements
- Document the actual RPC set from `../proto/proto/infracost/plugin/parser.proto`:
  - `GetParserConfig` → `identification_priority` (uint32, **higher = offered first**, recommended 0) and optional `config_file_project_type` (maps to `github.com/infracost/config` project types; defaults to plugin name).
  - `IdentifyProjects` → request `directory`; response `directory` (bool), `files` (mutually exclusive with `directory: true`), `dependency_paths`, and `raw_options` (JSON seed blob, opaque to the caller). Non-recursive contract.
  - `IdentifyEnvironments` → **optional** RPC (returning `codes.Unimplemented` means "no environment support"); request `directory`, `attributed_files` (Terraform/Terragrunt-only migration aid others must ignore), `raw_options`; response `environments[]` with `name`, `path`, `files`, `dependency_paths`, `raw_options`. Distinguish "Unimplemented" from "empty list" semantics.
  - `Parse` → request `path`, `generic_options`, `raw_options` (**always JSON**; `raw_options_format` was removed from the proto and must not be documented); response `tree.Tree`, `diagnostics`, `requested_dependencies`.
- Document `GetPluginInfo` (from `plugin.proto`): `type` (must be `PARSER`), `name`, `version`, `description`, `url`, `author`. No `display_name`, `file_extensions`, or `supports_directories` fields exist.
- Priority guidance must match reality: real plugins use 0 (terraform, cloudformation, kubernetes, ciscostacks), 1 (terragrunt, to outrank terraform), 10 (terraform-plan). Remove the 10/25/30 scheme and the "lower = first" rule.
- Remove: detection-confidence levels, `content`/`content_provided` LSP fields, `SupportedResources`/`Initialize`, format-specific target/result oneofs, "detection < 100ms" and "startup 10s" limits (actual plugin start timeout is 60s, 180s on Windows; `GetPluginInfo` query timeout 30s).
- The `raw_options` lifecycle deserves an explicit section: seeded by `IdentifyProjects`, refined per-environment by `IdentifyEnvironments`, persisted in the config file as a readable YAML map, passed verbatim into `ParseRequest.raw_options`; always JSON, schema owned and documented by the plugin.
- Reference implementations section must point at `../parser` repo's real layout: `plugin/{terraform,terragrunt,terraform-plan,cloudformation,kubernetes,ciscostacks}/` (a `main.go` plus `server/` package with one file per RPC). No `cmd/` directories, no ARM plugin.
- Architecture overview must attribute project identification to the config-library flow (the `infracost/config` module drives `IdentifyProjects`/`IdentifyEnvironments` during autodetect; the CLI's scan path calls `Parse`), rather than implying the CLI calls every RPC inline.

## Acceptance Criteria
- [ ] Every RPC, message, field name, and field semantics in `parser/SPEC.md` exists in the tagged proto release the example builds against (spot-verified against `parser.proto`).
- [ ] Priority semantics documented as higher-wins with real values from `../parser/plugin/*/server/get_parser_config.go`.
- [ ] `IdentifyEnvironments` documented, including optionality and Unimplemented-vs-empty semantics.
- [ ] `raw_options` documented as always-JSON with no `raw_options_format` field.
- [ ] Constraints section contains only limits verifiable in code (64 MiB message size, 60s/180s start timeout, 30s GetPluginInfo timeout, non-recursive fast identification).
- [ ] Reference-implementation paths resolve in ../parser HEAD.

## Edge Cases
- `identification_priority` is defined in the proto but not read by ../cli production code; identification ordering is exercised via the ../config flow. The doc's claim about ordering must be phrased to match where it is actually enforced (verify in ../config during implementation).
- Terragrunt-over-Terraform is the one real-world priority example; use it instead of invented ARM/CloudFormation values.
- The kubernetes parser hardcodes a literal `"kubernetes"` project type with an explanatory comment — a good example for the `config_file_project_type` docs.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [handshake-documentation](handshake-documentation.md)
- [discovery-and-naming-documentation](discovery-and-naming-documentation.md)
