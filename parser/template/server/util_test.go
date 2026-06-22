package server

import (
	"context"
	"testing"

	goplugin "github.com/hashicorp/go-plugin"
	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"google.golang.org/grpc"
)

type testClients struct {
	Plugin pluginpb.PluginServiceClient
	Parser pluginpb.ParserServiceClient
}

func newTestClients(t *testing.T) *testClients {
	t.Helper()

	client, _ := goplugin.TestPluginGRPCConn(t, true, map[string]goplugin.Plugin{
		"plugin": &testPlugin{},
	})
	t.Cleanup(func() { _ = client.Close() })

	raw, err := client.Dispense("plugin")
	if err != nil {
		t.Fatalf("dispense: %v", err)
	}
	return raw.(*testClients)
}

func newTestClient(t *testing.T) pluginpb.ParserServiceClient {
	return newTestClients(t).Parser
}

type testPlugin struct {
	goplugin.NetRPCUnsupportedPlugin
}

func (p *testPlugin) GRPCServer(_ *goplugin.GRPCBroker, g *grpc.Server) error {
	s := New()
	pluginpb.RegisterPluginServiceServer(g, s)
	pluginpb.RegisterParserServiceServer(g, s)
	return nil
}

func (p *testPlugin) GRPCClient(_ context.Context, _ *goplugin.GRPCBroker, c *grpc.ClientConn) (any, error) {
	return &testClients{
		Plugin: pluginpb.NewPluginServiceClient(c),
		Parser: pluginpb.NewParserServiceClient(c),
	}, nil
}
