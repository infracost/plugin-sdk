package server

import (
	"context"

	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
)

// version is a placeholder for the plugin's build version. A real plugin
// typically sets this at build time, e.g.:
//
//	go build -ldflags "-X github.com/you/your-plugin/server.version=$(git describe --tags)"
var version = "dev"

func (s *Server) GetPluginInfo(_ context.Context, _ *pluginpb.GetPluginInfoRequest) (*pluginpb.GetPluginInfoResponse, error) {
	return &pluginpb.GetPluginInfoResponse{
		// Name is this plugin's identity: "<namespace>/<name>". Official
		// plugins use the infracost/ namespace; community plugins should use
		// their own, e.g. "acme/my-format".
		Name:        "acme/my-plugin-name-goes-here",
		Version:     version,
		Description: "Add a description here",
		Url:         "https://github.com/acme/infracost-parser-my-plugin",
		Author:      "Add your name or company here",
		Type:        pluginpb.PluginType_PARSER,
	}, nil
}
