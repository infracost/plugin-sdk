package server

import (
	"context"
	"testing"

	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentifyProjects(t *testing.T) {
	svc := newTestClient(t)

	tests := []struct {
		name          string
		dir           string
		wantDirectory bool
	}{
		{"directory containing the marker file is claimed whole", "testdata/basic", true},
		{"directory without the marker file returns an empty response", "testdata", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.IdentifyProjects(context.Background(), &pluginpb.IdentifyProjectsRequest{
				Directory: tc.dir,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.wantDirectory, resp.GetDirectory())
			assert.Empty(t, resp.GetFiles(), "directory and files are mutually exclusive")
		})
	}
}
