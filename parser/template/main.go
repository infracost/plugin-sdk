// infracost-parser-template is the production-shaped starting point for a new
// parser plugin: main.go stays a thin entrypoint, and the RPC implementations
// live one-per-file in server/. Copy this directory, rename the module, and
// replace the placeholder logic in server/ with your format's real behaviour.
package main

import (
	"context"
	"fmt"

	goplugin "github.com/hashicorp/go-plugin"
	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"google.golang.org/grpc"

	"github.com/infracost/plugin-sdk/parser/template/server"
)

const maxMessageSize = 64 * 1024 * 1024

// handshake is shared by every Infracost plugin, parser and provider alike.
// The CLI identifies the plugin type at runtime via the GetPluginInfo RPC, not
// from the handshake or the binary name, so there is a single magic cookie.
var handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "INFRACOST_PLUGIN",
	MagicCookieValue: "de8c7e96-497c-4168-80c4-fc875c8ce764",
}

func main() {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]goplugin.Plugin{
			// The dispense key is always "plugin".
			"plugin": &templatePlugin{},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			opts = append(opts,
				grpc.MaxRecvMsgSize(maxMessageSize),
				grpc.MaxSendMsgSize(maxMessageSize),
			)
			return grpc.NewServer(opts...)
		},
	})
}

var (
	_ goplugin.Plugin     = (*templatePlugin)(nil)
	_ goplugin.GRPCPlugin = (*templatePlugin)(nil)
)

type templatePlugin struct {
	goplugin.NetRPCUnsupportedPlugin
}

func (p *templatePlugin) GRPCServer(_ *goplugin.GRPCBroker, g *grpc.Server) error {
	s := server.New()
	// Every plugin implements PluginService (identity) plus one of
	// ParserService / ProviderService. Register both on the same server.
	pluginpb.RegisterPluginServiceServer(g, s)
	pluginpb.RegisterParserServiceServer(g, s)
	return nil
}

func (p *templatePlugin) GRPCClient(_ context.Context, _ *goplugin.GRPCBroker, _ *grpc.ClientConn) (any, error) {
	return nil, fmt.Errorf("not implemented")
}
