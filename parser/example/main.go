// infracost-parser-plugin-example is a minimal parser plugin that demonstrates
// the full interface contract. It detects ".example" files and returns a single
// dummy resource for each one. Use this as a starting point for new plugins.
package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-plugin"
	"github.com/infracost/proto/gen/go/infracost/parser/api"
	"github.com/infracost/proto/gen/go/infracost/parser/cloudformation"
	"github.com/infracost/proto/gen/go/infracost/parser/options"
	"google.golang.org/grpc"
)

const maxMessageSize = 64 * 1024 * 1024

var (
	_ plugin.Plugin     = (*exampleParser)(nil)
	_ plugin.GRPCPlugin = (*exampleParser)(nil)
)

type exampleParser struct {
	plugin.NetRPCUnsupportedPlugin
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "INFRACOST_PARSER_PLUGIN_MAGIC_COOKIE",
			MagicCookieValue: "ac92b06c592f",
		},
		Plugins: map[string]plugin.Plugin{
			"parser": new(exampleParser),
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

func (p *exampleParser) GRPCServer(_ *plugin.GRPCBroker, g *grpc.Server) error {
	api.RegisterParserServiceServer(g, &exampleService{})
	return nil
}

func (p *exampleParser) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, _ *grpc.ClientConn) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

// exampleService implements the ParserService gRPC server.
type exampleService struct {
	api.UnimplementedParserServiceServer
}

func (s *exampleService) Describe(_ context.Context, _ *api.DescribeRequest) (*api.DescribeResponse, error) {
	return &api.DescribeResponse{
		Name:                "plugins.infracost.io/example/example",
		DisplayName:         "Example Format",
		Priority:            50,
		FileExtensions:      []string{".example"},
		SupportsDirectories: false,
	}, nil
}

func (s *exampleService) Detect(_ context.Context, req *api.DetectRequest) (*api.DetectResponse, error) {
	path := req.GetPath()
	if path == "" {
		return &api.DetectResponse{Detected: false}, nil
	}

	if len(path) > 8 && path[len(path)-8:] == ".example" {
		return &api.DetectResponse{
			Detected:    true,
			ProjectType: "example",
			Confidence:  api.DetectConfidence_DETECT_CONFIDENCE_HIGH,
		}, nil
	}

	return &api.DetectResponse{Detected: false}, nil
}

func (s *exampleService) Initialize(_ context.Context, _ *api.InitializeRequest) (*api.InitializeResponse, error) {
	return &api.InitializeResponse{}, nil
}

func (s *exampleService) Parse(_ context.Context, req *api.ParseRequest) (*api.ParseResponse, error) {
	if req == nil || req.Target == nil {
		return nil, fmt.Errorf("request and target cannot be nil")
	}

	// This example returns a CloudFormation-shaped result with a single
	// dummy resource. A real plugin would define its own target/result
	// proto messages and parse actual IaC files here.
	return &api.ParseResponse{
		Result: &api.ParseResponseResult{
			Value: &api.ParseResponseResult_Cloudformation{
				Cloudformation: &cloudformation.Result{
					Resources: map[string]*cloudformation.Resource{
						"ExampleResource": {
							Id:        "ExampleResource",
							Type:      "Example::Service::Resource",
							Supported: false,
						},
					},
				},
			},
		},
	}, nil
}

func (s *exampleService) ParseToTree(_ context.Context, req *api.ParseToTreeRequest) (*api.ParseToTreeResponse, error) {
	if req == nil || req.Target == nil {
		return nil, fmt.Errorf("request and target cannot be nil")
	}

	return &api.ParseToTreeResponse{}, nil
}

// Ensure GenericOptions is referenced so go mod tidy keeps it.
var _ = (*options.GenericOptions)(nil)
