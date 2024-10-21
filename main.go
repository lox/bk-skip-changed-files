package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/alecthomas/kong"
	"github.com/lox/bk-skip-unchanged-files/git"
	"github.com/lox/bk-skip-unchanged-files/pipeline"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var CLI struct {
	PipelineFile string `arg:"" help:"Path to the pipeline.yml file" default:"pipeline.yml"`
	BaseBranch   string `help:"The base branch to compare against" env:"BUILDKITE_PULL_REQUEST_BASE_BRANCH" default:"origin/main"`
	Upload       bool   `help:"Upload the modified pipeline.yml to Buildkite"`
}

func init() {
	// Configure zerolog
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
}

var getChangedFiles = git.GetChangedFiles

func main() {
	ctx := kong.Parse(&CLI)

	// Read the pipeline file
	data, err := os.ReadFile(CLI.PipelineFile)
	if err != nil {
		log.Fatal().Err(err).Str("file", CLI.PipelineFile).Msg("Failed to read pipeline file")
	}
	log.Info().Str("file", CLI.PipelineFile).Msg("Pipeline file read successfully")

	changedFiles, err := getChangedFiles(CLI.BaseBranch)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	// Process the pipeline
	processedPipeline, err := pipeline.ProcessPipeline(data, changedFiles)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	// Write the modified pipeline back to YAML
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	err = encoder.Encode(processedPipeline)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}
	modifiedData := buf.Bytes()

	if CLI.Upload {
		// Upload to Buildkite
		cmd := exec.Command("buildkite-agent", "pipeline", "upload")
		cmd.Stdin = bytes.NewReader(modifiedData)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to upload pipeline to Buildkite")
		}
		log.Info().Msg("Pipeline uploaded to Buildkite successfully")
	} else {
		// Write to stdout
		fmt.Println(string(modifiedData))
		log.Info().Msg("Modified pipeline written to stdout")
	}
}
