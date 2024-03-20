package argocd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/infro-io/infro-core/internal/model"
)

type (
	Client struct {
		cfg *Config
	}
	Config struct {
		Name      string `validate:"required"`
		Endpoint  string `validate:"required"`
		AuthToken string `validate:"required"`
	}
	appDiffOptions struct {
		Endpoint  string
		AuthToken string
		AppName   string
		Revision  string
	}
)

func NewClient(cfg *Config) *Client {
	return &Client{cfg}
}

func (c *Client) ExecuteDryRuns(ctx context.Context, opts model.DryRunOpts) ([]model.DryRun, error) {
	apps, err := c.listAppsForRepo(ctx, opts.RepoURL)
	if err != nil {
		return nil, err
	}
	dryRuns := make([]model.DryRun, len(apps))
	for i, app := range apps {
		dryRuns[i] = c.executeDryRun(ctx, app.Name, opts.Revision)
	}
	return dryRuns, nil
}

func (c *Client) listAppsForRepo(ctx context.Context, repoURL string) ([]v1alpha1.Application, error) {
	apiClient, err := apiclient.NewClient(&apiclient.ClientOptions{
		ServerAddr: c.cfg.Endpoint,
		AuthToken:  c.cfg.AuthToken,
		Insecure:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate argocd client: %w", err)
	}
	_, appClient, err := apiClient.NewApplicationClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to argocd: %w", err)
	}
	applications, err := appClient.List(ctx, &application.ApplicationQuery{
		Repo: lo.ToPtr(repoURL),
	})
	if s, ok := status.FromError(err); ok {
		if s.Code() == codes.Unauthenticated {
			return nil, fmt.Errorf("bad auth token: %s", s.Message())
		}
		if s.Code() == codes.Unimplemented {
			return nil, fmt.Errorf("most likely not an Argo CD endpoint: %s", s.Message())
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list argocd applications for %s: %w", c.cfg.Name, err)
	}
	return applications.Items, nil
}

func (c *Client) executeDryRun(ctx context.Context, appName string, revision string) model.DryRun {
	diff, err := getAppDiff(ctx, appDiffOptions{
		Endpoint:  c.cfg.Endpoint,
		AuthToken: c.cfg.AuthToken,
		AppName:   appName,
		Revision:  revision,
	})
	filesChanged := 0
	if diff != nil {
		filesChanged = strings.Count(*diff, "\n===== /")
	}
	return model.DryRun{
		DeployerName:   c.cfg.Name,
		DeployerType:   model.DeployerTypeArgoCD,
		DeploymentName: appName,
		Diff:           diff,
		Err:            err,
		FilesChanged:   filesChanged,
	}
}

func getAppDiff(ctx context.Context, opts appDiffOptions) (*string, error) {
	//nolint:gosec // sanitised elsewhere
	cmd := exec.CommandContext(ctx,
		"argocd", "app", "diff", opts.AppName,
		"--insecure", "--server-side-generate",
		"--server", opts.Endpoint,
		"--revision", opts.Revision,
		"--auth-token", opts.AuthToken)
	output, err := cmd.CombinedOutput()

	// cli returns exit code 0 and no content if there isn't a diff
	if err == nil {
		return nil, nil
	}
	var ee *exec.Error
	if errors.As(err, &ee) {
		return nil, &model.CliError{Reason: ee.Error()}
	}
	diff := lo.ToPtr(string(output))
	var xe *exec.ExitError
	if errors.As(err, &xe) {
		// cli returns exit code 1 if a diff is detected
		if xe.ExitCode() == 1 {
			return diff, nil
		}
	}
	return nil, &model.DiffError{Reason: *diff}
}
