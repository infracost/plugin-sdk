package server

import (
	"context"
	"testing"

	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetParserConfig(t *testing.T) {
	svc := newTestClient(t)

	resp, err := svc.GetParserConfig(context.Background(), &pluginpb.GetParserConfigRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, uint32(0), resp.GetIdentificationPriority())
	// ConfigFileProjectType is left unset here, so it defaults to the plugin
	// name (see get_parser_config.go). Set it explicitly if your format should
	// map onto an existing infracost/config project type instead.
	assert.Nil(t, resp.ConfigFileProjectType)
}
