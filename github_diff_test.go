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
