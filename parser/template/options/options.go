// Package options holds this plugin's own options — the typed shape it expects
// in ParseRequest.raw_options. raw_options is always JSON, so the CLI/runner
// JSON-marshals a matching struct and sends it as ParseRequest.raw_options;
// Parse unmarshals the bytes back into Options. Leave it empty for plugins
// that need no options (e.g. a Terraform-plan-JSON parser); a Terraform
// parser fills it with source maps, vars, workspace, tfvars files, etc.
package options

type Options struct{}
