package server

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"gopkg.in/yaml.v3"
)

func (s *Server) IdentifyProjects(_ context.Context, req *pluginpb.IdentifyProjectsRequest) (*pluginpb.IdentifyProjectsResponse, error) {
	entries, err := os.ReadDir(req.Directory)
	if err != nil {
		return &pluginpb.IdentifyProjectsResponse{}, nil
	}

	var files []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(req.Directory, entry.Name())
		if identifyCloudFormationFile(path) {
			files = append(files, entry.Name())
		}
	}

	return &pluginpb.IdentifyProjectsResponse{
		Files: files,
	}, nil
}

func identifyCloudFormationFile(path string) bool {
	if isOfCDKOrigin(path) {
		return false
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return identifyCloudFormationJSON(path)
	case ".yml", ".yaml":
		return identifyCloudFormationYAML(path)
	default:
		return false
	}
}

type identificationSniff struct {
	Service   string `yaml:"service"`
	Resources map[string]struct {
		Type string `json:"Type" yaml:"Type"`
	} `json:"Resources" yaml:"Resources"`
	AWSTemplateFormatVersion string `json:"AWSTemplateFormatVersion" yaml:"AWSTemplateFormatVersion"`
}

func (sniff *identificationSniff) IsCF() bool {
	if sniff.AWSTemplateFormatVersion != "" {
		return true
	}
	for _, v := range sniff.Resources {
		if !strings.HasPrefix(v.Type, "::") && strings.Contains(v.Type, "::") {
			return true
		}
	}
	return false
}

func identifyCloudFormationJSON(path string) bool {
	// #nosec G304
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var sniff identificationSniff
	if err := json.Unmarshal(content, &sniff); err != nil {
		return false
	}
	return sniff.IsCF()
}

func identifyCloudFormationYAML(path string) bool {
	// #nosec G304
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var sniff identificationSniff
	if err := yaml.Unmarshal(content, &sniff); err != nil {
		return false
	}

	return sniff.IsCF()
}

// isOfCDKOrigin returns true if the given path is a CDK sample, or synthesized template
// CDK projects are added elsewhere, so ignore them when looking for native CFN
func isOfCDKOrigin(path string) bool {
	return strings.Contains(path, "node_modules") || strings.Contains(path, "infracost.cdk.out")
}
