package server

import (
	"context"
	"os"

	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
)

// projectMarkerFile is this template's placeholder identification signal: any
// directory containing a file with this name is claimed as a single project
// (the same pattern Terraform uses for *.tf files, one level up). Replace this
// with your format's real detection.
//
// Keep identification cheap: it runs for every directory of a repo, so reject
// files by extension and a byte scan for a distinctive marker before paying
// for a full decode.
//
// This demonstrates the directory:true branch of IdentifyProjectsResponse. If
// your format identifies individual files rather than whole directories (like
// CloudFormation or the file-based ../../example), populate Files instead —
// directory and files are mutually exclusive.
const projectMarkerFile = "template.config.json"

func (s *Server) IdentifyProjects(_ context.Context, req *pluginpb.IdentifyProjectsRequest) (*pluginpb.IdentifyProjectsResponse, error) {
	entries, err := os.ReadDir(req.GetDirectory())
	if err != nil {
		// Return an empty response rather than an error for paths we can't read.
		return &pluginpb.IdentifyProjectsResponse{}, nil
	}

	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() == projectMarkerFile {
			return &pluginpb.IdentifyProjectsResponse{Directory: true}, nil
		}
	}

	return &pluginpb.IdentifyProjectsResponse{}, nil
}
