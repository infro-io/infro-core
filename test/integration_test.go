package test_test

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"path/filepath"
	"testing"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/session"
	"github.com/argoproj/argo-cd/v2/util/io"
	"github.com/google/go-github/v57/github"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"infro.io/infro-core/cmd/comment"
	"infro.io/infro-core/internal/model"
)

const argoAddr = "localhost:8080"

var configTemplate, _ = template.New("config").Parse(`
deployers:
  - type: argocd
    name: my-argo-managed-cluster
    endpoint: {{.argoAddr}}
    authtoken: {{.argoAuthToken}}
  - type: terraform
    name: my-tf-managed-infra
    workdir: ./terraform
vcs:
  type: github
  authtoken: {{.githubAuthToken}}
`)

func TestCommentDiffs(t *testing.T) {
	// assemble
	owner := "infro-io"
	repo := "example-helm"
	ctx := context.Background()
	cfg := new(bytes.Buffer)
	githubAuthToken := getGithubAuthTokenOrDie()
	configTemplate.Execute(cfg, map[string]string{
		"argoAddr":        argoAddr,
		"argoAuthToken":   getArgoAuthTokenOrDie(ctx),
		"githubAuthToken": githubAuthToken,
	})
	cmd := comment.NewDiffsCommand()
	cmd.SetArgs([]string{
		"--repo", owner + "/" + repo,
		"--revision", "8d508b82cc188ce7c8244bfc280f24d75b4a1e7e",
		"--config", cfg.String(),
		"--pull-number", "1",
	})
	out := new(bytes.Buffer)
	cmd.SetOut(out)

	// act
	err := cmd.ExecuteContext(ctx)

	// assert
	require.NoError(t, err)
	var cmtMet model.CommentMetadata
	err = json.Unmarshal(out.Bytes(), &cmtMet)
	require.NoError(t, err)
	cmt := getGithubCommentOrDie(ctx, githubAuthToken, owner, repo, cmtMet.CommentID)
	require.Contains(t, *cmt.Body, "my_heart_is: full")
}

func getArgoAuthTokenOrDie(ctx context.Context) string {
	kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}
	kubeClient := kubernetes.NewForConfigOrDie(config)
	secret, err := kubeClient.CoreV1().Secrets("argocd").Get(ctx, "argocd-initial-admin-secret", v1.GetOptions{})
	if err != nil {
		panic(err)
	}
	argoPassword := string(secret.Data["password"])
	apiClient, err := apiclient.NewClient(&apiclient.ClientOptions{
		ServerAddr: argoAddr,
		Insecure:   true,
	})
	if err != nil {
		panic(err)
	}
	sessConn, sessClient := apiClient.NewSessionClientOrDie()
	defer io.Close(sessConn)
	sess, err := sessClient.Create(ctx, &session.SessionCreateRequest{
		Username: "admin",
		Password: argoPassword,
	})
	if err != nil {
		panic(err)
	}
	return sess.Token
}

func getGithubAuthTokenOrDie() string {
	ght := os.Getenv("GITHUB_TOKEN")
	if ght == "" {
		panic("github token not found")
	}
	return ght
}

func getGithubCommentOrDie(ctx context.Context, authToken string, owner string, repo string, commentID int64) *github.IssueComment {
	ghClient := github.NewClient(nil).WithAuthToken(authToken)
	cmt, _, err := ghClient.Issues.GetComment(ctx, owner, repo, commentID)
	if err != nil {
		panic(err)
	}
	return cmt
}
