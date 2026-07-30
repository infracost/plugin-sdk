# Plugin Naming Documentation

## Overview
The SDK docs must describe the implemented plugin identity and binary-name conventions without presenting abandoned registry naming as runtime behavior.

## Requirements
- Document a plugin's runtime identity as the `name` returned by `GetPluginInfo` plus its plugin `type`.
- Document bare `<namespace>/<name>` values as the naming convention, with `infracost/` used by official plugins.
- State that plugin names must be unique within a type among installed plugins; duplicate name-and-type identities are a load error.
- Document official binary names as `infracost-parser-<key>` and `infracost-provider-<key>`, without a `-plugin-` segment.
- Clearly distinguish binary naming as a packaging convention from discovery: the CLI does not infer plugin identity or type from a filename.
- Document `infracost-plugin-<key>` as legacy naming only if needed to explain why the CLI removes such binaries.
- Remove registry-host resolution, short-form expansion, `plugins.infracost.io`, double-dash namespace flattening, `-debug` suffix rules, and canonical registry names.

## Acceptance Criteria
- [ ] No occurrence of `plugins.infracost.io`, canonical registry naming, or short-form expansion remains in current-behavior documentation.
- [ ] Binary naming is consistently shown as `infracost-parser-<key>` or `infracost-provider-<key>`.
- [ ] Docs explicitly distinguish runtime identity from the binary filename.
- [ ] Duplicate-identity behavior matches the CLI implementation.

## Edge Cases
- Windows distribution filenames append `.exe`.
- Two plugins may share a returned name only when their plugin types differ.
- Third-party binaries may use descriptive filenames that do not follow the official convention and can still be discovered.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [handshake-documentation](handshake-documentation.md)
