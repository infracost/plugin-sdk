package server

import (
	"context"
	"testing"

	"github.com/infracost/config"
	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetParserConfig(t *testing.T) {
	t.Skip("template plugin — replace when implementing a real plugin")

	svc := newTestClient(t)

	resp, err := svc.GetParserConfig(context.Background(), &pluginpb.GetParserConfigRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	require.NotNil(t, resp.ConfigFileProjectType)
	assert.Equal(t, string(config.ProjectTypeCloudFormation), *resp.ConfigFileProjectType)
	assert.Equal(t, uint32(0), resp.IdentificationPriority)
}
