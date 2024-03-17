package github_test

import (
	"errors"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"infro.io/infro-core/internal/github"
	"infro.io/infro-core/internal/model"
)

const expected = `## Infro diff for 8d508b8
<img src="https://argo-cd.readthedocs.io/en/stable/assets/favicon.png" width="20"/> **my-argo-cluster > my-app** *(1 files changed)*
<details open>

~~~diff
>   this_is_a: change
~~~

</details>

<img src="https://registry.terraform.io/images/favicons/favicon-32x32.png" width="20"/> **terraform > my-terraform** ❌
>some dry run error
`

func TestFormat(t *testing.T) {
	body := github.Format("8d508b82cc188ce7c8244bfc280f24d75b4a1e7e", []model.DryRun{
		{
			DeployerName:   "my-argo-cluster",
			DeployerType:   model.DeployerTypeArgoCD,
			DeploymentName: "my-app",
			Diff:           lo.ToPtr(">   this_is_a: change"),
			Err:            nil,
			FilesChanged:   1,
		},
		{
			DeployerName:   "terraform",
			DeployerType:   model.DeployerTypeTerraform,
			DeploymentName: "my-terraform",
			Diff:           nil,
			Err:            errors.New("some dry run error"),
			FilesChanged:   1,
		},
	})
	require.Equal(t, expected, *body)
}
