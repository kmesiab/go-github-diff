package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/v57/github"

	ghdiff "github.com/kmesiab/go-github-diff" // Ensure this import path is correct
)

func main() {
	// Parse a pull request URL
	prURL, _ := ghdiff.ParsePullRequestURL(
		"https://github.com/kmesiab/ai-code-critic/pull/50",
	)

	// Create a new github client (You can pass a mock HTTP Client, instead of nil)
	client := github.NewClient(nil)

	// Create a new GitHubClientWrapper using the github client
	ghClient := ghdiff.GitHubClientWrapper{Client: client}

	// Process the pull request using the new function
	prString, err := ghdiff.GetPullRequestWithClient(context.TODO(), prURL, &ghClient)
	if err != nil {
		fmt.Printf("Error getting pull request: %s\n", err)

		return
	}

	// Print the raw diff string
	fmt.Printf("Diff:\n\n%s", prString)

	// Construct a list of file extensions to ignore
	ignoreList := []string{".mod"}

	// Parse the diff string into a list of diff files
	diffFiles := ghdiff.ParseGitDiff(prString, ignoreList)

	for _, diffFile := range diffFiles {

		fmtString := "File old: %s - File new: %s (index: %s)\n%s\n\n"

		fmt.Printf(fmtString,
			diffFile.FilePathOld,
			diffFile.FilePathNew,
			diffFile.Index,
			diffFile.DiffContents,
		)
	}
}
