package github

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

var _ GitHubClient = (*MockGitHubClient)(nil)

type MockGitHubClient struct{}

func (m *MockGitHubClient) ListOpenPRs(_ context.Context, owner, repo string) ([]PullRequest, error) {
	now := time.Now()
	prs := make([]PullRequest, 5)
	for i := range prs {
		n := i + 1
		prs[i] = PullRequest{
			Number:       n,
			Title:        fmt.Sprintf("[dev] Update %s infra (PR #%d)", repo, n),
			Author:       "dev-user",
			AuthorAvatar: "https://avatars.githubusercontent.com/u/0?v=4",
			Branch:       fmt.Sprintf("feature/change-%d", n),
			BaseBranch:   "main",
			HeadSHA:      fmt.Sprintf("abc%04d", n),
			Approved:     n%2 == 0,
			CreatedAt:    now.Add(-time.Duration(n) * 24 * time.Hour),
			UpdatedAt:    now.Add(-time.Duration(n) * time.Hour),
		}
	}
	return prs, nil
}

func (m *MockGitHubClient) GetPR(_ context.Context, _, repo string, prID int) (*PullRequest, error) {
	now := time.Now()
	return &PullRequest{
		Number:       prID,
		Title:        fmt.Sprintf("[dev] Update %s infra (PR #%d)", repo, prID),
		Author:       "dev-user",
		AuthorAvatar: "https://avatars.githubusercontent.com/u/0?v=4",
		Branch:       fmt.Sprintf("feature/change-%d", prID),
		BaseBranch:   "main",
		HeadSHA:      fmt.Sprintf("abc%04d", prID),
		Approved:     prID%2 == 0,
		CreatedAt:    now.Add(-48 * time.Hour),
		UpdatedAt:    now.Add(-1 * time.Hour),
	}, nil
}

func (m *MockGitHubClient) GetApprovalStatus(_ context.Context, _, _ string, prID int) (bool, error) {
	return prID%2 == 0, nil
}

func (m *MockGitHubClient) IsMember(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}

func (m *MockGitHubClient) PostComment(_ context.Context, owner, repo string, prID int, body string) error {
	slog.Debug("mock PostComment", "owner", owner, "repo", repo, "pr", prID, "body", body)
	return nil
}

func (m *MockGitHubClient) GetFileContent(_ context.Context, _, _, ref, _ string) ([]byte, error) {
	// Simulate a PR with no kiln.yaml (last mock PR).
	if ref == "abc00005" {
		return nil, fmt.Errorf("get file kiln.yaml: 404 Not Found")
	}

	// TODO(pulumi): Add Pulumi project back when Pulumi support is implemented.
	return []byte(`profiles:
  dev:
    env:
      AWS_PROFILE: dev-account

terraform_version: "1.8.4"

projects:
  - name: networking
    dir: terraform
    engine: terraform
    stacks: [default]
    profile: dev
`), nil
}
