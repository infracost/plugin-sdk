// Package options holds this plugin's own options — the typed shape it expects
// in ParseRequest.raw_options. The CLI/runner JSON-marshals a matching struct
// and sends it with raw_options_format = "application/json"; Parse unmarshals
// the bytes back into Options. Leave it empty for plugins that need no options
// (e.g. a Terraform-plan-JSON parser); a Terraform parser fills it with source
// maps, vars, workspace, tfvars files, etc.
package options

type Options struct{}
