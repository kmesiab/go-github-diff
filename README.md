# GitHub Pull Request Diff Library

![Golang](https://img.shields.io/badge/Go-00add8.svg?labelColor=171e21&style=for-the-badge&logo=go)

![Build](https://github.com/kmesiab/go-github-diff/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/kmesiab/go-github-diff)](https://goreportcard.com/report/github.com/kmesiab/go-github-diff)

This Go library provides a set of functions to interact with GitHub Pull
Requests, specifically for parsing and processing the contents of Git
diffs associated with pull requests.

## Features

- Parse GitHub Pull Request URLs to extract owner, repository, and PR number.
- Retrieve the contents of a Pull Request's Git diff from GitHub.
- Parse combined Git diffs into individual file diffs.
- Filter out file diffs based on a list of ignored file extensions.
- Comprehensive regex-based file path matching for filtering file diffs.
- Robust and extensive unit testing to ensure reliability and functionality.

## Installation

To use this library, install it using `go get`:

```bash
go get github.com/kmesiab/go-github-diff
```

## Usage

### ParsePullRequestURL

```go
prURL, err := github.ParsePullRequestURL(
    "https://github.com/username/repository/pull/123", 
)

if err != nil {
    // Handle error
}

// Use prURL.Owner, prURL.Repo, and prURL.PRNumber
```

### GetPullRequest

```go
prURL := &github.PullRequestURL{
    Owner: "username",
    Repo: "repository",
    PRNumber: 123,
}
diff, err := github.GetPullRequest(context.Background(), prURL)

if err != nil {
    // Handle error
}

// Use diff as a string containing the Git diff
```

### ParseGitDiff

```go
diff := "..." // Git diff string
ignoreList := []string{".md", ".txt"}
gitDiffs := github.ParseGitDiff(diff, ignoreList)
for _, gitDiff := range gitDiffs {
// Process each gitDiff
}
```

---

## Contributing

Contributions to this library are welcome! Here's how you can contribute:

### Forking the Repository

1. Go to the GitHub repository page: [GitHub Repository URL]
2. Click on the 'Fork' button at the top right corner of the page. This creates
a copy
of the repository in your GitHub account.
3. Clone the forked repository to your local machine:

   ```bash
   git clone https://github.com/yourusername/reponame.git
   ```

4. Navigate to the cloned directory:

   ```bash
   cd reponame
   ```

### Making Changes and Contributing

1. Create a new branch for your changes:

   ```bash
   git checkout -b feature-branch-name
   ```

2. Make your changes in the new branch. Be sure to follow the project's coding
standards and guidelines.
3. Commit your changes with a descriptive message:

   ```bash
   git commit -am "Add a brief description of your changes"
   ```

4. Push the changes to your fork:

   ```bash
   git push origin feature-branch-name
   ```

5. Go to your forked repository on GitHub, and click 'New Pull Request'.
6. Ensure the 'base fork' points to the original repository, and the 'head fork'
points to your fork.
7. Provide a clear and detailed description of your changes in the pull request.
Reference any related issues.
8. Click 'Create Pull Request' to submit your changes for review.

### After Your Pull Request

- Wait for feedback or approval from the repository maintainers.
- If requested, make any necessary updates to your pull request.
- Once your pull request is merged, you can pull the changes from the original
repository to keep your fork up to date.

Thank you for your contributions!
