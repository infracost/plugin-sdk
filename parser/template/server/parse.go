package server

import (
	"context"
	"encoding/json"

	"github.com/infracost/go-proto/pkg/tree"
	"github.com/infracost/parser/pkg/diagnostic"
	"github.com/infracost/parser/plugin/template/options"
	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
)

func (s *Server) Parse(_ context.Context, req *pluginpb.ParseRequest) (*pluginpb.ParseResponse, error) {
	// Decode this plugin's own options out of raw_options. The CLI sends them as
	// JSON (raw_options_format = "application/json"); a plugin that needs no
	// options can ignore them. Document your raw_options shape so callers know
	// what to send.
	var pluginOptions options.Options
	if len(req.RawOptions) > 2 {
		_ = json.Unmarshal(req.RawOptions, &pluginOptions)
	}

	var parsedTree tree.Tree
	var diags *diagnostic.Diagnostics

	// TODO: parse req into a tree.Tree

	protoTree, err := parsedTree.ToProto()
	if err != nil {
		return nil, err
	}

	return &pluginpb.ParseResponse{
		Diagnostics: diags.ToProto(),
		Tree:        protoTree,
	}, nil
}
