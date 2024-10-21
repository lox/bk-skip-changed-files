package pipeline

import (
	"github.com/mattn/go-zglob"
	"gopkg.in/yaml.v3"
)

type Step struct {
	SkipIfUnchanged []string `yaml:"skip_if_unchanged,omitempty"`
}

func ProcessPipeline(pipelineData []byte, changedFiles []string) (*yaml.Node, error) {
	var pipeline yaml.Node
	err := yaml.Unmarshal(pipelineData, &pipeline)
	if err != nil {
		return nil, err
	}

	stepsNode := pipeline.Content[0].Content[1]
	err = processSteps(stepsNode, changedFiles)
	if err != nil {
		return nil, err
	}

	return &pipeline, nil
}

func processSteps(stepsNode *yaml.Node, changedFiles []string) error {
	for i := 0; i < len(stepsNode.Content); i++ {
		stepNode := stepsNode.Content[i]

		// Check if this is a group
		var group struct {
			Group string    `yaml:"group"`
			Steps yaml.Node `yaml:"steps"`
		}
		if err := stepNode.Decode(&group); err == nil && group.Group != "" {
			// Process steps within the group
			err := processSteps(&group.Steps, changedFiles)
			if err != nil {
				return err
			}
			continue
		}

		var s Step
		err := stepNode.Decode(&s)
		if err != nil {
			continue
		}

		if len(s.SkipIfUnchanged) > 0 {
			shouldSkip := true
			for _, file := range changedFiles {
				if matchesGlobs(file, s.SkipIfUnchanged) {
					shouldSkip = false
					break
				}
			}

			if shouldSkip {
				// Add skip field
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

	return nil
}

func matchesGlobs(file string, globs []string) bool {
	for _, glob := range globs {
		match, err := zglob.Match(glob, file)
		if err != nil {
			// Handle error (log it or continue)
			continue
		}
		if match {
			return true
		}
	}
	return false
}
