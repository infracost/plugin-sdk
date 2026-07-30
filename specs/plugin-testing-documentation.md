# Plugin Testing Documentation

## Overview
The SDK docs must give plugin authors a testing workflow that uses the verification mechanisms available today.

## Requirements
- Document Go unit tests against the service implementation, following the official parser and provider repositories' fixture-based test patterns.
- Demonstrate go-plugin's in-process `TestPluginGRPCConn` harness with dispense key `"plugin"` so both `PluginService` and the type-specific service are exercised through gRPC.
- Document a manual end-to-end check: build the binary, install it into the plugin directory, run `infracost plugin list`, and run the CLI against a representative project.
- Enumerate the implicit load-time checks: executable status on non-Windows systems, successful shared handshake, non-nil `GetPluginInfo`, `GetParserConfig` for parser plugins, and unique returned name plus type.
- Explain that a discovered plugin still needs an RPC-level functional test against representative inputs.
- Ensure example Makefiles provide working `build`, `test`, and install-to-plugin-directory targets.
- Treat execution against an already-installed, compatible `infracost` CLI as an optional smoke test. The required automated verification is the in-process gRPC harness plus module build, vet, and test; implementation must not build sibling repositories solely to perform the smoke test.

## Acceptance Criteria
- [ ] Parser and provider onboarding docs each contain a testing section covering unit and manual end-to-end checks.
- [ ] Each example has in-process tests for every RPC it intentionally implements.
- [ ] Example Makefile targets execute successfully without relying on nonexistent CLI commands.
- [ ] Manual instructions explain how to diagnose a binary that does not appear in `plugin list`.
- [ ] A compatible installed CLI is used for a smoke test when readily available; its absence does not fail acceptance when the harness and documented workflow are complete.

## Edge Cases
- On Windows, discovery does not use the Unix executable-bit check.
- A plugin binary can load and appear in `plugin list` while its parsing or pricing behavior is still incorrect.
- Discovery skips broken candidate binaries, so authors should enable CLI logging via `LOG_LEVEL` when a plugin is absent.
- Optional parser `IdentifyEnvironments` should be tested either for implemented behavior or for the intentional `codes.Unimplemented` response.
- An installed CLI can differ from the sibling source baseline; do not treat behavior from an incompatible installed version as authoritative.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [handshake-documentation](handshake-documentation.md)
- [plugin-discovery-documentation](plugin-discovery-documentation.md)
