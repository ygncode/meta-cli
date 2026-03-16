package cmd_impl

import (
	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/ratings"
)

func init() {
	ratingsCmd := &cobra.Command{
		Use:     "rating",
		Aliases: []string{"ratings"},
		Short:   "View page ratings and reviews",
	}

	ratingsCmd.AddCommand(ratingListCmd())
	ratingsCmd.AddCommand(ratingSummaryCmd())
	rootCmd.AddCommand(ratingsCmd)
}

func ratingListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List page ratings and reviews",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := ratings.New(rctx.Client)
			list, err := svc.List(cmd.Context(), rctx.PageID, limit)
			if err != nil {
				return err
			}

			if len(list) == 0 {
				rctx.Printer.OK("No ratings found")
				return nil
			}
			return rctx.Printer.Print(list)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 25, "Number of ratings to fetch")
	return cmd
}

func ratingSummaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Show overall page rating",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := ratings.New(rctx.Client)
			summary, err := svc.Summary(cmd.Context(), rctx.PageID)
			if err != nil {
				return err
			}

			return rctx.Printer.PrintOne(summary)
		},
	}
}
