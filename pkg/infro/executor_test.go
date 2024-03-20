package infro_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"infro.io/infro-core/internal/model"
	"infro.io/infro-core/pkg/infro"
)

type mockDeployerClient struct {
	mock.Mock
}

func (m *mockDeployerClient) ExecuteDryRuns(_ context.Context, _ model.DryRunOpts) ([]model.DryRun, error) {
	return []model.DryRun{{Diff: lo.ToPtr("diff")}}, nil
}

type mockVCSClient struct {
	mock.Mock
}

func (m *mockVCSClient) UpsertComment(ctx context.Context, opts model.UpsertCommentOpts) (*model.CommentMetadata, error) {
	return nil, m.Called(ctx, opts).Error(0)
}

func (m *mockVCSClient) RepoURL(_, _ string) string {
	return "http://example.org"
}

func (m *mockVCSClient) ListPullRequests(context.Context, model.ListPullRequestsOpts) ([]model.PullRequest, error) {
	return nil, nil
}

func TestCommentDiffs(t *testing.T) {
	// assemble
	mockDepClient := new(mockDeployerClient)
	mockVCSClt := new(mockVCSClient)
	mockVCSClt.On("UpsertComment", mock.Anything, mock.Anything).Return(nil)
	ex := infro.NewExecutor([]infro.DeployerClient{mockDepClient}, mockVCSClt)

	// act
	_, err := ex.Comment(context.Background(), infro.CommentOpts{
		Revision: "abc123",
		Owner:    "owner",
		Repo:     "repo",
	})

	// assert
	require.NoError(t, err)
	mockVCSClt.AssertNumberOfCalls(t, "UpsertComment", 1)
}

func TestNewExecutorFromConfig_Valid(t *testing.T) {
	e, err := infro.NewExecutorFromConfig(&infro.Config{
		Deployers: []infro.DeployerConfig{{Value: infro.ArgoCDConfig{
			Name:      "example",
			Endpoint:  "example.org",
			AuthToken: "token",
		}}},
		VCS: infro.VCSConfig{Value: &infro.GithubConfig{
			AuthToken: "token",
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, e)
}

func TestNewExecutorFromConfig_Invalid(t *testing.T) {
	testCases := []struct {
		name string
		cfg  infro.Config
	}{
		{
			name: "missing clients",
			cfg:  infro.Config{Deployers: []infro.DeployerConfig{}},
		},
		{
			name: "missing client fields",
			cfg: infro.Config{
				Deployers: []infro.DeployerConfig{{Value: infro.ArgoCDConfig{}}},
				VCS:       infro.VCSConfig{Value: infro.GithubConfig{AuthToken: "token"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := infro.NewExecutorFromConfig(&tc.cfg)
			require.Error(t, err)
		})
	}
}
