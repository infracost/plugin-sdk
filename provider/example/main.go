// infracost-provider-plugin-example is a minimal provider plugin that
// demonstrates the full interface contract. It prices "aws_instance" resources
// with a hardcoded hourly rate. Use this as a starting point for new providers.
package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/go-plugin"
	parserapi "github.com/infracost/proto/gen/go/infracost/parser/api"
	"github.com/infracost/proto/gen/go/infracost/provider"
	"github.com/infracost/proto/gen/go/infracost/rational"
	"google.golang.org/grpc"
)

const maxMessageSize = 64 * 1024 * 1024

var (
	_ plugin.Plugin     = (*exampleProvider)(nil)
	_ plugin.GRPCPlugin = (*exampleProvider)(nil)
)

type exampleProvider struct {
	plugin.NetRPCUnsupportedPlugin
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "INFRACOST_PROVIDER_PLUGIN_MAGIC_COOKIE",
			MagicCookieValue: "04d179d767fc",
		},
		Plugins: map[string]plugin.Plugin{
			"provider": new(exampleProvider),
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

func (p *exampleProvider) GRPCServer(_ *plugin.GRPCBroker, g *grpc.Server) error {
	provider.RegisterProviderServiceServer(g, &exampleService{})
	return nil
}

func (p *exampleProvider) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, _ *grpc.ClientConn) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

type exampleService struct {
	provider.UnimplementedProviderServiceServer
}

func (s *exampleService) Describe(_ context.Context, _ *provider.DescribeRequest) (*provider.DescribeResponse, error) {
	return &provider.DescribeResponse{
		Name:        "plugins.infracost.io/example/example",
		DisplayName: "Example Provider",
	}, nil
}

func (s *exampleService) ListSupportedResources(_ context.Context, _ *provider.ListSupportedResourcesRequest) (*provider.ListSupportedResourcesResponse, error) {
	return &provider.ListSupportedResourcesResponse{
		Terraform: &parserapi.SupportedResources{
			ResourceTypes: []*parserapi.SupportedResource{
				{ResourceType: "aws_instance"},
			},
		},
	}, nil
}

func (s *exampleService) Process(_ context.Context, req *provider.ProcessRequest) (*provider.ProcessResponse, error) {
	if req == nil || req.Input == nil {
		return &provider.ProcessResponse{Output: &provider.Output{}}, nil
	}

	return &provider.ProcessResponse{
		Output: &provider.Output{
			Resources: []*provider.Resource{
				{
					Id:          "example-instance",
					Type:        "aws_instance",
					Name:        "example",
					Region:      "us-east-1",
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
				},
			},
		},
	}, nil
}

func (s *exampleService) ProcessTree(_ context.Context, _ *provider.ProcessTreeRequest) (*provider.ProcessTreeResponse, error) {
	return &provider.ProcessTreeResponse{
		Output: &provider.Output{},
	}, nil
}

func (s *exampleService) ListFinopsPolicies(_ context.Context, _ *provider.ListFinopsPoliciesRequest) (*provider.ListFinopsPoliciesResponse, error) {
	return &provider.ListFinopsPoliciesResponse{}, nil
}

func rat(num, denom int64) *rational.Rat {
	return &rational.Rat{
		Numerator:   big.NewInt(num).Bytes(),
		Denominator: big.NewInt(denom).Bytes(),
	}
}
