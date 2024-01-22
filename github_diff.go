package github

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v57/github"
)

type PullRequestURL struct {
	Owner    string
	Repo     string
	PRNumber int
}

type GitDiff struct {
	// FilePathOld represents the old file path in the diff, typically
	// indicated by a line starting with "---". This is the file path
	// before the changes were made.
	FilePathOld string

	// FilePathNew represents the new file path in the diff, typically
	// indicated by a line starting with "+++ ". This is the file path
	// after the changes were made. In most cases, it is the same as
	// FilePathOld unless the file was renamed or moved.
	FilePathNew string

	// Index is a string that usually contains the hash values before
	// and after the changes, along with some additional metadata.
	// This line typically starts with "index" in the diff output.
	Index string

	// DiffContents contains the actual content of the diff. This part
	// of the struct includes the changes made to the file, typically
	// represented by lines starting with "+" (additions) or "-"
	// (deletions). It includes all the lines that show the modifications
	// to the file.
	DiffContents string
}

// ParsePullRequestURL parses a GitHub pull request URL and returns the owner, repository,
// and pull request number. The function expects a standard GitHub pull request URL format.
// It splits the URL into segments and extracts the relevant information.
//
// The expected URL format is: https://github.com/[owner]/[repo]/pull/[prNumber]
// where [owner] is the GitHub username or organization name, [repo] is the repository name,
// and [prNumber] is the pull request number.
//
// Parameters:
//   - pullRequestURL: A string representing the full URL of a GitHub pull request.
//
// Returns:
//   - A pointer to a PullRequestURL struct containing the extracted information (owner, repo, PRNumber).
//   - An error if the URL format is invalid or if the pull request number cannot be converted to an integer.
//
// Example:
//
//	prURL, err := ParsePullRequestURL("https://github.com/username/repository/pull/123")
//	if err != nil {
//	  // Handle error
//	}
//	// Use prURL.Owner, prURL.Repo, and prURL.PRNumber
//
// This function is particularly useful for applications that need to process or respond to GitHub pull requests,
// allowing them to easily extract and use the key components of a pull request URL.
func ParsePullRequestURL(pullRequestURL string) (*PullRequestURL, error) {
	parts := strings.Split(pullRequestURL, "/")

	if len(parts) != 7 {
		return nil, errors.New("invalid pull request URL")
	}

	owner := parts[3]
	repo := parts[4]
	prNumber, err := strconv.Atoi(parts[6])
	if err != nil {
		return nil, err
	}

	return &PullRequestURL{
		Owner:    owner,
		Repo:     repo,
		PRNumber: prNumber,
	}, nil
}

// GetPullRequest retrieves the contents of a pull request's Git diff from GitHub.
// The function takes a context and a PullRequestURL struct, which contains the
// information needed to identify the specific pull request. It uses the GitHub API
// client to fetch the pull request and then calls getDiffContents to obtain the
// raw diff data.
//
// Parameters:
//   - ctx: A context.Context object, which allows for managing the lifecycle of
//     the request, such as canceling it or setting a timeout.
//   - pr: A pointer to a PullRequestURL struct, which contains the owner,
//     repository, and pull request number required to identify the pull request.
//
// Returns:
//   - A string containing the raw contents of the Git diff for the specified pull request.
//   - An error if the pull request retrieval fails or if there is an issue obtaining
//     the diff contents.
//
// The function first creates a new GitHub API client. It then uses this client to
// fetch the pull request specified by the PullRequestURL struct. If the pull request
// is successfully retrieved, the function extracts the URL to the pull request's diff
// and uses getDiffContents to fetch the diff data.
//
// Example:
//
//	prURL := &PullRequestURL{Owner: "username", Repo: "repository", PRNumber: 123}
//	diff, err := GetPullRequest(context.Background(), prURL)
//	if err != nil {
//	  // Handle error
//	}
//	// Use diff as a string containing the Git diff
//
// This function is useful in applications that need to programmatically access
// and process the contents of pull requests from GitHub, such as in automated
// code review tools, continuous integration systems, or other development workflows.
func GetPullRequest(ctx context.Context, pr *PullRequestURL, client *github.Client) (string, error) {
	pullRequest, _, err := client.PullRequests.Get(ctx, pr.Owner, pr.Repo, pr.PRNumber)
	if err != nil {
		return "", err
	}

	return getDiffContents(pullRequest.GetDiffURL())
}

// getDiffContents retrieves the contents of a Git diff from a specified URL. The function
// makes an HTTP GET request to the provided diffURL and returns the content as a string.
// This function is designed to work with URLs pointing to raw diff data, typically used
// in the context of GitHub or similar version control systems.
//
// Parameters:
//   - diffURL: A string representing the URL from which the Git diff contents are to be retrieved.
//
// Returns:
//   - A string containing the contents of the Git diff.
//   - An error if the HTTP request fails, or if reading the response body fails.
//
// The function handles HTTP errors and read errors by returning an empty string and the
// respective error. It ensures that the body of the HTTP response is read completely into
// a byte slice, which is then converted into a string.
//
// Example:
//
//	diff, err := getDiffContents("https://github.com/user/repo/pull/123.diff")
//	if err != nil {
//	  // Handle error
//	}
//	// Use diff as a string containing the Git diff
//
// This function is useful in scenarios where an application needs to process or analyze
// the contents of a Git diff, such as in automated code review tools, continuous integration
// systems, or other applications that interact with version control systems.
func getDiffContents(diffURL string) (string, error) {
	diffContents, err := http.Get(diffURL)
	if err != nil {
		return "", err
	}

	bodyBytes, err := io.ReadAll(diffContents.Body)
	if err != nil {
		return "", err
	}

	if diffContents.StatusCode != http.StatusOK {
		return "", errors.New("failed to get diff contents")
	}

	return string(bodyBytes), nil
}

