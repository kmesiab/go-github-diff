package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/v57/github"

	ghdiff "github.com/kmesiab/go-github-diff"
)

func main() {
	// Parse a pull request URL
	prURL, _ := ghdiff.ParsePullRequestURL(
		"https://github.com/kmesiab/ai-code-critic/pull/50",
	)

	// Create a new github client
	client := github.NewClient(nil)

	// Process the pull request
	prString, err := ghdiff.GetPullRequest(context.TODO(), prURL, client)
	if err != nil {
		fmt.Printf("Error getting pull request: %s\n", err)
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
