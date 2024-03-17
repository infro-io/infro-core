package root

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"infro.io/infro-core/cmd/comment"
	"infro.io/infro-core/cmd/poll"
	"infro.io/infro-core/internal/xzap"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "infro",
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log, err := zap.NewProduction()
			if err != nil {
				return err
			}
			ctx := xzap.NewContext(cmd.Context(), log)
			cmd.SetContext(ctx)
			return nil
		},
	}
	cmd.AddCommand(comment.NewCommand())
	cmd.AddCommand(poll.NewCommand())
	return cmd
}
