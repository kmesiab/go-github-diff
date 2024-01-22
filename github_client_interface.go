package github

import (
	"context"

	"github.com/google/go-github/v57/github"
)

// GitHubClientInterface defines an interface for interacting with GitHub,
// specifically for retrieving pull requests. This interface can be implemented
// by any client that can fetch pull request data from GitHub.
type GitHubClientInterface interface {
	// Get retrieves a specific pull request from GitHub based on the provided
	// owner, repository name, and pull request number. It returns the pull request
	// details along with the response from the GitHub API.
	Get(
		ctx context.Context,
		owner string,
		repo string,
		number int,
	) (*github.PullRequest, *github.Response, error)
}

// GitHubClientWrapper is a wrapper around the official GitHub client provided
// by the go-github library. It implements the GitHubClientInterface, enabling
// it to fetch pull request data.  Wrap the official GitHub client in this struct
// to use it with the go-github-diff library.
type GitHubClientWrapper struct {
	*github.Client
}

// Get fetches a specific pull request from GitHub using the official GitHub client.
// It takes the context, owner, repository name, and pull request number as parameters
// and returns the pull request details along with the response from the GitHub API.
func (c *GitHubClientWrapper) Get(
	ctx context.Context,
	owner string,
	repo string,
	number int,
) (*github.PullRequest, *github.Response, error) {
	return c.PullRequests.Get(ctx, owner, repo, number)
}

// MockGitClient is a mock implementation of the GitHubClientInterface, intended for
// use in unit tests. It allows for setting custom behavior for the Get method, enabling
// developers to test their code without making actual API calls to GitHub.
type MockGitClient struct {
	// MockGet is a function that simulates the Get method of GitHubClientInterface.
	// This function can be customized in test scenarios to return specific values or errors.
	MockGet func(ctx context.Context, owner string, repo string, number int) (*github.PullRequest, *github.Response, error)
}

// Get calls the mock implementation of the Get method. If MockGet is set to a custom function,
// that function is executed and its result returned. If MockGet is not set, the method returns
// nil values, simulating no data being fetched.
func (m *MockGitClient) Get(ctx context.Context, owner string, repo string, number int) (*github.PullRequest, *github.Response, error) {
	if m.MockGet != nil {
		return m.MockGet(ctx, owner, repo, number)
	}
	return nil, nil, nil
}
