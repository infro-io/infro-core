package github

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/google/go-github/v57/github"

	"infro.io/infro-core/internal/model"
)

type (
	Client struct {
		cfg *Config
	}
	Config struct {
		AuthToken string `validate:"required"`
	}
)

func NewClient(cfg *Config) *Client {
	return &Client{cfg}
}

func (c Client) RepoURL(owner string, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

func (c Client) UpsertComment(ctx context.Context, opts model.UpsertCommentOpts) (*model.CommentMetadata, error) {
	ghClient := gh.NewClient(nil).WithAuthToken(c.cfg.AuthToken)
	comments, _, err := ghClient.Issues.ListComments(ctx, opts.Owner, opts.Repo, opts.PullNumber, nil)
	if err != nil {
		return nil, err
	}
	comment := getCommentWithSubstring(comments, htmlTag)
	body := Format(opts.Revision, opts.DryRuns)
	*body += "\n" + htmlTag
	if comment == nil {
		newComment := &gh.IssueComment{Body: body}
		comment, _, err = ghClient.Issues.CreateComment(ctx, opts.Owner, opts.Repo, opts.PullNumber, newComment)
	} else {
		comment.Body = body
		comment, _, err = ghClient.Issues.EditComment(ctx, opts.Owner, opts.Repo, *comment.ID, comment)
	}
	if err != nil {
		return nil, err
	}
	return &model.CommentMetadata{CommentID: *comment.ID}, nil
}

func getCommentWithSubstring(comments []*gh.IssueComment, substring string) *gh.IssueComment {
	for _, cmt := range comments {
		if strings.Contains(*cmt.Body, substring) {
			return cmt
		}
	}
	return nil
}

const htmlTag = "<!-- INFRO -->"
