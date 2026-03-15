package cmd_impl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/posts"
)

func init() {
	reelCmd := &cobra.Command{
		Use:     "reel",
		Aliases: []string{"reels"},
		Short:   "Manage page reels",
	}
	reelCmd.AddCommand(reelCreateCmd())
	rootCmd.AddCommand(reelCmd)
}

func reelCreateCmd() *cobra.Command {
	var video, message, title, schedule, tz string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Publish a reel (short-form video)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if video == "" {
				return fmt.Errorf("--video is required")
			}

			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			var schedOpts *posts.ScheduleOpts
			if schedule != "" {
				t, err := parseScheduleTime(schedule, tz)
				if err != nil {
					return fmt.Errorf("invalid --schedule value: %w", err)
				}
				schedOpts = &posts.ScheduleOpts{PublishTime: t}
			}

			svc := posts.New(rctx.Client)
			result, err := svc.CreateReel(cmd.Context(), rctx.PageID, posts.ReelOpts{
				FilePath: video,
				Title:    title,
				Message:  message,
			}, schedOpts)
			if err != nil {
				return err
			}

			return rctx.Printer.PrintOne(result)
		},
	}

	cmd.Flags().StringVar(&video, "video", "", "Path to video file (required)")
	cmd.Flags().StringVar(&message, "message", "", "Reel description")
	cmd.Flags().StringVar(&title, "title", "", "Reel title")
	cmd.Flags().StringVar(&schedule, "schedule", "", "Schedule for future publishing (e.g. \"2026-03-20 14:00\")")
	cmd.Flags().StringVar(&tz, "tz", "", "Timezone for --schedule (e.g. \"Asia/Yangon\"), defaults to local")
	return cmd
}
