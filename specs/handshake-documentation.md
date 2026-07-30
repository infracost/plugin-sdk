# Handshake Documentation

## Overview
The go-plugin handshake sections in `parser/SPEC.md` and `provider/SPEC.md` must describe the single shared handshake that all Infracost plugins actually use, instead of the fictional per-type handshakes currently documented.

## Requirements
- Document one handshake shared by parser and provider plugins:
  - `MagicCookieKey: "INFRACOST_PLUGIN"`
  - `MagicCookieValue: "de8c7e96-497c-4168-80c4-fc875c8ce764"`
  - `ProtocolVersion: 1`
  - Dispense key (plugin map key): `"plugin"` — not `"parser"` or `"provider"`
  - gRPC server options `MaxRecvMsgSize`/`MaxSendMsgSize` of 64 MiB (matches the CLI's client-side `MaxCallRecvMsgSize`/`MaxCallSendMsgSize`)
- Document that every plugin registers **two** gRPC services on the same server: `PluginService` (with `GetPluginInfo`) plus either `ParserService` or `ProviderService`, and that plugin type is resolved at runtime from `GetPluginInfoResponse.type` — never from the cookie or the binary name.
- Handshake code samples must compile against the generated bindings in `github.com/infracost/proto/gen/go/infracost/plugin`.
- Ground truth to match: `../parser/internal/plugin/harness.go`, `../providers/internal/plugin/harness.go`, `../cli/pkg/plugins/manager.go`, `../config/plugin/list.go`.
- Note that the reference `Expose()` serve helpers in ../parser and ../providers live under `internal/` and cannot be imported by third-party plugins; the SDK docs/examples must show the equivalent inline `plugin.Serve` code.

## Acceptance Criteria
- [ ] Both SPECs show identical handshake constants matching `../cli/pkg/plugins/manager.go`.
- [ ] No mention of `INFRACOST_PARSER_PLUGIN_MAGIC_COOKIE`, `INFRACOST_PROVIDER_PLUGIN_MAGIC_COOKIE`, or the values `ac92b06c592f`/`04d179d767fc`.
- [ ] Dispense key documented as `"plugin"` in both SPECs.
- [ ] Both SPECs state that `PluginService` registration is mandatory and explain why (type discovery).
- [ ] Handshake code samples are lifted from (or verified equal to) the working examples in this repo.

## Edge Cases
- Only gRPC is allowed (`AllowedProtocols` in the CLI is gRPC-only); plugins embedding `NetRPCUnsupportedPlugin` should be shown as the pattern.
- The ../config library client sets no max message size — the 64 MiB limit is a CLI-side call option plus plugin-side server option; docs should attribute the limit correctly.
- Plugins receive `LOG_LEVEL` from the CLI's environment for verbosity passthrough; worth documenting as part of the process contract if the docs discuss logging.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
