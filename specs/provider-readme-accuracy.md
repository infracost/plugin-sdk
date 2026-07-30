# Provider Readme Accuracy

## Overview
`provider/README.md` must provide a concise, accurate onboarding path for provider authors and defer wire-level detail to the provider specification.

## Requirements
- Explain that provider plugins receive the shared tree and add pricing plus FinOps results; they do not parse IaC source.
- List the live interface: mandatory `GetPluginInfo`, `Process`, and `ListFinopsPolicies`.
- State that every loaded provider receives the tree and is responsible for ignoring or pruning resources it does not handle.
- Present `provider/example/` as both the minimal walkthrough and the de-facto community provider starter because no separate provider template exists.
- Use the `infracost-provider-<key>` binary convention and actual plugin-directory installation flow.
- Link to the provider specification for message fields, handshake details, pricing strategies, and FinOps policy shapes.
- Include the current unit-test and manual end-to-end testing workflow.
- List all four current provider implementations: AWS, Azure, Google, and Kubernetes.

## Acceptance Criteria
- [ ] The interface table contains only `GetPluginInfo`, `Process`, and `ListFinopsPolicies`.
- [ ] Quick-start commands build and install the example using its actual binary name.
- [ ] The README clearly identifies the example as the available provider starter.
- [ ] No `Describe`, `ProcessTree`, `ListSupportedResources`, registry name, per-type handshake, or unsupported CLI command appears.
- [ ] Detailed contract claims link to `provider/SPEC.md` instead of duplicating field tables.

## Edge Cases
- A provider with no matching resources should return an empty output rather than fail the entire processing pass.
- Kubernetes is both a parser and provider plugin; the provider list should identify its role clearly.
- Community providers may use hardcoded prices or external APIs and need not use Infracost Cloud credentials.

## Dependencies
- [provider-spec-accuracy](provider-spec-accuracy.md)
- [provider-example-accuracy](provider-example-accuracy.md)
- [plugin-testing-documentation](plugin-testing-documentation.md)
- [plugin-naming-documentation](plugin-naming-documentation.md)
