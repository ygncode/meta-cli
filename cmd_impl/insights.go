package cmd_impl

import (
	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/insights"
)

func init() {
	insightCmd := &cobra.Command{
		Use:     "insight",
		Aliases: []string{"insights"},
		Short:   "Page and post analytics",
	}

	insightCmd.AddCommand(insightPageCmd())
	insightCmd.AddCommand(insightPostCmd())
	rootCmd.AddCommand(insightCmd)
}

func insightPageCmd() *cobra.Command {
	var metric string
	var period string

	cmd := &cobra.Command{
		Use:   "page",
		Short: "Show page-level insights",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := insights.New(rctx.Client)
			metrics, err := svc.GetPageInsights(cmd.Context(), rctx.PageID, metric, period)
			if err != nil {
				return err
			}

			rows := insights.Flatten(metrics)
			if len(rows) == 0 {
				rctx.Printer.OK("No insights data available")
				return nil
			}
			return rctx.Printer.Print(rows)
		},
	}

	cmd.Flags().StringVar(&metric, "metric", "page_impressions,page_impressions_unique,page_engaged_users,page_post_engagements,page_views_total", "Comma-separated metrics")
	cmd.Flags().StringVar(&period, "period", "day", "Metric period (day, week, days_28, month, lifetime)")
	return cmd
}

func insightPostCmd() *cobra.Command {
	var metric string

	cmd := &cobra.Command{
		Use:   "post <post-id>",
		Short: "Show post-level insights",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := insights.New(rctx.Client)
			metrics, err := svc.GetPostInsights(cmd.Context(), args[0], metric)
			if err != nil {
				return err
			}

			rows := insights.Flatten(metrics)
			if len(rows) == 0 {
				rctx.Printer.OK("No insights data available")
				return nil
			}
			return rctx.Printer.Print(rows)
		},
	}

	cmd.Flags().StringVar(&metric, "metric", "post_impressions,post_impressions_unique,post_engaged_users,post_clicks", "Comma-separated metrics")
	return cmd
}
