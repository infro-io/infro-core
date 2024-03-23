package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"

	gh "github.com/google/go-github/v57/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/infro-io/infro-core/internal/model"
	"github.com/infro-io/infro-core/internal/xzap"
)

type (
	Client struct {
		cfg *Config
	}
	Config struct {
		AuthToken string `validate:"required"`
	}
	queryErrors []struct {
		Message string
	}
)

func (e queryErrors) Error() string {
	return e[0].Message
}

const graphQLEndpoint = "https://api.github.com/graphql"

func NewClient(cfg *Config) *Client {
	return &Client{cfg}
}

func (c Client) RepoURL(owner string, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

//nolint:gochecknoglobals // fine for templates
var queryTemplate, _ = template.New("query").Parse(`
{
	search(query: "owner:{{.owner}} is:pr state:open updated:>={{.updated}}", type: ISSUE, first: 100) {
		nodes {
			... on PullRequest {
				number
				repository {
					name
				}
				commits(last: 1) {
					nodes {
						commit {
							oid
						}
					}
				}
			}
		}
	}
}`)

// ListPullRequests lists the recently updated pull requests which do not yet have a comment.
func (c Client) ListPullRequests(ctx context.Context, opts model.ListPullRequestsOpts) ([]model.PullRequest, error) {
	log := xzap.FromContext(ctx)
	token := &oauth2.Token{AccessToken: c.cfg.AuthToken}
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	buf := new(bytes.Buffer)
	updated := opts.UpdatedSince.Format("2006-01-02T15:04:05")
	if err := queryTemplate.Execute(buf, map[string]string{"owner": opts.Owner, "updated": updated}); err != nil {
		return nil, err
	}
	log.Info("finding pull requests", zap.String("updatedSince", updated))
	in := struct {
		Query string `json:"query"`
	}{
		Query: buf.String(),
	}
	q, _ := json.Marshal(in)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, graphQLEndpoint, bytes.NewBuffer(q))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-200 OK status code: %v body: %q", resp.Status, body)
	}

	var out struct {
		Data struct {
			Search struct {
				Nodes []struct {
					Repository struct {
						Name string
					}
					Number  int
					Commits struct {
						Nodes []struct {
							Commit struct {
								OID string
							}
						}
					}
				}
			}
		}
		Errors queryErrors
	}
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, err
	}
	if out.Errors != nil {
		return nil, out.Errors
	}
	nodes := out.Data.Search.Nodes
	pulls := make([]model.PullRequest, len(nodes))
	for i, pull := range nodes {
		pulls[i] = model.PullRequest{
			Repo:     pull.Repository.Name,
			Number:   pull.Number,
			Revision: pull.Commits.Nodes[0].Commit.OID,
		}
	}
	return pulls, nil
}

func (c Client) UpsertComment(ctx context.Context, opts model.UpsertCommentOpts) (*model.Comment, error) {
	log := xzap.FromContext(ctx)
	client := gh.NewClient(nil).WithAuthToken(c.cfg.AuthToken)
	comments, _, err := client.Issues.ListComments(ctx, opts.Owner, opts.Repo, opts.PullNumber, nil)
	if err != nil {
		return nil, err
	}
	comment := getCommentWithSubstring(comments, htmlTag)
	body := Format(opts.Revision, opts.DryRuns)
	*body += "\n" + htmlTag
	if comment == nil {
		log.Info("creating new comment")
		newComment := &gh.IssueComment{Body: body}
		comment, _, err = client.Issues.CreateComment(ctx, opts.Owner, opts.Repo, opts.PullNumber, newComment)
	} else {
		if *comment.Body == *body {
			log.Info("comment unchanged, skipping")
		} else {
			log.Info("updating comment")
			comment, _, err = client.Issues.EditComment(ctx, opts.Owner, opts.Repo, *comment.ID, &gh.IssueComment{Body: body})
		}
	}
	if err != nil {
		return nil, err
	}
	return &model.Comment{ID: *comment.ID, Body: body}, nil
}

func getCommentWithSubstring(comments []*gh.IssueComment, substring string) *gh.IssueComment {
	for _, cmt := range comments {
		if strings.Contains(*cmt.Body, substring) {
			return cmt
		}
	}
	return nil
}

const htmlTag = "<!-- INFRO -->"
