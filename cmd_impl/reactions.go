package cmd_impl

import (
	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/reactions"
)

func init() {
	reactionsCmd := &cobra.Command{
		Use:     "reaction",
		Aliases: []string{"reactions"},
		Short:   "View reactions on posts and comments",
	}

	reactionsCmd.AddCommand(reactionListCmd())
	rootCmd.AddCommand(reactionsCmd)
}

func reactionListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list <object-id>",
		Short: "List reactions on a post or comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := reactions.New(rctx.Client)
			list, err := svc.List(cmd.Context(), args[0], limit)
			if err != nil {
				return err
			}

			if len(list) == 0 {
				rctx.Printer.OK("No reactions found")
				return nil
			}
			return rctx.Printer.Print(list)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "Number of reactions to fetch")
	return cmd
}
