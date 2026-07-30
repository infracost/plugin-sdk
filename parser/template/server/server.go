// Package server implements the gRPC services this plugin exposes:
// PluginService (identity) and ParserService (parsing). Each RPC lives in its
// own file below, mirroring the layout of the reference plugins in the
// infracost/parser repo.
package server

import (
	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
)

// Server implements both PluginService and ParserService. Embedding the
// Unimplemented servers keeps it forward-compatible if new RPCs are added; it
// is also what makes the not-implemented-here IdentifyEnvironments RPC
// correctly return codes.Unimplemented, a valid "no environment support"
// response (see README.md).
type Server struct {
	pluginpb.UnimplementedPluginServiceServer
	pluginpb.UnimplementedParserServiceServer
}

func New() *Server {
	return &Server{}
}
