// infracost-parser-plugin-example is a minimal parser plugin that demonstrates
// the full plugin interface contract. It identifies ".example" files in a
// directory and returns an empty cost tree for each one. Use this as a starting
// point for new parser plugins.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	goplugin "github.com/hashicorp/go-plugin"
	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"github.com/infracost/proto/gen/go/infracost/tree"
	"google.golang.org/grpc"
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

var (
	_ goplugin.Plugin     = (*examplePlugin)(nil)
	_ goplugin.GRPCPlugin = (*examplePlugin)(nil)
)

type examplePlugin struct {
	goplugin.NetRPCUnsupportedPlugin
}

func main() {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]goplugin.Plugin{
			// The dispense key is always "plugin".
			"plugin": &examplePlugin{},
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

func (p *examplePlugin) GRPCServer(_ *goplugin.GRPCBroker, g *grpc.Server) error {
	svc := &exampleService{}
	// Every plugin implements PluginService (identity) plus one of
	// ParserService / ProviderService. Register both on the same server.
	pluginpb.RegisterPluginServiceServer(g, svc)
	pluginpb.RegisterParserServiceServer(g, svc)
	return nil
}

func (p *examplePlugin) GRPCClient(_ context.Context, _ *goplugin.GRPCBroker, _ *grpc.ClientConn) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

// exampleService implements both PluginService and ParserService. Embedding the
// Unimplemented servers keeps it forward-compatible if new RPCs are added.
type exampleService struct {
	pluginpb.UnimplementedPluginServiceServer
	pluginpb.UnimplementedParserServiceServer
}

// GetPluginInfo (PluginService) reports the plugin's identity and type. The CLI
// calls this first to decide how to use the binary.
func (s *exampleService) GetPluginInfo(_ context.Context, _ *pluginpb.GetPluginInfoRequest) (*pluginpb.GetPluginInfoResponse, error) {
	return &pluginpb.GetPluginInfoResponse{
		Type:        pluginpb.PluginType_PARSER,
		Name:        "example/example",
		Version:     "0.1.0",
		Description: "Parses .example files",
		Url:         "https://github.com/acme/infracost-parser-plugin-example",
		Author:      "Acme",
	}, nil
}

// GetParserConfig (ParserService) tells the CLI how this plugin participates in
// project identification.
func (s *exampleService) GetParserConfig(_ context.Context, _ *pluginpb.GetParserConfigRequest) (*pluginpb.GetParserConfigResponse, error) {
	return &pluginpb.GetParserConfigResponse{
		// 0 is the recommended default. Plugins with a higher priority are
		// offered a directory before lower-priority plugins.
		IdentificationPriority: 0,
		// ConfigFileProjectType is left unset, so it defaults to the plugin
		// name. Set it to map onto an existing infracost/config project type.
	}, nil
}

// IdentifyProjects (ParserService) inspects a single directory (it must NOT
// recurse) and reports which paths this plugin can parse.
func (s *exampleService) IdentifyProjects(_ context.Context, req *pluginpb.IdentifyProjectsRequest) (*pluginpb.IdentifyProjectsResponse, error) {
	entries, err := os.ReadDir(req.GetDirectory())
	if err != nil {
		// Return an empty response rather than an error for paths we can't read.
		return &pluginpb.IdentifyProjectsResponse{}, nil
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".example") {
			// Each matching file is its own project. Return paths relative to
			// the directory. To claim the whole directory as a single project
			// instead, set Directory: true and leave Files empty.
			files = append(files, entry.Name())
		}
	}

	return &pluginpb.IdentifyProjectsResponse{Files: files}, nil
}

// Parse (ParserService) reads the IaC at req.Path and returns an IaC-agnostic
// cost tree. A real plugin would build the tree from the parsed resources; this
// example returns an empty tree.
func (s *exampleService) Parse(_ context.Context, req *pluginpb.ParseRequest) (*pluginpb.ParseResponse, error) {
	if req.GetPath() == "" {
		return nil, fmt.Errorf("path is required")
	}

	// req.GenericOptions carries IaC-agnostic settings (working directory,
	// dependency requests, etc.). Plugin-specific options arrive as raw bytes
	// in req.RawOptions, encoded as req.RawOptionsFormat (e.g. "application/json").

	return &pluginpb.ParseResponse{
		Tree: &tree.Tree{},
	}, nil
}
