package server

import (
	"context"

	"github.com/infracost/parser/internal/plugin"
	"github.com/infracost/parser/internal/version"
	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
)

func (s *Server) GetPluginInfo(_ context.Context, _ *pluginpb.GetPluginInfoRequest) (*pluginpb.GetPluginInfoResponse, error) {
	return &pluginpb.GetPluginInfoResponse{
		Name:        "infracost/my-plugin-name-goes-here",
		Version:     version.Version,
		Description: "Add a description here",
		Url:         plugin.URL,
		Author:      plugin.Author,
		Type:        pluginpb.PluginType_PARSER,
	}, nil
}
