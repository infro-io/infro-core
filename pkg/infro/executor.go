package infro

import (
	"context"

	"github.com/go-playground/validator/v10"

	"infro.io/infro-core/internal/argocd"
	"infro.io/infro-core/internal/github"
	"infro.io/infro-core/internal/model"
	"infro.io/infro-core/internal/terraform"
)

type (
	Executor struct {
		deployerClients []DeployerClient
		vcsClient       VCSClient
	}
	DeployerClient interface {
		ExecuteDryRuns(context.Context, model.DryRunOpts) []model.DryRun
	}
	VCSClient interface {
		RepoURL(owner string, repo string) string
		UpsertComment(context.Context, model.UpsertCommentOpts) (*model.CommentMetadata, error)
	}
	CommentDiffOpts struct {
		Revision   string
		Owner      string
		Repo       string
		PullNumber int
	}
)

func NewExecutorFromConfig(cfg *Config) (*Executor, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, err
	}

	depClients := make([]DeployerClient, len(cfg.Deployers))
	for i, depCfg := range cfg.Deployers {
		var depClient DeployerClient
		switch v := depCfg.Value.(type) {
		case *ArgoCDConfig:
			depClient = argocd.NewClient(v)
		case *TerraformConfig:
			depClient = terraform.NewClient(v)
		}
		depClients[i] = depClient
	}

	vcsClt := github.NewClient(cfg.VCS.Value.(*GithubConfig))
	return NewExecutor(depClients, vcsClt), nil
}

func NewExecutor(depClients []DeployerClient, vcsClient VCSClient) *Executor {
	return &Executor{depClients, vcsClient}
}

func (e *Executor) CommentDiffs(ctx context.Context, opts CommentDiffOpts) (*model.CommentMetadata, error) {
	var dryRuns []model.DryRun
	for _, depClient := range e.deployerClients {
		newDryRuns := depClient.ExecuteDryRuns(ctx, model.DryRunOpts{
			Revision: opts.Revision,
			RepoURL:  e.vcsClient.RepoURL(opts.Owner, opts.Repo),
		})
		dryRuns = append(dryRuns, newDryRuns...)
	}
	return e.vcsClient.UpsertComment(ctx, model.UpsertCommentOpts{
		DryRuns:    dryRuns,
		Owner:      opts.Owner,
		Repo:       opts.Repo,
		PullNumber: opts.PullNumber,
		Revision:   opts.Revision,
	})
}
