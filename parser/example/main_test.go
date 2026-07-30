package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	goplugin "github.com/hashicorp/go-plugin"
	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"google.golang.org/grpc"
)

// testClients bundles both service clients for in-process testing.
type testClients struct {
	Plugin pluginpb.PluginServiceClient
	Parser pluginpb.ParserServiceClient
}

type testPlugin struct {
	goplugin.NetRPCUnsupportedPlugin
}

func (p *testPlugin) GRPCServer(_ *goplugin.GRPCBroker, g *grpc.Server) error {
	svc := &exampleService{}
	pluginpb.RegisterPluginServiceServer(g, svc)
	pluginpb.RegisterParserServiceServer(g, svc)
	return nil
}

func (p *testPlugin) GRPCClient(_ context.Context, _ *goplugin.GRPCBroker, c *grpc.ClientConn) (any, error) {
	return &testClients{
		Plugin: pluginpb.NewPluginServiceClient(c),
		Parser: pluginpb.NewParserServiceClient(c),
	}, nil
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

func TestGetPluginInfo(t *testing.T) {
	clients := newTestClients(t)
	resp, err := clients.Plugin.GetPluginInfo(context.Background(), &pluginpb.GetPluginInfoRequest{})
	if err != nil {
		t.Fatalf("GetPluginInfo: %v", err)
	}
	if resp.Type != pluginpb.PluginType_PARSER {
		t.Errorf("Type = %v, want PARSER", resp.Type)
	}
	if resp.Name == "" {
		t.Error("Name must not be empty")
	}
}

func TestGetParserConfig(t *testing.T) {
	clients := newTestClients(t)
	resp, err := clients.Parser.GetParserConfig(context.Background(), &pluginpb.GetParserConfigRequest{})
	if err != nil {
		t.Fatalf("GetParserConfig: %v", err)
	}
	// identification_priority 0 is the recommended default; just confirm it round-trips.
	_ = resp.GetIdentificationPriority()
}

func TestIdentifyProjects(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "infra.example"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	clients := newTestClients(t)

	tests := []struct {
		name    string
		dir     string
		wantLen int
	}{
		{"finds .example files", dir, 1},
		{"empty dir returns empty response", t.TempDir(), 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := clients.Parser.IdentifyProjects(context.Background(), &pluginpb.IdentifyProjectsRequest{Directory: tc.dir})
			if err != nil {
				t.Fatalf("IdentifyProjects: %v", err)
			}
			if got := len(resp.GetFiles()); got != tc.wantLen {
				t.Errorf("files count = %d, want %d", got, tc.wantLen)
			}
		})
	}
}

func TestParse(t *testing.T) {
	clients := newTestClients(t)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid path returns tree", "/example/project.example", false},
		{"empty path returns error", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := clients.Parser.Parse(context.Background(), &pluginpb.ParseRequest{Path: tc.path})
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if resp.GetTree() == nil {
				t.Fatal("Parse returned nil tree")
			}
			if len(resp.GetTree().GetProviders()) == 0 {
				t.Error("Parse returned tree with no providers")
			}
		})
	}
}
