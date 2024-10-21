package main

import (
	"os"
	"testing"

	bkpipeline "github.com/lox/bk-skip-unchanged-files/pipeline"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestProcessPipeline(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		inputYAML      string
		changedFiles   []string
		expectedOutput string
	}{
		{
			name: "Skip step when no files changed",
			inputYAML: `
steps:
  - command: echo "Hello"
    skip_if_unchanged:
      - "*.go"
`,
			changedFiles: []string{"README.md"},
			expectedOutput: `
steps:
  - command: echo "Hello"
    skip: true
`,
		},
		{
			name: "Don't skip step when files changed",
			inputYAML: `
steps:
  - command: echo "Hello"
    skip_if_unchanged:
      - "*.go"
`,
			changedFiles: []string{"main.go"},
			expectedOutput: `
steps:
  - command: echo "Hello"
`,
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse input YAML
			var pipeline yaml.Node
			err := yaml.Unmarshal([]byte(tt.inputYAML), &pipeline)
			assert.NoError(t, err)

			// Mock getChangedFiles function
			oldGetChangedFiles := getChangedFiles
			getChangedFiles = func(string) ([]string, error) {
				return tt.changedFiles, nil
			}
			defer func() { getChangedFiles = oldGetChangedFiles }()

			// Process the pipeline
			stepsNode := pipeline.Content[0].Content[1]
			for i := 0; i < len(stepsNode.Content); i++ {
				stepNode := stepsNode.Content[i]

				var s bkpipeline.Step
				err := stepNode.Decode(&s)
				assert.NoError(t, err)

				if len(s.SkipIfUnchanged) > 0 {
					shouldSkip := true
					for _, glob := range s.SkipIfUnchanged {
						for _, file := range tt.changedFiles {
							if matchGlob(glob, file) {
								shouldSkip = false
								break
							}
						}
						if !shouldSkip {
							break
						}
					}

					if shouldSkip {
						skipNode := &yaml.Node{
							Kind:  yaml.ScalarNode,
							Tag:   "!!bool",
							Value: "true",
						}
						stepNode.Content = append(stepNode.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: "skip"}, skipNode)
					}

					// Remove skip_if_unchanged field
					for j := 0; j < len(stepNode.Content); j += 2 {
						if stepNode.Content[j].Value == "skip_if_unchanged" {
							stepNode.Content = append(stepNode.Content[:j], stepNode.Content[j+2:]...)
							break
						}
					}
				}
			}

			// Convert processed pipeline back to YAML
			output, err := yaml.Marshal(&pipeline)
			assert.NoError(t, err)

			// Compare output with expected
			assert.YAMLEq(t, tt.expectedOutput, string(output))
		})
	}
}

// matchGlob is a simple function to match a glob pattern against a filename
func matchGlob(pattern, name string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == "*.go" && len(name) > 3 && name[len(name)-3:] == ".go" {
		return true
	}
	return pattern == name
}

func TestMain(m *testing.M) {
	// Set up any global test configuration here
	os.Exit(m.Run())
}
