# CLI Commands Documentation

## Overview
The SDK docs promise CLI tooling (`infracost plugin validate`, `infracost plugin add`, `--fixture` flags, a conformance test suite) that does not exist. Docs must describe only the commands the CLI actually ships, and give plugin authors a real testing story in place of the fictional validator.

## Requirements
- Document the real `infracost plugin` subcommands from `../cli/internal/cmds/plugin.go`: `plugin list` (groups installed plugins into Parsers/Providers; no flags) and `plugin update` (no flags; errors when `INFRACOST_CLI_PLUGIN_DIR` is set).
- Remove all references to `infracost plugin validate`, `infracost plugin add`, `--fixture`/`--fixtures`, and the itemized conformance checks in the READMEs, SPECs, and example Makefiles (both example Makefiles currently have a `validate` target that invokes the nonexistent command).
- Replace the validation story with what actually verifies a plugin today:
  - Go unit tests against the service implementation (the pattern used by every plugin in ../parser: `server/*_test.go` with `testdata/` fixtures, plus go-plugin's `TestPluginGRPCConn` in-process harness dispensing key `"plugin"`).
  - Manual end-to-end check: drop the binary into the plugin directory and run the CLI against a project; the implicit load-time checks (executable bit, handshake, non-nil `GetPluginInfo`, `GetParserConfig` for parsers, unique name+type) are what a plugin must pass to load.
- `plugin validate` and `plugin add` are now planned CLI features, tracked in the CLI project as issues `cli-b10` (validate) and `cli-6f8` (add). SDK docs must not present them as existing commands; a clearly-labeled future-work note referencing those issues is acceptable and preferred over silent removal, so readers know the gap is acknowledged.

## Acceptance Criteria
- [ ] No SDK doc or Makefile invokes `plugin validate`, `plugin add`, or fixture flags as if they exist; any mention is a labeled future-work note citing `cli-b10`/`cli-6f8`.
- [ ] `plugin list` and `plugin update` are documented with behavior matching ../cli.
- [ ] A "testing your plugin" section exists covering unit-test and drop-in-and-run approaches, with the load-time checks enumerated.
- [ ] Example Makefiles' targets all work end-to-end (build, test, install-to-plugin-dir).

## Edge Cases
- `plugin update` intentionally fails when a custom plugin dir is set — worth stating so authors don't file bugs.
- Load-time failures are silent skips during discovery (a broken binary just doesn't appear in `plugin list`); docs should tell authors to check CLI logs (`LOG_LEVEL`) when their plugin doesn't show up.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [discovery-and-naming-documentation](discovery-and-naming-documentation.md)
