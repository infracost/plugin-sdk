package server

import (
	"github.com/infracost/parser/internal/plugin"
	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
)

type Server struct {
	pluginpb.UnimplementedPluginServiceServer
	pluginpb.UnimplementedParserServiceServer
}

var _ plugin.ParserPlugin = (*Server)(nil)

func New() *Server {
	return &Server{}
}
