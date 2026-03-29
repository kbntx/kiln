package github

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	gogithub "github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

// RealClient implements GitHubClient using the go-github library.
type RealClient struct {
	client *gogithub.Client
}

// NewRealClient creates a new GitHub client authenticated with the given token.
func NewRealClient(token string) *RealClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	return &RealClient{
		client: gogithub.NewClient(tc),
	}
}

// ListOpenPRs lists all open pull requests for the given repository.
func (r *RealClient) ListOpenPRs(ctx context.Context, owner, repo string) ([]PullRequest, error) {
	var allPRs []PullRequest

	opts := &gogithub.PullRequestListOptions{
		State: "open",
		ListOptions: gogithub.ListOptions{
			PerPage: 50,
		},
	}

	for {
		prs, resp, err := r.client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("listing pull requests: %w", err)
		}

		for _, pr := range prs {
			approved, err := r.GetApprovalStatus(ctx, owner, repo, pr.GetNumber())
			if err != nil {
				return nil, fmt.Errorf("checking approval for PR #%d: %w", pr.GetNumber(), err)
			}

			allPRs = append(allPRs, mapPullRequest(pr, approved))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

// GetPR retrieves a single pull request by number.
func (r *RealClient) GetPR(ctx context.Context, owner, repo string, prID int) (*PullRequest, error) {
	pr, _, err := r.client.PullRequests.Get(ctx, owner, repo, prID)
	if err != nil {
		return nil, fmt.Errorf("getting pull request #%d: %w", prID, err)
	}

	approved, err := r.GetApprovalStatus(ctx, owner, repo, prID)
	if err != nil {
		return nil, fmt.Errorf("checking approval for PR #%d: %w", prID, err)
	}

	result := mapPullRequest(pr, approved)
	return &result, nil
}

// GetApprovalStatus checks whether a pull request has at least one approving review.
func (r *RealClient) GetApprovalStatus(ctx context.Context, owner, repo string, prID int) (bool, error) {
	reviews, _, err := r.client.PullRequests.ListReviews(ctx, owner, repo, prID, nil)
	if err != nil {
		return false, fmt.Errorf("listing reviews for PR #%d: %w", prID, err)
	}

	for _, review := range reviews {
		if review.GetState() == "APPROVED" {
			return true, nil
		}
	}

	return false, nil
}

// IsMember checks whether a user is a member of the given organization.
func (r *RealClient) IsMember(ctx context.Context, user, org string) (bool, error) {
	isMember, _, err := r.client.Organizations.IsMember(ctx, org, user)
	if err != nil {
		return false, fmt.Errorf("checking membership of %s in %s: %w", user, org, err)
	}
	return isMember, nil
}

// PostComment creates a comment on the given pull request.
func (r *RealClient) PostComment(ctx context.Context, owner, repo string, prID int, body string) error {
	comment := &gogithub.IssueComment{
		Body: gogithub.String(body),
	}
	_, _, err := r.client.Issues.CreateComment(ctx, owner, repo, prID, comment)
	if err != nil {
		return fmt.Errorf("posting comment on PR #%d: %w", prID, err)
	}
	return nil
}

// GetFileContent fetches a single file from a repository at a given ref (branch/tag/sha).
// If OVERRIDE_KILN_CONFIG is set and the requested file is kiln.yaml, it reads from that local path instead.
func (r *RealClient) GetFileContent(ctx context.Context, owner, repo, ref, path string) ([]byte, error) {
	if path == "kiln.yaml" {
		if override := os.Getenv("OVERRIDE_KILN_CONFIG"); override != "" {
			slog.Debug("using kiln config override", "path", override)
			return os.ReadFile(override)
		}
	}

	opts := &gogithub.RepositoryContentGetOptions{Ref: ref}
	file, _, _, err := r.client.Repositories.GetContents(ctx, owner, repo, path, opts)
	if err != nil {
		return nil, fmt.Errorf("get file %s: %w", path, err)
	}
	if file == nil {
		return nil, fmt.Errorf("get file %s: path is a directory, not a file", path)
	}
	content, err := file.GetContent()
	if err != nil {
		return nil, fmt.Errorf("decode file %s: %w", path, err)
	}
	return []byte(content), nil
}

// mapPullRequest converts a go-github PullRequest to our domain PullRequest.
func mapPullRequest(pr *gogithub.PullRequest, approved bool) PullRequest {
	return PullRequest{
		Number:       pr.GetNumber(),
		Title:        pr.GetTitle(),
		Author:       pr.GetUser().GetLogin(),
		AuthorAvatar: pr.GetUser().GetAvatarURL(),
		Branch:       pr.GetHead().GetRef(),
		BaseBranch:   pr.GetBase().GetRef(),
		HeadSHA:      pr.GetHead().GetSHA(),
		Approved:     approved,
		CreatedAt:    pr.GetCreatedAt().Time,
		UpdatedAt:    pr.GetUpdatedAt().Time,
	}
}
