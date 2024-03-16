package github

import (
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"infro.io/infro-core/internal/model"
)

const (
	longDiffLineCount = 30
)

func Format(revision string, dryRuns []model.DryRun) *string {
	s := fmt.Sprintf("## Infro diff for %s\n", revision[:7])
	if len(dryRuns) == 0 {
		s += "*(no deployments)*"
		return &s
	}
	for _, dryRun := range dryRuns {
		icon := fmt.Sprintf("<img src=\"%s\" width=\"20\"/>", getPic(dryRun.DeployerType))
		s += fmt.Sprintf("%s **%s > %s** %s", icon, dryRun.DeployerName, dryRun.DeploymentName, formatDeployment(dryRun))
	}
	return &s
}

func formatDeployment(dryRun model.DryRun) string {
	if dryRun.Err != nil {
		if errors.As(dryRun.Err, &model.NoChangesError{}) {
			return "*(no changes)*\n"
		}
		return fmt.Sprintf("❌\n>%s\n\n", dryRun.Err.Error())
	}
	diffString := strings.Trim(*dryRun.Diff, "\n")
	lineCount := strings.Count(diffString, "\n")
	open := lineCount < longDiffLineCount
	detailsOpen := lo.Ternary(open, " open", "")
	return fmt.Sprintf("*(%d files changed)*\n<details%s>\n\n~~~diff\n%s\n~~~\n\n</details>\n\n",
		dryRun.FilesChanged, detailsOpen, diffString)
}

func getPic(deployerType string) string {
	switch deployerType {
	case model.DeployerTypeArgoCD:
		return "https://argo-cd.readthedocs.io/en/stable/assets/favicon.png"
	case model.DeployerTypeTerraform:
		return "https://registry.terraform.io/images/favicons/favicon-32x32.png"
	}
	return "https://infro.io/app/favicon.ico"
}
