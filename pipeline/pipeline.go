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
	for i := 0; i < len(stepsNode.Content); i++ {
		stepNode := stepsNode.Content[i]

		var s Step
		err := stepNode.Decode(&s)
		if err != nil {
			continue
		}

		if len(s.SkipIfUnchanged) > 0 {
			shouldSkip := true
			for _, glob := range s.SkipIfUnchanged {
				for _, file := range changedFiles {
					match, err := zglob.Match(glob, file)
					if err != nil {
						return nil, err
					}
					if match {
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

	return &pipeline, nil
}
