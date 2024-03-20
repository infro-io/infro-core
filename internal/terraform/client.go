package terraform

import (
	"bytes"
	"context"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"

	"github.com/infro-io/infro-core/internal/model"
)

type (
	Client struct {
		cfg *Config
	}
	Config struct {
		Name    string `validate:"required"`
		Workdir string `validate:"required"`
	}
)

func NewClient(cfg *Config) *Client {
	return &Client{cfg}
}

func (c *Client) ExecuteDryRuns(ctx context.Context, _ model.DryRunOpts) ([]model.DryRun, error) {
	dryRun, err := c.executeDryRun(ctx)
	if err != nil {
		return nil, err
	}
	return []model.DryRun{*dryRun}, nil
}

func (c *Client) executeDryRun(ctx context.Context) (*model.DryRun, error) {
	dryRun := &model.DryRun{
		DeployerName:   "terraform",
		DeployerType:   model.DeployerTypeTerraform,
		DeploymentName: c.cfg.Name,
	}
	tf, err := tfexec.NewTerraform(c.cfg.Workdir, "terraform")
	if err != nil {
		return nil, err
	}
	if err = tf.Init(ctx, tfexec.Upgrade(true)); err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	tf.SetStdout(buf)
	changed, err := tf.Plan(ctx)
	if err != nil {
		dryRun.Err = err
		return dryRun, nil //nolint:nilerr // on purpose here
	}
	if changed {
		diff := formatDiffMarkdown(buf.String())
		dryRun.Diff = &diff
		dryRun.FilesChanged = strings.Count(diff, "\n  #")
	}
	return dryRun, nil
}

func formatDiffMarkdown(diff string) string {
	lines := strings.Split(diff, "\n")
	for i, line := range lines {
		switch firstChar(line) {
		case "+":
			lines[i] = "+" + removeFirst(line, "+")
		case "-":
			lines[i] = "-" + removeFirst(line, "-")
		}
	}
	return strings.Join(lines, "\n")
}

func firstChar(str string) string {
	trimmed := strings.TrimSpace(str)
	if trimmed == "" {
		return ""
	}
	return string(trimmed[0])
}

func removeFirst(str string, substring string) string {
	return strings.Replace(str, substring, "", 1)
}
