package git

import (
	"errors"
	"os/exec"
	"strings"
)

var ErrStagedChanges = errors.New("there are staged changes, please commit or stash them")

// GetChangedFiles returns a list of files that have changed since the branch point
func GetChangedFiles(baseBranch string) ([]string, error) {
	// Check for staged changes
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return nil, err
	}
	if len(statusOutput) > 0 {
		return nil, ErrStagedChanges
	}

	// Find the merge base
	mergeBaseCmd := exec.Command("git", "merge-base", baseBranch, "HEAD")
	mergeBase, err := mergeBaseCmd.Output()
	if err != nil {
		return nil, err
	}

	// Get changed files
	cmd := exec.Command("git", "diff", "--name-status", "--diff-filter=ACMRD", "--find-renames", strings.TrimSpace(string(mergeBase)), "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	var files []string
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			file := parts[len(parts)-1]
			files = append(files, file)
		}
	}

	return files, nil
}
