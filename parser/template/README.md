# Parser plugin template

This is the **canonical parser plugin template**. It shows the recommended
structure for a real, production parser plugin, split into one file per RPC:

```
template/
├── go.mod                  # standalone module: no siblings, no replace directives
├── main.go                 # entrypoint: goplugin.Serve(...) + Server registration
├── options/options.go      # this plugin's own options (decoded from raw_options)
└── server/
    ├── server.go            # Server type implementing the plugin interface
    ├── get_plugin_info.go   ─┐
    ├── get_parser_config.go  ├─ the RPC methods, one file each
    ├── identify_projects.go  │  (IdentifyEnvironments is optional — see below)
    ├── parse.go             ─┘
    ├── *_test.go            # per-method tests + a golden-file Parse harness
    └── testdata/basic/…     # golden fixtures
```

## Standalone and buildable

Unlike a plain snapshot, this copy is a real Go module: `go build ./...`,
`go vet ./...`, and `go test ./...` all pass here with no sibling repositories
and no `replace` directives. It imports only public packages — no
`internal/` package from another module.

```bash
go build ./...
go vet ./...
go test ./...
```

**Start here** if you want a production-shaped starting point to copy into
your own repo: rename the module in `go.mod`, replace the placeholder logic in
`server/` (metadata, identification, parsing) with your format's real
behaviour, and update the tests and `testdata/` fixtures to match.

If you just want the smallest possible working example to read end-to-end
before copying anything, see [`../example`](../example) instead — a single
`main.go` with no package split. Use `template/` when you're ready to build a
real plugin; use `example/` to learn the contract first.

## Canonical source

This repo (`parser-plugin-sdk`) is the canonical home of this template.
`infracost/parser`'s `plugin/template` is a downstream copy; this copy was
last reconciled against `infracost/parser` `main` @ `c533a18` (2026-07-31).
Repointing that copy back at this repo is a follow-up tracked in the
`infracost/parser` repo, not done here.

## The RPCs

| RPC | Purpose |
| --- | --- |
| `GetPluginInfo` | Identity: name, version, description, URL, author, type (`PARSER`). |
| `GetParserConfig` | `identification_priority` (higher wins for the same files) and the optional `config_file_project_type`. |
| `IdentifyProjects` | Given a directory (non-recursive), report whether the whole dir is one project, and/or which files are projects. |
| `IdentifyEnvironments` *(optional)* | Refine an identified project into named environments. Not implemented here — `Server` embeds `pluginpb.UnimplementedParserServiceServer`, so this RPC correctly returns `codes.Unimplemented`, which the CLI treats as "this plugin has no environment concept" (distinct from returning an empty list). Implement it if your format has an equivalent of Terraform workspaces or var-file sets; see the `infracost/parser` repo's `kubernetes`/`terraform` plugins for real implementations. |
| `Parse` | Turn an identified path into an IaC-agnostic `tree.Tree`. |

`identify_projects.go` demonstrates the `directory: true` branch (a marker
file claims the whole directory, like Terraform's `*.tf` files); the
file-based `files[]` branch (like CloudFormation, where each file is its own
project) is demonstrated instead in [`../example`](../example). Pick whichever
matches your format — the two are mutually exclusive per response.

## Options

A plugin can take its own options. `ParseRequest.raw_options` is **always
JSON** (proto field 4, `raw_options_format`, is reserved and dropped) — the
CLI/runner JSON-marshals a typed struct and sends it on
`ParseRequest.raw_options`; `Parse` unmarshals the bytes into this plugin's
`options.Options`. Keep `options.Options` empty for plugins that need nothing
(e.g. a Terraform-plan-JSON parser); a Terraform parser fills it with source
maps, vars, workspace, tfvars files, etc. See [`../SPEC.md`](../SPEC.md) for
the full options and `raw_options` lifecycle.
