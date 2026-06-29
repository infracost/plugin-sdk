// infracost-provider-plugin-example is a minimal provider plugin that
// demonstrates the full plugin interface contract. It walks the cost tree
// produced by a parser plugin and prices every "aws_instance" resource with a
// hardcoded hourly rate. Use this as a starting point for new providers.
package main

import (
	"context"
	"fmt"
	"math/big"

	goplugin "github.com/hashicorp/go-plugin"
	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"github.com/infracost/proto/gen/go/infracost/provider"
	"github.com/infracost/proto/gen/go/infracost/rational"
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
	pluginpb.RegisterProviderServiceServer(g, svc)
	return nil
}

func (p *examplePlugin) GRPCClient(_ context.Context, _ *goplugin.GRPCBroker, _ *grpc.ClientConn) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

// exampleService implements both PluginService and ProviderService. Embedding
// the Unimplemented servers keeps it forward-compatible if new RPCs are added.
type exampleService struct {
	pluginpb.UnimplementedPluginServiceServer
	pluginpb.UnimplementedProviderServiceServer
}

// GetPluginInfo (PluginService) reports the plugin's identity and type. The CLI
// calls this first to decide how to use the binary.
func (s *exampleService) GetPluginInfo(_ context.Context, _ *pluginpb.GetPluginInfoRequest) (*pluginpb.GetPluginInfoResponse, error) {
	return &pluginpb.GetPluginInfoResponse{
		Type:        pluginpb.PluginType_PROVIDER,
		Name:        "example/example",
		Version:     "0.1.0",
		Description: "Prices aws_instance resources with hardcoded rates",
		Url:         "https://github.com/acme/infracost-provider-plugin-example",
		Author:      "Acme",
	}, nil
}

// Process (ProviderService) receives the IaC-agnostic cost tree and returns
// priced resources. The tree is organised as providers -> services ->
// resources.
func (s *exampleService) Process(_ context.Context, req *pluginpb.ProcessRequest) (*pluginpb.ProcessResponse, error) {
	in := req.GetInput()
	if in == nil {
		return &pluginpb.ProcessResponse{Output: &provider.Output{}}, nil
	}

	var resources []*provider.Resource
	for _, prov := range in.GetTree().GetProviders() {
		for _, svc := range prov.GetServices() {
			for _, res := range svc.GetResources() {
				if res.GetType() != "aws_instance" {
					continue
				}
				resources = append(resources, &provider.Resource{
					Id:          res.GetId(),
					Type:        res.GetType(),
					Name:        res.GetId(),
					Region:      res.GetRegion(),
					IsSupported: true,
					Costs: &provider.ResourceCosts{
						Components: []*provider.CostComponent{
							{
								Name:              "Compute (t3.micro, on-demand)",
								Unit:              "hours",
								PriceWasHardcoded: true,
								PeriodPrice: &provider.PeriodPrice{
									Price:  rat(116, 10000),
									Period: provider.Period_HOUR,
								},
								Quantity: rat(730, 1),
							},
						},
					},
				})
			}
		}
	}

	return &pluginpb.ProcessResponse{
		Output: &provider.Output{Resources: resources},
	}, nil
}

// ListFinopsPolicies (ProviderService) returns the FinOps policies this
// provider can evaluate. This example evaluates none.
func (s *exampleService) ListFinopsPolicies(_ context.Context, _ *pluginpb.ListFinopsPoliciesRequest) (*pluginpb.ListFinopsPoliciesResponse, error) {
	return &pluginpb.ListFinopsPoliciesResponse{}, nil
}

func rat(num, denom int64) *rational.Rat {
	return &rational.Rat{
		Numerator:   big.NewInt(num).Bytes(),
		Denominator: big.NewInt(denom).Bytes(),
	}
}
