package git

import (
	"os"
	"os/exec"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetChangedFiles(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T)
		expectedFiles []string
		expectedError string
	}{
		{
			name: "Modified file",
			setupFunc: func(t *testing.T) {
				modifyAndStageFile(t, "file1.txt", "modified content")
				runGitCommand(t, "commit", "-m", "Modified file1.txt")
			},
			expectedFiles: []string{
				"file1.txt",
			},
		},
		{
			name: "Added file",
			setupFunc: func(t *testing.T) {
				createAndStageFile(t, "file4.txt", "new file")
				runGitCommand(t, "commit", "-m", "Added file4.txt")
			},
			expectedFiles: []string{
				"file4.txt",
			},
		},
		{
			name: "Removed file",
			setupFunc: func(t *testing.T) {
				runGitCommand(t, "rm", "file2.txt")
				runGitCommand(t, "commit", "-m", "Removed file2.txt")
			},
			expectedFiles: []string{
				"file2.txt",
			},
		},
		{
			name: "Renamed file",
			setupFunc: func(t *testing.T) {
				runGitCommand(t, "mv", "file3.txt", "file3_renamed.txt")
				runGitCommand(t, "commit", "-m", "Renamed file3.txt to file3_renamed.txt")
			},
			expectedFiles: []string{
				"file3_renamed.txt",
			},
		},
		{
			name: "Staged changes",
			setupFunc: func(t *testing.T) {
				modifyAndStageFile(t, "file1.txt", "modified content")
			},
			expectedError: "there are staged changes",
		},
		{
			name: "Symlink change",
			setupFunc: func(t *testing.T) {
				os.Symlink("file1.txt", "symlink.txt")
				runGitCommand(t, "add", "symlink.txt")
				runGitCommand(t, "commit", "-m", "Add symlink")
			},
			expectedFiles: []string{
				"symlink.txt",
			},
		},
		{
			name: "File mode change",
			setupFunc: func(t *testing.T) {
				os.Chmod("file1.txt", 0755)
				runGitCommand(t, "add", "file1.txt")
				runGitCommand(t, "commit", "-m", "Change file mode")
			},
			expectedFiles: []string{
				"file1.txt",
			},
		},
		{
			name: "Empty commit",
			setupFunc: func(t *testing.T) {
				runGitCommand(t, "commit", "--allow-empty", "-m", "Empty commit")
			},
			expectedFiles: []string{},
		},
		{
			name: "Whitespace-only change",
			setupFunc: func(t *testing.T) {
				modifyAndStageFile(t, "file1.txt", "initial content\n")
				runGitCommand(t, "commit", "-m", "Add newline")
			},
			expectedFiles: []string{
				"file1.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test repository
			tempDir := setupTestRepo(t)
			defer os.RemoveAll(tempDir)

			// Run the test-specific setup
			tt.setupFunc(t)

			// Get changed files
			changedFiles, err := GetChangedFiles("main")

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				sort.Strings(changedFiles)
				assert.ElementsMatch(t, tt.expectedFiles, changedFiles)
			}
		})
	}
}

func setupTestRepo(t *testing.T) string {
	// Create a temporary directory for the test repository
	tempDir, err := os.MkdirTemp("", "git-test")
	require.NoError(t, err)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) })

	// Initialize git repository
	runGitCommand(t, "init")
	runGitCommand(t, "config", "user.email", "test@example.com")
	runGitCommand(t, "config", "user.name", "Test User")
	runGitCommand(t, "config", "commit.gpgsign", "false")

	// Create and commit initial files
	createAndCommitFile(t, "file1.txt", "initial content")
	createAndCommitFile(t, "file2.txt", "initial content")
	createAndCommitFile(t, "file3.txt", "initial content")

	// Create main branch
	runGitCommand(t, "branch", "-M", "main")

	// Create and switch to a new branch
	runGitCommand(t, "checkout", "-b", "feature-branch")

	return tempDir
}

func runGitCommand(t *testing.T, args ...string) {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Git command failed: %s\nOutput: %s", cmd.Args, output)
}

func createAndCommitFile(t *testing.T, filename, content string) {
	createAndStageFile(t, filename, content)
	runGitCommand(t, "commit", "-m", "Add "+filename)
}

func createAndStageFile(t *testing.T, filename, content string) {
	err := os.WriteFile(filename, []byte(content), 0644)
	require.NoError(t, err)
	runGitCommand(t, "add", filename)
}

func modifyAndStageFile(t *testing.T, filename, content string) {
	err := os.WriteFile(filename, []byte(content), 0644)
	require.NoError(t, err)
	runGitCommand(t, "add", filename)
}

func runAndLogGitCommand(t *testing.T, args ...string) {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Git command failed: %s\nOutput: %s", cmd.Args, output)
	t.Logf("Git command '%s' output:\n%s", cmd.Args, output)
}
