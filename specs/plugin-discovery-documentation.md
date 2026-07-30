# Plugin Discovery Documentation

## Overview
The SDK docs must accurately explain how the CLI and config library locate, launch, and accept plugin binaries.

## Requirements
- Document that the CLI launches every executable in the plugin directory and calls `GetPluginInfo`; there is no filename-pattern filter.
- Document that directories and sidecar files (`.sha256`, `.version`) are skipped, while binaries that fail to launch, handshake, or return plugin information are skipped.
- Document the default plugin directory as `os.UserCacheDir()/infracost/plugins`, including useful per-OS examples.
- Document `INFRACOST_CLI_PLUGIN_DIR` as the plugin-directory override and explain that setting it disables automatic plugin installation.
- Document the related configuration variables only where they affect discovery or installation: `INFRACOST_CLI_PLUGIN_BASE_URL`, `INFRACOST_CLI_PLUGIN_CACHE_DIRECTORY`, `INFRACOST_CLI_PLUGIN_AUTO_UPDATE`, and `INFRACOST_CLI_PLUGIN_<KEY>_VERSION`.
- Distinguish the two discovery paths: the CLI loads parsers and providers, while the `github.com/infracost/config` discovery path retains parser plugins only.
- Document the official plugin set managed by the CLI: parsers `terraform`, `terragrunt`, `cloudformation`, `ciscostacks`, `terraform-plan`, `kubernetes`; providers `aws`, `google`, `azure`, `kubernetes`.

## Acceptance Criteria
- [ ] The default plugin directory and environment-variable overrides match `../cli/pkg/plugins/config.go` and `../config/plugin/list.go`.
- [ ] Docs state that all executable candidates are probed and plugin type is obtained from `GetPluginInfo`.
- [ ] Docs distinguish candidate-skipping behavior from fatal duplicate-identity errors.
- [ ] Docs do not claim filename-based type detection.
- [ ] ARM is not listed as an installed parser.

## Edge Cases
- Windows appends `.exe` to installed binary names and does not require the Unix executable-bit check.
- A development build reporting version `dev` is not automatically updated.
- Kubernetes parser and provider binaries can be installed while their execution remains feature-gated.
- Discovery failures generally appear as a plugin missing from `plugin list`; authors need `LOG_LEVEL` guidance to inspect the underlying failure.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [handshake-documentation](handshake-documentation.md)
- [plugin-naming-documentation](plugin-naming-documentation.md)
