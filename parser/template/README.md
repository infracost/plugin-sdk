# Parser plugin template

This is a snapshot of the parser-plugin scaffolding used **inside the
[`infracost/parser`](https://github.com/infracost/parser) repo** (`plugin/template`).
It shows the recommended structure for a real, production parser plugin:

```
template/
├── main.go                 # entrypoint: plugin.Expose(server.New())
├── options/options.go      # this plugin's own options (decoded from raw_options)
└── server/
    ├── server.go           # Server type implementing the plugin interface
    ├── get_plugin_info.go   ─┐
    ├── get_parser_config.go  ├─ the 4 RPC methods, one file each
    ├── identify_projects.go  │
    ├── parse.go             ─┘
    ├── *_test.go            # per-method tests + a golden-file Parse harness
    └── testdata/basic/…     # golden fixtures
```

## In-repo vs standalone

This template depends on helpers that live in the `infracost/parser` repo and is
meant to be copied **within that repo** (duplicate `plugin/template`, rename, and
implement):

- `internal/plugin.Expose` — wraps the go-plugin handshake + server registration
- `internal/plugin.URL` / `internal/plugin.Author` — shared plugin metadata
- `internal/version.Version` — build version
- `pkg/diagnostic`, `github.com/infracost/go-proto/pkg/tree` — tree + diagnostics

It therefore does **not** build on its own outside that repo.

**Writing a plugin in your own repo?** Start from
[`../example`](../example) instead — it is fully self-contained: it does the raw
[go-plugin](https://github.com/hashicorp/go-plugin) handshake inline (no
`internal` helpers), has its own `go.mod`, and builds standalone. See
[`../SPEC.md`](../SPEC.md) for the full contract.

## The four methods

| Method | Purpose |
| --- | --- |
| `GetPluginInfo` | Identity: name, version, description, URL, author, type (`PARSER`). |
| `GetParserConfig` | `identification_priority` (higher wins for the same files) and the optional `config_file_project_type`. |
| `IdentifyProjects` | Given a directory (non-recursive), report whether the whole dir is one project, and/or which files are projects. |
| `Parse` | Turn an identified path into an IaC-agnostic `tree.Tree`. |

## Options

A plugin can take its own options. The CLI/runner JSON-marshals a typed struct
and sends it on `ParseRequest.raw_options` with
`raw_options_format = "application/json"`; `Parse` unmarshals the bytes into this
plugin's `options.Options`. Keep `options.Options` empty for plugins that need
nothing (e.g. a Terraform-plan-JSON parser); a Terraform parser fills it with
source maps, vars, workspace, tfvars files, etc. See [`../SPEC.md`](../SPEC.md)
for the full options contract — and note the provider SPEC's matching
`TreeInput.raw_options` channel for **provider** plugins.
