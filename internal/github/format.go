package github

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/samber/lo"

	"github.com/infro-io/infro-core/internal/model"
)

const (
	longDiffLineCount = 30
)

func Format(revision string, dryRuns []model.DryRun) *string {
	s := fmt.Sprintf("## Infro diff for %s", revision[:7])
	for _, dryRun := range dryRuns {
		s += formatDryRun(dryRun)
	}
	return &s
}

//nolint:gochecknoglobals // fine for templates
var diffTemplate, _ = template.New("diff").Parse(`
{{.icon}} **{{.deployer}} > {{.deployment}}** *({{.files}} files changed)*
<details{{.open}}>

~~~diff
{{.diff}}
~~~

</details>
`)

//nolint:gochecknoglobals // fine for templates
var errTemplate, _ = template.New("diff").Parse(`
{{.icon}} **{{.deployer}} > {{.deployment}}** ❌
>{{.err}}
`)

func formatDryRun(dryRun model.DryRun) string {
	buf := new(bytes.Buffer)
	icon := fmt.Sprintf("<img src=\"%s\" width=\"20\"/>", getPic(dryRun.DeployerType))
	if dryRun.Err != nil {
		//nolint:errcheck // checked elsewhere
		errTemplate.Execute(buf, map[string]any{
			"icon":       icon,
			"deployer":   dryRun.DeployerName,
			"deployment": dryRun.DeploymentName,
			"err":        dryRun.Err.Error(),
		})
		return buf.String()
	}
	diffString := strings.Trim(*dryRun.Diff, "\n")
	lineCount := strings.Count(diffString, "\n")
	//nolint:errcheck // checked elsewhere
	diffTemplate.Execute(buf, map[string]any{
		"icon":       icon,
		"deployer":   dryRun.DeployerName,
		"deployment": dryRun.DeploymentName,
		"files":      dryRun.FilesChanged,
		"open":       lo.Ternary(lineCount < longDiffLineCount, " open", ""),
		"diff":       diffString,
	})
	return buf.String()
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
