package model

type (
	DryRun struct {
		DeployerName   string
		DeployerType   string
		DeploymentName string

		// diff in Github markdown syntax or nil if no diff
		Diff *string

		// err to bubble up
		Err error

		FilesChanged int
	}
	DryRunOpts struct {
		Revision string
		RepoURL  string
	}
)
