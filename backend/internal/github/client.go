package github

import (
	"context"
	"time"
)

type PullRequest struct {
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	AuthorAvatar string    `json:"authorAvatar"`
	Branch       string    `json:"branch"`
	BaseBranch   string    `json:"baseBranch"`
	HeadSHA      string    `json:"headSha"`
	Approved     bool      `json:"approved"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type GitHubClient interface {
	ListOpenPRs(ctx context.Context, owner, repo string) ([]PullRequest, error)
	GetPR(ctx context.Context, owner, repo string, prID int) (*PullRequest, error)
	GetApprovalStatus(ctx context.Context, owner, repo string, prID int) (bool, error)
	IsMember(ctx context.Context, user, org string) (bool, error)
	PostComment(ctx context.Context, owner, repo string, prID int, body string) error
	GetFileContent(ctx context.Context, owner, repo, ref, path string) ([]byte, error)
}
