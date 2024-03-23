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
	Comment struct {
		ID   int64   `json:"id"`
		Body *string `json:"body"`
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
