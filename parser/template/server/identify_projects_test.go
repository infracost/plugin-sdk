package server

import (
	"context"
	"testing"

	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentifyProjects_Basic(t *testing.T) {
	t.Skip("template plugin — replace when implementing a real plugin")

	svc := newTestClient(t)

	resp, err := svc.IdentifyProjects(context.Background(), &pluginpb.IdentifyProjectsRequest{
		Directory: "testdata/basic",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.False(t, resp.Directory, "expected testdata/basic not to be identified as a cloudformation project directory")
	assert.Len(t, resp.Files, 1)
	assert.Contains(t, resp.Files, "template.yml")
}
