package pipeline

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestProcessPipeline(t *testing.T) {
	tests := []struct {
		name         string
		pipelineYAML string
		changedFiles []string
		expected     string
	}{
		{
			name: "Top-level steps",
			pipelineYAML: `
steps:
  - label: "Step 1"
    command: "echo 1"
    skip_if_unchanged:
      - "*.go"
  - label: "Step 2"
    command: "echo 2"
    skip_if_unchanged:
      - "*.js"
`,
			changedFiles: []string{"main.js"},
			expected: `
steps:
  - label: "Step 1"
    command: "echo 1"
    skip: true
  - label: "Step 2"
    command: "echo 2"
`,
		},
		{
			name: "Group with steps",
			pipelineYAML: `
steps:
  - group: "Group 1"
    steps:
      - label: "Step 1.1"
        command: "echo 1.1"
        skip_if_unchanged:
          - "*.go"
      - label: "Step 1.2"
        command: "echo 1.2"
        skip_if_unchanged:
          - "*.js"
  - label: "Step 2"
    command: "echo 2"
    skip_if_unchanged:
      - "*.py"
`,
			changedFiles: []string{"main.go"},
			expected: `
steps:
  - group: "Group 1"
    steps:
      - label: "Step 1.1"
        command: "echo 1.1"
      - label: "Step 1.2"
        command: "echo 1.2"
        skip: true
  - label: "Step 2"
    command: "echo 2"
    skip: true
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processedPipeline, err := ProcessPipeline([]byte(tt.pipelineYAML), tt.changedFiles)
			assert.NoError(t, err)

			var buf bytes.Buffer
			encoder := yaml.NewEncoder(&buf)
			encoder.SetIndent(2)
			err = encoder.Encode(processedPipeline)
			assert.NoError(t, err)

			assert.YAMLEq(t, tt.expected, buf.String())
		})
	}
}

func TestProcessPipelineErrors(t *testing.T) {
	tests := []struct {
		name         string
		pipelineYAML string
		changedFiles []string
		expectedErr  string
	}{
		{
			name:         "Invalid YAML",
			pipelineYAML: "invalid: [yaml",
			changedFiles: []string{},
			expectedErr:  "yaml: line 1: did not find expected ',' or ']'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ProcessPipeline([]byte(tt.pipelineYAML), tt.changedFiles)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
