package comment

import (
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/json"

	"infro.io/infro-core/pkg/infro"
)

func NewDiffsCommand() *cobra.Command {
	var config string
	var repo string
	var revision string
	var pullNumber int
	cmd := &cobra.Command{
		Use:   "comment diffs",
		Short: "Comment diffs",
		Long:  "Perform a diff on the configured deployers and comment to a VCS",
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg infro.Config
			if err := yaml.Unmarshal([]byte(config), &cfg); err != nil {
				return err
			}
			ex, err := infro.NewExecutorFromConfig(&cfg)
			if err != nil {
				return err
			}
			ownerAndRepo := strings.Split(repo, "/")
			comment, err := ex.CommentDiffs(cmd.Context(), infro.CommentDiffOpts{
				Revision:   revision,
				Owner:      ownerAndRepo[0],
				Repo:       ownerAndRepo[1],
				PullNumber: pullNumber,
			})
			if err != nil {
				return err
			}
			output, err := json.Marshal(comment)
			cmd.Print(string(output))
			return err
		},
	}
	cmd.Flags().StringVar(&config, "config", "", "The config as a string.")
	cmd.Flags().StringVar(&repo, "repo", "", "The repository in the form owner/repo.")
	cmd.Flags().StringVar(&revision, "revision", "", "The commit SHA.")
	cmd.Flags().IntVar(&pullNumber, "pull-number", 0, "The pull request number.")
	return cmd
}
