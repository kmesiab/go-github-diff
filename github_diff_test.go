package github

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGithub_ParseGithubPullRequestURL(t *testing.T) {
	pullRequestURL, err := ParsePullRequestURL(
		"https://github.com/google/go-github/pull/1234",
	)

	require.NoError(t, err)

	if pullRequestURL.Owner != "google" || pullRequestURL.Repo != "go-github" || pullRequestURL.PRNumber != 1234 {
		t.Error("failed to parse pull request URL")
	}
}

func TestGithub_ParseGithubPullRequestInvalidURL(t *testing.T) {
	pullRequestURL, err := ParsePullRequestURL("foo")

	require.Error(t, err)
	require.Nil(t, pullRequestURL)
}

func TestGetFileExtensions(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "File in root directory",
			path:     "/myfile.txt",
			expected: ".txt",
		},
		{
			name:     "File in nested directory",
			path:     "/path/to/myfile.mp3",
			expected: ".mp3",
		},
		{
			name:     "File with no extension",
			path:     "/path/to/myfile",
			expected: "",
		},
		{
			name:     "File with dot in name",
			path:     "/path.to/my.file.txt",
			expected: ".txt",
		},
		{
			name:     "Empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "Dot file",
			path:     "/path/to/.myfile",
			expected: ".myfile",
		},
		{
			name:     "Path with no file",
			path:     "/path/to/",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFileExtension(tt.path)
			if result != tt.expected {
				t.Errorf("getFileExtension(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestSplitDiffIntoFiles(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
index acdee69..e522a2d 100644
--- a/main.go
+++ b/main.go
@@ -115,6 +115,7 @@ func onAnalyzeButtonClickedHandler() {
+	// Additional line of code
}

diff --git a/ui/critic_window.go b/ui/critic_window.go
index d9e3436..96c7eb8 100644
--- a/ui/critic_window.go
+++ b/ui/critic_window.go
@@ -7,13 +7,14 @@ type CriticWindow struct {
+	// Another line of code
}`

	expected := []string{
		`diff --git a/main.go b/main.go
index acdee69..e522a2d 100644
--- a/main.go
+++ b/main.go
@@ -115,6 +115,7 @@ func onAnalyzeButtonClickedHandler() {
+	// Additional line of code
}`,
		`diff --git a/ui/critic_window.go b/ui/critic_window.go
index d9e3436..96c7eb8 100644
--- a/ui/critic_window.go
+++ b/ui/critic_window.go
@@ -7,13 +7,14 @@ type CriticWindow struct {
+	// Another line of code
}`,
	}

	files := splitDiffIntoFiles(diff)

	if !reflect.DeepEqual(files, expected) {
		t.Errorf("splitDiffIntoFiles() = %v, want %v", files, expected)
	}
}

func TestSplitDiffIntoFilesNoFiles(t *testing.T) {
	diff := ``
	files := splitDiffIntoFiles(diff)

	if len(files) != 0 {
		t.Error("Expected 0 files, got files")
	}
}

func TestSplitDiffIntoFilesOneFile(t *testing.T) {
	diff := `--git a/file.txt b/file.txt`

	files := splitDiffIntoFiles(diff)

	if len(files) != 1 {
		t.Error("Expected 1 file, got ", len(files))
	}
}

func TestSplitDiffIntoFilesLastEmpty(t *testing.T) {
	diff := `diff --git a/f1 b/f1
diff --git c/f2 d/f2
`

	expected := []string{
		"diff --git a/f1 b/f1",
		"diff --git c/f2 d/f2",
	}

	files := splitDiffIntoFiles(diff)

	if len(files) != len(expected) {
		t.Errorf("Expected %d files, got %d files", len(expected), len(files))
	}

	for i := range files {
		if files[i] != expected[i] {
			t.Errorf("File %d does not match: %v vs %v", i, files[i], expected[i])
		}
	}
}

func TestParseGitDiffFileString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *GitDiff
		wantErr error
	}{
		{
			name: "Valid Git Diff",
			input: `diff --git a/file1.go b/file1.go
index 123abc..456def 100644
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
+import "fmt"`,
			want: &GitDiff{
				FilePathOld:  "a/file1.go",
				FilePathNew:  "b/file1.go",
				Index:        "123abc..456def 100644",
				DiffContents: "--- a/file1.go\n+++ b/file1.go\n@@ -1,3 +1,4 @@\n+import \"fmt\"",
			},
			wantErr: nil,
		},
		{
			name: "Invalid File Paths",
			input: `diff --git a/file1.go
index 123abc..456def 100644
--- a/file1.go
+++ b/file1.go`,
			want:    nil,
			wantErr: errors.New("invalid file paths"),
		},
		{
			name: "Missing Index",
			input: `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go`,
			want:    nil,
			wantErr: errors.New("invalid git diff format"),
		},
		{
			name:    "Empty Input",
			input:   "",
			want:    nil,
			wantErr: errors.New("invalid git diff format"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGitDiffFileString(tt.input)
			if (err != nil) != (tt.wantErr != nil) || (err != nil && err.Error() != tt.wantErr.Error()) {
				t.Errorf("parseGitDiffFileString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseGitDiffFileString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseGitDiff(t *testing.T) {
	diff := `diff --git a/file1.go b/file1.go
index 123abc..456def 100644
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
+import "fmt"
diff --git a/go.mod b/go.mod
index 234bcd..567efg 100644
--- a/go.mod
+++ b/go.mod
@@ -2,5 +2,6 @@
+module example.com/project`

	ignoreList := []string{".mod"}

	expected := []*GitDiff{
		{
			FilePathOld:  "a/file1.go",
			FilePathNew:  "b/file1.go",
			Index:        "123abc..456def 100644",
			DiffContents: "--- a/file1.go\n+++ b/file1.go\n@@ -1,3 +1,4 @@\n+import \"fmt\"",
		},
		// go.mod is ignored based on the ignoreList
	}

	result := ParseGitDiff(diff, ignoreList)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ParseGitDiff() = %v, want %v", result, expected)
	}
}

func TestGetDiffContents(t *testing.T) {
	// Mock HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/valid-diff" {
			_, _ = w.Write([]byte("mock diff content"))
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer testServer.Close()

	// Test cases
	tests := []struct {
		name    string
		diffURL string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid Diff URL",
			diffURL: testServer.URL + "/valid-diff",
			want:    "mock diff content",
			wantErr: false,
		},
		{
			name:    "Invalid Diff URL",
			diffURL: testServer.URL + "/invalid-diff",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDiffContents(tt.diffURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDiffContents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getDiffContents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchIgnoreFilter_SinglePattern(t *testing.T) {
	testCases := []struct {
		name            string
		gitDiff         *GitDiff
		ignorePattern   string
		shouldBeIgnored bool
	}{
		{
			name:            "Ignore node_modules directory",
			gitDiff:         &GitDiff{FilePathNew: "node_modules/package.json"},
			ignorePattern:   `node_modules\/.*`,
			shouldBeIgnored: true,
		},
		{
			name:            "Ignore .env files",
			gitDiff:         &GitDiff{FilePathNew: ".env.local"},
			ignorePattern:   `.*\.env`,
			shouldBeIgnored: true,
		},
		{
			name:            "Ignore temporary files",
			gitDiff:         &GitDiff{FilePathNew: "temp/file.tmp"},
			ignorePattern:   `.*\.tmp`,
			shouldBeIgnored: true,
		},
		{
			name:            "Ignore binary files",
			gitDiff:         &GitDiff{FilePathNew: "bin/executable"},
			ignorePattern:   `bin\/.*`,
			shouldBeIgnored: true,
		},
		{
			name:            "Ignore source files",
			gitDiff:         &GitDiff{FilePathNew: "src/main.go"},
			ignorePattern:   `src\/.*`,
			shouldBeIgnored: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ignoreList := []string{tc.ignorePattern}
			if matchIgnoreFilter(tc.gitDiff, ignoreList) != tc.shouldBeIgnored {
				t.Errorf("matchIgnoreFilter() for %s returned incorrect result, got: %t, want: %t", tc.name, !tc.shouldBeIgnored, tc.shouldBeIgnored)
			}
		})
	}
}

func TestMatchIgnoreFilter_NonMatchingPattern(t *testing.T) {
	gitDiff := &GitDiff{FilePathNew: "src/main.go"}

	ignoreList := []string{
		`node_modules\/.*`, // Regex pattern to ignore node_modules directory
		`.*\.env`,          // Regex pattern to ignore .env files
		`.*\.tmp`,          // Regex pattern to ignore temporary files
		`bin\/.*`,          // Regex pattern to ignore binary files in bin directory
	}

	if matchIgnoreFilter(gitDiff, ignoreList) {
		t.Errorf("matchIgnoreFilter() incorrectly ignored file %s, but it should not have", gitDiff.FilePathNew)
	}
}

func TestMatchIgnoreFilter_EmptyIgnoreList(t *testing.T) {
	gitDiff := &GitDiff{FilePathNew: "src/main.go"}

	ignoreList := []string{} // Empty ignore list

	if matchIgnoreFilter(gitDiff, ignoreList) {
		t.Errorf("matchIgnoreFilter() incorrectly ignored file %s with an empty ignore list", gitDiff.FilePathNew)
	}
}

func TestMatchIgnoreFilter_InvalidRegexPattern(t *testing.T) {
	gitDiff := &GitDiff{FilePathNew: "src/main.go"}

	ignoreList := []string{
		"[invalid-regex", // An intentionally invalid regex pattern
	}

	if matchIgnoreFilter(gitDiff, ignoreList) {
		t.Errorf(
			"matchIgnoreFilter() incorrectly ignored file %s with an invalid regex pattern",
			gitDiff.FilePathNew,
		)
	}
}

func TestMatchFile_MatchingPattern(t *testing.T) {
	filePath := "src/main.go"
	pattern := `src\/.*\.go` // Regex pattern to match Go files in the src directory

	match, err := matchFile(pattern, filePath)
	if err != nil {
		t.Errorf("matchFile() returned an error: %v", err)
	}

	if !match {
		t.Errorf("matchFile() did not match file %s with pattern %s, but it should have", filePath, pattern)
	}
}

func TestMatchFile_NonMatchingPattern(t *testing.T) {
	filePath := "images/photo.jpg"
	pattern := `src\/.*\.go` // Regex pattern to match Go files in the src directory

	match, err := matchFile(pattern, filePath)
	if err != nil {
		t.Errorf("matchFile() returned an error: %v", err)
	}

	if match {
		t.Errorf("matchFile() incorrectly matched file %s with pattern %s", filePath, pattern)
	}
}

func TestMatchFile_InvalidRegexPattern(t *testing.T) {
	filePath := "src/main.go"
	pattern := `[invalid-regex` // An intentionally invalid regex pattern

	_, err := matchFile(pattern, filePath)
	if err == nil {
		t.Errorf("matchFile() did not return an error for invalid pattern %s", pattern)
	}
}

func TestMatchFile_EmptyFilePath(t *testing.T) {
	filePath := ""
	pattern := `src\/.*\.go`

	match, err := matchFile(pattern, filePath)
	if err != nil {
		t.Errorf("matchFile() returned an error: %v", err)
	}

	if match {
		t.Errorf("matchFile() incorrectly matched an empty file path with pattern %s", pattern)
	}
}

func TestMatchFile_EmptyPattern(t *testing.T) {
	filePath := "src/main.go"
	pattern := ""

	match, err := matchFile(pattern, filePath)
	if err != nil {
		t.Errorf("matchFile() returned an error: %v", err)
	}

	if match {
		t.Errorf("matchFile() incorrectly matched file %s with an empty pattern", filePath)
	}
}

func TestMatchFile_ComplexPattern(t *testing.T) {
	filePath := "src/main_test.go"
	pattern := `src\/[a-z]+_test\.go` // Regex pattern for test files in the src directory

	match, err := matchFile(pattern, filePath)
	if err != nil {
		t.Errorf("matchFile() returned an error: %v", err)
	}

	if !match {
		t.Errorf("matchFile() failed to match file %s with complex pattern %s", filePath, pattern)
	}
}

func TestMatchFile_SpecialCharactersInFilePath(t *testing.T) {
	filePath := "src/[special]file.go"
	pattern := `src\/\[special\]file\.go`

	match, err := matchFile(pattern, filePath)
	if err != nil {
		t.Errorf("matchFile() returned an error: %v", err)
	}

	if !match {
		t.Errorf("matchFile() failed to match file %s with pattern %s containing special characters", filePath, pattern)
	}
}

func TestMatchFile_CaseInsensitivePattern(t *testing.T) {
	filePath := "src/Main.go"
	pattern := `(?i)src\/main\.go` // Case insensitive pattern

	match, err := matchFile(pattern, filePath)
	if err != nil {
		t.Errorf("matchFile() returned an error: %v", err)
	}

	if !match {
		t.Errorf("matchFile() failed to match file %s with case insensitive pattern %s", filePath, pattern)
	}
}

func TestMatchFile_MetaCharactersInPattern(t *testing.T) {
	filePath := "src/functions.go"
	pattern := `src\/.*\.go` // Pattern using regex meta characters

	match, err := matchFile(pattern, filePath)
	if err != nil {
		t.Errorf("matchFile() returned an error: %v", err)
	}

	if !match {
		t.Errorf("matchFile() failed to match file %s with pattern %s using meta characters", filePath, pattern)
	}
}
