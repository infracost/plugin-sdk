package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	optionspb "github.com/infracost/proto/gen/go/infracost/parser/options"
	pluginpb "github.com/infracost/proto/gen/go/infracost/plugin"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestParse(t *testing.T) {
	t.Skip("template plugin — replace when implementing a real plugin")

	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			dir := filepath.Join("testdata", entry.Name())
			req := loadRequest(t, dir)

			svc := newTestClient(t)

			resp, err := svc.Parse(context.Background(), req)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			actual, err := marshalStripped(resp)
			if err != nil {
				t.Fatalf("marshal actual: %v", err)
			}

			expectedPath := filepath.Join(dir, "expected.json")
			expectedRaw, err := os.ReadFile(expectedPath)
			if errors.Is(err, os.ErrNotExist) {
				if err := os.WriteFile(expectedPath, actual, 0o644); err != nil { //nolint:gosec // G306: test fixture
					t.Fatalf("write %s: %v", expectedPath, err)
				}
				t.Logf("wrote %s — review and re-run the test", expectedPath)
				return
			}
			if err != nil {
				t.Fatalf("read %s: %v", expectedPath, err)
			}

			expected, err := stripJSON(expectedRaw)
			if err != nil {
				t.Fatalf("strip expected %s: %v", expectedPath, err)
			}

			if diff := cmp.Diff(string(expected), string(actual)); diff != "" {
				t.Errorf("Parse mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func loadRequest(t *testing.T, dir string) *pluginpb.ParseRequest {
	t.Helper()

	var generic optionspb.GenericOptions
	if genericBytes, err := os.ReadFile(filepath.Join(dir, "generic.options.json")); err == nil {
		if err := protojson.Unmarshal(genericBytes, &generic); err != nil {
			t.Fatalf("unmarshal generic.options.json: %v", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("read generic.options.json: %v", err)
	}

	var rawOptions []byte
	if data, err := os.ReadFile(filepath.Join(dir, "cloudformation.options.json")); err == nil {
		rawOptions = data
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("read cloudformation.options.json: %v", err)
	}

	return &pluginpb.ParseRequest{
		Path:             filepath.Join(dir, "template.yml"),
		GenericOptions:   &generic,
		RawOptions:       rawOptions,
		RawOptionsFormat: "application/json",
	}
}

// marshalStripped marshals resp to JSON and strips keys with empty-object
// values so the fixture stays small and stable when new tree services/fields
// are added with no data.
func marshalStripped(resp *pluginpb.ParseResponse) ([]byte, error) {
	data, err := protojson.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return stripJSON(data)
}

func stripJSON(data []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	v = stripEmpty(v)
	out, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var indented bytes.Buffer
	if err := json.Indent(&indented, out, "", "  "); err != nil {
		return nil, err
	}
	indented.WriteByte('\n')
	return indented.Bytes(), nil
}

// stripEmpty recursively removes object keys whose values are empty
// map[string]any{} (after recursion). Arrays are left unchanged structurally.
func stripEmpty(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			val = stripEmpty(val)
			if m, ok := val.(map[string]any); ok && len(m) == 0 {
				continue
			}
			out[k] = val
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, val := range x {
			out[i] = stripEmpty(val)
		}
		return out
	default:
		return v
	}
}
