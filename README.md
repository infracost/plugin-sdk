# Infracost Plugin SDK

Build custom plugins to extend [Infracost](https://infracost.io) with new IaC formats and cloud providers.

Infracost has two types of plugins:

| | Parser Plugin | Provider Plugin |
|---|---|---|
| **Purpose** | Extract resources from IaC files | Price extracted resources |
| **Binary naming** | `infracost-parser-plugin-<name>` | `infracost-provider-plugin-<name>` |
| **Magic cookie** | `INFRACOST_PARSER_PLUGIN_MAGIC_COOKIE` | `INFRACOST_PROVIDER_PLUGIN_MAGIC_COOKIE` |
| **RPCs** | Describe, Detect, Initialize, Parse, ParseToTree | Describe, ListSupportedResources, Process, ProcessTree, ListFinopsPolicies |
| **Examples** | Terraform, CloudFormation, ARM/Bicep | AWS, Azure, GCP |
| **Use case** | "I have a new IaC format" | "I have a new cloud to price" |

## Getting started

- **Parser plugins** — See [parser/README.md](parser/README.md) and [parser/SPEC.md](parser/SPEC.md)
- **Provider plugins** — See [provider/README.md](provider/README.md) and [provider/SPEC.md](provider/SPEC.md)

Each directory contains a working example you can copy as a starting point.

## Plugin naming

Every plugin has a canonical name in `registry/namespace/name` format:

| Short form | Canonical form |
|---|---|
| `terraform` | `plugins.infracost.io/infracost/terraform` |
| `acme/crossplane` | `plugins.infracost.io/acme/crossplane` |
| `plugins.example.com/acme/pulumi` | `plugins.example.com/acme/pulumi` |

If the first segment contains a dot, it's treated as a registry host. Otherwise the default registry (`plugins.infracost.io`) is used. The `infracost/` namespace is reserved for official plugins.

Binary filenames flatten the namespace: `infracost-parser-plugin-acme--crossplane` maps to canonical name `plugins.infracost.io/acme/crossplane`. Official plugins omit the namespace: `infracost-parser-plugin-terraform`.

## Validation

The Infracost CLI auto-detects the plugin type from the binary name and runs the appropriate conformance checks:

```bash
# Validate a parser plugin
infracost plugin validate ./infracost-parser-plugin-myformat

# Validate a provider plugin
infracost plugin validate ./infracost-provider-plugin-myprovider
```

## Managing plugins

```bash
# List installed plugins
infracost plugin list

# Show installation instructions for a new plugin
infracost plugin add parser myformat
infracost plugin add provider myprovider
```
