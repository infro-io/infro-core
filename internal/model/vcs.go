package model

import "time"

type (
	UpsertCommentOpts struct {
		DryRuns    []DryRun
		Owner      string
		Repo       string
		PullNumber int
		Revision   string
	}
	CommentMetadata struct {
		CommentID int64 `json:"commentId"`
	}
	PullRequest struct {
		Repo     string
		Number   int
		Revision string
	}
	ListPullRequestsOpts struct {
		Owner        string
		UpdatedSince time.Time
	}
)

const VCSTypeGithub = "github"
