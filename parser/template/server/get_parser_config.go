package server

import (
	"context"

	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
)

func (s *Server) GetParserConfig(_ context.Context, _ *pluginpb.GetParserConfigRequest) (*pluginpb.GetParserConfigResponse, error) {
	return &pluginpb.GetParserConfigResponse{
		IdentificationPriority: 0,
	}, nil
}
