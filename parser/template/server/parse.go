package server

import (
	"context"
	"encoding/json"
	"fmt"

	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"github.com/infracost/proto/gen/go/infracost/tree"

	"github.com/infracost/plugin-sdk/parser/template/options"
)

func (s *Server) Parse(_ context.Context, req *pluginpb.ParseRequest) (*pluginpb.ParseResponse, error) {
	if req.GetPath() == "" {
		return nil, fmt.Errorf("path is required")
	}

	// Decode this plugin's own options out of raw_options. raw_options is
	// always JSON (proto field 4, raw_options_format, is reserved and
	// dropped); a plugin that needs no options can ignore it entirely.
	var pluginOptions options.Options
	if len(req.GetRawOptions()) > 0 {
		if err := json.Unmarshal(req.GetRawOptions(), &pluginOptions); err != nil {
			return nil, fmt.Errorf("unmarshal raw_options: %w", err)
		}
	}

	// req.GenericOptions carries IaC-agnostic settings (working directory,
	// dependency requests, etc.) — see infracost/parser/options/options.proto.

	// TODO: replace this placeholder tree with resources parsed from req.Path.
	return &pluginpb.ParseResponse{
		Tree: &tree.Tree{
			Providers: map[string]*tree.Provider{
				"template": {
					Services: map[string]*tree.Service{
						"template": {
							Resources: []*tree.Resource{
								{
									Id:   req.GetPath(),
									Type: "template_resource",
								},
							},
						},
					},
				},
			},
		},
	}, nil
}
