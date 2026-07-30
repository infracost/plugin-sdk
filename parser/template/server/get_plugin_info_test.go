package server

import (
	"context"
	"testing"

	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPluginInfo(t *testing.T) {
	clients := newTestClients(t)

	resp, err := clients.Plugin.GetPluginInfo(context.Background(), &pluginpb.GetPluginInfoRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, pluginpb.PluginType_PARSER, resp.Type)
	assert.NotEmpty(t, resp.Name)
	assert.NotEmpty(t, resp.Description)
	assert.NotEmpty(t, resp.Url)
	assert.NotEmpty(t, resp.Author)
}