// ParseGitDiff takes a string representing a combined Git diff and a list of
// file extensions to ignore. It returns a slice of GitDiff structs, each representing
// a parsed file diff. The function performs the following steps:
//  1. Splits the combined Git diff into individual file diffs using the
//     splitDiffIntoFiles function. This function looks for "diff --git" as a
//     delimiter to separate each file's diff.
//  2. Iterates over each file diff string. For each string, it:
//     a. Attempts to parse the string into a GitDiff struct using the
//     parseGitDiffFileString function. This function extracts the old and new
//     file paths, index information, and the actual diff content.
//     b. Checks for parsing errors. If an error occurs, it skips the current file
//     diff and continues with the next one.
//  3. Filters out file diffs based on the provided ignore list. The ignore list
//     contains file extensions (e.g., ".mod"). The function uses the
//     getFileExtension helper to extract the file extension from the new file path
//     (FilePathNew) of each GitDiff struct. If the extension matches any in the
//     ignore list, the file diff is skipped.
//  4. Appends the successfully parsed and non-ignored GitDiff structs to the
//     filteredList slice.
//
// Parameters:
//   - diff: A string representing the combined Git diff.
//   - ignoreList: A slice of strings representing the file extensions to ignore.
//
// Returns:
//   - A slice of GitDiff structs, each representing a parsed and non-ignored file diff.
func ParseGitDiff(diff string, ignoreList []string) []*GitDiff {
	files := splitDiffIntoFiles(diff)
	var filteredList []*GitDiff

	for _, file := range files {

		gitDiff, err := parseGitDiffFileString(file)

		if err != nil {
			continue
		}

		if matchIgnoreFilter(gitDiff, ignoreList) {
			continue
		}

		filteredList = append(filteredList, gitDiff)
	}

	return filteredList
}

func matchIgnoreFilter(file *GitDiff, ignoreList []string) bool {

	for _, pattern := range ignoreList {
		match, err := matchFile(pattern, file.FilePathNew)

		if err != nil {
			// consider finding a way to notify the caller
			// an error has occurred.
			return false
		}

		if match {

			return true
		}
	}

	return false
}

// matchFile takes a regex pattern and a file path and returns true if the
// file path matches the pattern, and false otherwise. It returns an error
// if the regex pattern is invalid.
func matchFile(pattern, file string) (bool, error) {

	if pattern == "" {

		return false, nil
	}

	rx, err := regexp.Compile(pattern)

	if err != nil {
		return false, err
	}

	return rx.MatchString(file), nil
}

// splitDiffIntoFiles splits a single diff string into a slice of
// strings, where each string represents the diff of an individual file.
// It assumes that 'diff --git' is used as a delimiter between file diffs.
func splitDiffIntoFiles(diff string) []string {
	var files []string
	var curFile strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(diff))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "diff --git") {
			// Detected start of new file
			if curFile.Len() > 0 {
				files = append(files, strings.TrimSpace(curFile.String()))
				curFile.Reset()
			}
			curFile.WriteString(line + "\n")
		} else {
			curFile.WriteString(line + "\n")
		}
	}

	// Add the last file diff to the list
	if curFile.Len() > 0 {
		files = append(files, strings.TrimSpace(curFile.String()))
	}

	return files
}

// ParseGitDiffFileString takes a string input representing a Git diff of a single file
// and returns a GitDiff struct containing the parsed information. The input
// string is expected to contain at least four lines, including the file paths
// line, the index line, and the diff content. The function performs the following
// steps to parse the diff:
//  1. Split the input string into lines.
//  2. Validate that there are enough lines to form a valid Git diff.
//  3. Extract the old and new file paths from the first line. The line is
//     expected to contain two file paths separated by a space.
//  4. Extract the index information from the second line. The line should
//     start with "index " followed by the index information.
//  5. Join the remaining lines, starting from the third line, to form the
//     diff content.
//
// The function returns an error if the input is not in the expected format,
// such as if there are not enough lines, if the file paths line is invalid,
// or if the index line is incorrectly formatted.
//
// Parameters:
//   - input: A string representing the Git diff of a single file.
//
// Returns:
//   - A pointer to a GitDiff struct containing the parsed file paths, index,
//     and diff content.
//   - An error if the input string is not in the expected format or if any
//     parsing step fails.
func parseGitDiffFileString(input string) (*GitDiff, error) {
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanLines)

	var (
		filePaths []string
		index     string
		diff      []string
	)

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "diff --git"):
			filePaths = strings.Fields(line)[2:]
			if len(filePaths) != 2 {
				return nil, errors.New("invalid file paths")
			}
		case strings.HasPrefix(line, "index "):
			index = strings.TrimSpace(line[6:])
		default:
			diff = append(diff, line)
		}
	}

	if len(filePaths) == 0 || len(index) == 0 || len(diff) == 0 {
		return nil, errors.New("invalid git diff format")
	}

	return &GitDiff{
		FilePathOld:  filePaths[0],
		FilePathNew:  filePaths[1],
		Index:        index,
		DiffContents: strings.Join(diff, "\n"),
	}, nil
}

func getFileExtension(path string) string {
	// If the path ends with a slash, it's a directory; return an empty string
	if strings.HasSuffix(path, string(filepath.Separator)) {
		return ""
	}

	fileName := filepath.Base(path)

	// Check if the path is a directory or empty
	if fileName == "." || fileName == "/" || fileName == "" {
		return ""
	}

	// Check for dot files (hidden files in Unix-based systems)
	if len(fileName) > 1 && fileName[0] == '.' && strings.Count(fileName, ".") == 1 {
		return fileName
	}

	// Extract the extension
	return filepath.Ext(fileName)
}
