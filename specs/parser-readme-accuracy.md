# Parser Readme Accuracy

## Overview
`parser/README.md` must provide a concise, accurate onboarding path for parser authors and defer wire-level detail to the parser specification.

## Requirements
- Explain that parser plugins identify projects and convert IaC input into the shared cost tree; they do not calculate prices.
- List the live interface: mandatory `GetPluginInfo`, `GetParserConfig`, `IdentifyProjects`, and `Parse`, plus optional `IdentifyEnvironments`.
- Attribute project and environment identification to the config-library autodetection flow rather than implying every RPC is called inline by the CLI.
- Provide one unambiguous starting path: use `parser/example/` for a minimal walkthrough and `parser/template/` for a production-shaped starter.
- Use the `infracost-parser-<key>` binary convention and actual plugin-directory installation flow.
- Link to the parser specification for message fields, handshake details, priorities, and `raw_options`.
- Include the current unit-test and manual end-to-end testing workflow.
- List only parser implementations that exist in the CLI-managed set.

## Acceptance Criteria
- [ ] The interface table includes `IdentifyEnvironments` and labels it optional.
- [ ] Quick-start commands build and install the example using its actual binary name.
- [ ] Example and template roles are distinct and consistent with their own READMEs.
- [ ] No abandoned RPC, registry name, per-type handshake, unsupported CLI command, or ARM parser appears.
- [ ] Detailed contract claims link to `parser/SPEC.md` instead of duplicating field tables.

## Edge Cases
- Returning `codes.Unimplemented` from `IdentifyEnvironments` is valid and should not make a basic parser appear incomplete.
- Kubernetes is both a parser and provider plugin; the parser list should identify its role clearly.

## Dependencies
- [parser-spec-accuracy](parser-spec-accuracy.md)
- [template-alignment](template-alignment.md)
- [plugin-testing-documentation](plugin-testing-documentation.md)
- [plugin-naming-documentation](plugin-naming-documentation.md)
