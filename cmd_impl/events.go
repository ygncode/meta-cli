package cmd_impl

import (
	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/events"
)

func init() {
	eventsCmd := &cobra.Command{
		Use:     "event",
		Aliases: []string{"events"},
		Short:   "Manage page events",
	}

	eventsCmd.AddCommand(eventListCmd())
	rootCmd.AddCommand(eventsCmd)
}

func eventListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List page events",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := events.New(rctx.Client)
			list, err := svc.List(cmd.Context(), rctx.PageID, limit)
			if err != nil {
				return err
			}

			if len(list) == 0 {
				rctx.Printer.OK("No events found")
				return nil
			}
			return rctx.Printer.Print(list)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 25, "Number of events to fetch")
	return cmd
}
