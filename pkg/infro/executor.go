package infro

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"infro.io/infro-core/internal/argocd"
	"infro.io/infro-core/internal/github"
	"infro.io/infro-core/internal/model"
	"infro.io/infro-core/internal/terraform"
	"infro.io/infro-core/internal/xzap"
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
		ListPullRequests(context.Context, model.ListPullRequestsOpts) ([]model.PullRequest, error)
	}
	CommentOpts struct {
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

func (e *Executor) Comment(ctx context.Context, opts CommentOpts) (*model.CommentMetadata, error) {
	log := xzap.FromContext(ctx)
	log = log.With(
		zap.String("repo", opts.Repo),
		zap.String("revision", opts.Revision),
		zap.Int("pullNumber", opts.PullNumber))
	ctx = xzap.NewContext(ctx, log)

	var dryRuns []model.DryRun
	for _, depClient := range e.deployerClients {
		newDryRuns := depClient.ExecuteDryRuns(ctx, model.DryRunOpts{
			Revision: opts.Revision,
			RepoURL:  e.vcsClient.RepoURL(opts.Owner, opts.Repo),
		})
		for _, newDryRun := range newDryRuns {
			if newDryRun.Diff == nil && newDryRun.Err == nil {
				log.Info("no diff for deployment",
					zap.String("deployer", newDryRun.DeployerName),
					zap.String("deployment", newDryRun.DeploymentName))
				continue
			}
			dryRuns = append(dryRuns, newDryRun)
		}
	}
	if len(dryRuns) == 0 {
		log.Info("no diffs, skipping comment")
		return nil, nil
	}
	return e.vcsClient.UpsertComment(ctx, model.UpsertCommentOpts{
		DryRuns:    dryRuns,
		Owner:      opts.Owner,
		Repo:       opts.Repo,
		PullNumber: opts.PullNumber,
		Revision:   opts.Revision,
	})
}

func (e *Executor) Poll(ctx context.Context, owner string, interval time.Duration) {
	log := xzap.FromContext(ctx)
	log = log.With(zap.String("owner", owner))
	ctx = xzap.NewContext(ctx, log)
	lastPoll := time.Now()
	for {
		time.Sleep(interval)
		start := time.Now()
		if err := e.CommentOnPullRequests(ctx, owner, lastPoll); err != nil {
			log.Error("failed to list pull requests", zap.Error(err))
		}
		lastPoll = start
	}
}

func (e *Executor) CommentOnPullRequests(ctx context.Context, owner string, updatedSince time.Time) error {
	log := xzap.FromContext(ctx)
	pulls, err := e.vcsClient.ListPullRequests(ctx, model.ListPullRequestsOpts{
		Owner:        owner,
		UpdatedSince: updatedSince,
	})
	if err != nil {
		return err
	}
	for _, pull := range pulls {
		_, err = e.Comment(ctx, CommentOpts{
			Revision:   pull.Revision,
			Owner:      owner,
			Repo:       pull.Repo,
			PullNumber: pull.Number,
		})
		if err != nil {
			log.Error("failed to comment", zap.Error(err))
		}
	}
	return nil
}
