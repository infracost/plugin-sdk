# Root Readme Accuracy

## Overview
The top-level `README.md` is the SDK's front door. Its comparison table, naming section, and command examples are built on the abandoned design and must be rewritten to summarize the live contract without duplicating the SPECs.

## Requirements
- The parser-vs-provider comparison table must reflect reality:
  - RPCs: parser = `GetPluginInfo`, `GetParserConfig`, `IdentifyProjects`, `IdentifyEnvironments` (optional), `Parse`; provider = `GetPluginInfo`, `Process`, `ListFinopsPolicies`.
  - One shared magic cookie (`INFRACOST_PLUGIN`) — the per-type cookie row must go.
  - Binary naming: `infracost-parser-<key>` / `infracost-provider-<key>` as convention, with type resolved via `GetPluginInfo`.
  - Examples row: parsers Terraform/Terragrunt/CloudFormation/Kubernetes/CiscoStacks/Terraform-plan; providers AWS/Azure/Google/Kubernetes (ARM out, Kubernetes in).
- Remove the "Plugin naming" canonical-form/registry table and replace with the `<namespace>/<name>` convention (`infracost/` reserved-by-convention for official plugins).
- Remove the "Validation" section's `plugin validate` examples and the `plugin add` example; "Managing plugins" shows `plugin list` and `plugin update` only.
- Keep the two-plugin-type framing and links to `parser/` and `provider/` docs; each linked file must exist.

## Acceptance Criteria
- [ ] Every command shown in the root README executes successfully against the current CLI (or is clearly marked as sample output).
- [ ] Table values are consistent with both SPECs (no README/SPEC contradictions).
- [ ] No references to magic-cookie values, registries, or validation tooling that don't exist.

## Edge Cases
- The README should not enumerate proto field details (it will drift); it should link to the SPECs for anything below RPC level.

## Dependencies
- [implementation-baseline](implementation-baseline.md)
- [parser-spec-accuracy](parser-spec-accuracy.md)
- [provider-spec-accuracy](provider-spec-accuracy.md)
- [cli-commands-documentation](cli-commands-documentation.md)
- [discovery-and-naming-documentation](discovery-and-naming-documentation.md)
