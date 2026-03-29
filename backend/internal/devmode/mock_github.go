package devmode

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kbntx/kiln/internal/github"
)

var _ github.GitHubClient = (*MockGitHubClient)(nil)

type MockGitHubClient struct{}

func (m *MockGitHubClient) ListOpenPRs(_ context.Context, owner, repo string) ([]github.PullRequest, error) {
	now := time.Now()
	prs := make([]github.PullRequest, 5)
	for i := range prs {
		n := i + 1
		prs[i] = github.PullRequest{
			Number:       n,
			Title:        fmt.Sprintf("[dev] Update %s infra (PR #%d)", repo, n),
			Author:       "dev-user",
			AuthorAvatar: "https://avatars.githubusercontent.com/u/0?v=4",
			Branch:       fmt.Sprintf("feature/change-%d", n),
			BaseBranch:   "main",
			Approved:     n%2 == 0,
			CreatedAt:    now.Add(-time.Duration(n) * 24 * time.Hour),
			UpdatedAt:    now.Add(-time.Duration(n) * time.Hour),
		}
	}
	return prs, nil
}

func (m *MockGitHubClient) GetPR(_ context.Context, _, repo string, prID int) (*github.PullRequest, error) {
	now := time.Now()
	return &github.PullRequest{
		Number:       prID,
		Title:        fmt.Sprintf("[dev] Update %s infra (PR #%d)", repo, prID),
		Author:       "dev-user",
		AuthorAvatar: "https://avatars.githubusercontent.com/u/0?v=4",
		Branch:       fmt.Sprintf("feature/change-%d", prID),
		BaseBranch:   "main",
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
	log.Printf("[devmode] PostComment %s/%s#%d: %s", owner, repo, prID, body)
	return nil
}
