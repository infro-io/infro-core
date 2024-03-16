package model

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
)

const VCSTypeGithub = "github"
