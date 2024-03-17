package poll

import (
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"infro.io/infro-core/pkg/infro"
)

func NewCommand() *cobra.Command {
	var config string
	var owner string
	var interval string
	cmd := &cobra.Command{
		Use:   "poll",
		Short: "Poll pull requests",
		Long:  "Poll pull requests for new changes and comment diffs for all deployers",
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg infro.Config
			if err := yaml.Unmarshal([]byte(config), &cfg); err != nil {
				return err
			}
			ex, err := infro.NewExecutorFromConfig(&cfg)
			if err != nil {
				return err
			}
			pollInt, err := time.ParseDuration(interval)
			if err != nil {
				return err
			}
			ex.Poll(cmd.Context(), owner, pollInt)
			return nil
		},
	}
	cmd.Flags().StringVar(&config, "config", "", "The config as a string.")
	cmd.Flags().StringVar(&owner, "owner", "", "The org or user.")
	cmd.Flags().StringVar(&interval, "interval", "10s", "The interval.")
	return cmd
}
