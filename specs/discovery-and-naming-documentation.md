# Discovery and Naming Documentation

## Overview
The docs' plugin-naming scheme (registry hosts, canonical `plugins.infracost.io/namespace/name` names, double-dash filename encoding, filename-based type detection) does not exist anywhere in the implementation. The docs must describe how plugins are actually discovered, identified, and named.

## Requirements
- Discovery: the CLI (and the ../config library) launch **every executable** in the plugin directory and call `GetPluginInfo`; there is no filename-pattern filter. Directories and sidecar files (`.sha256`, `.version`) are skipped; binaries that fail to launch or handshake are skipped.
- Plugin directory: default is `os.UserCacheDir()/infracost/plugins` (per-OS paths worth listing); overridable via `INFRACOST_CLI_PLUGIN_DIR` (which also implies skipping auto-install). Related env vars: `INFRACOST_CLI_PLUGIN_BASE_URL` (default `https://releases.infracost.io`), `INFRACOST_CLI_PLUGIN_CACHE_DIRECTORY`, `INFRACOST_CLI_PLUGIN_AUTO_UPDATE`, `INFRACOST_CLI_PLUGIN_<KEY>_VERSION`.
- Identity: a plugin's identity is the `name` returned by `GetPluginInfo` plus its `type`. Names are bare `<namespace>/<name>` strings by convention (`infracost/terraform`, `infracost/aws`); the `infracost/` namespace is used by official plugins; names must be unique across installed plugins (duplicates are a fatal error at load).
- Official binary names are `infracost-parser-<key>` and `infracost-provider-<key>` (NO `-plugin-` segment) — matching `../cli/pkg/plugins/required.go` and both repos' `plugins-release.yml`. `infracost-plugin-<key>` is the legacy naming that the CLI actively removes. A descriptive name is recommended for third-party plugins but not required for discovery.
- Windows: `.exe` appended to installed binary names; exec-bit check skipped on Windows.
- Remove entirely: registry-host name resolution, short-form expansion rules, `plugins.infracost.io`, double-dash namespace flattening, `-debug` suffix handling, and any claim that the CLI auto-detects plugin type from the binary name.
- Document the official plugin set the CLI requires/installs: parsers `terraform`, `terragrunt`, `cloudformation`, `ciscostacks`, `terraform-plan`, `kubernetes`; providers `aws`, `google`, `azure`, `kubernetes`.

## Acceptance Criteria
- [ ] No occurrence of `plugins.infracost.io` or "canonical name"/"registry" naming anywhere in the SDK docs.
- [ ] Binary naming documented as `infracost-parser-<key>` / `infracost-provider-<key>` with the convention-vs-requirement distinction stated.
- [ ] Default plugin directory and the env-var overrides documented and verified against `../cli/pkg/plugins/config.go` and `../config/plugin/list.go`.
- [ ] Docs state that type detection happens via `GetPluginInfo`, with binaries that fail handshake skipped.
- [ ] ARM is not listed as an existing plugin (no ARM parser exists in ../parser HEAD or the CLI's required set).

## Edge Cases
- The ../config library's discovery keeps only PARSER-type plugins (providers are filtered out in that path); the CLI handles both. Docs describing "the discovery flow" should not conflate the two.
- Dev builds: version `dev` is never auto-updated by the CLI — useful guidance for plugin authors iterating locally.
- Kubernetes parser/provider plugins are downloaded but feature-gated at execution time; docs should avoid promising they always run.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [handshake-documentation](handshake-documentation.md)
