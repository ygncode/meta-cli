package cmd_impl

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/posts"
)

func init() {
	postsCmd := &cobra.Command{
		Use:     "post",
		Aliases: []string{"posts"},
		Short:   "Manage page posts",
	}

	postsCmd.AddCommand(postListCmd())
	postsCmd.AddCommand(postCreateCmd())
	postsCmd.AddCommand(postUpdateCmd())
	postsCmd.AddCommand(postEditCmd())
	postsCmd.AddCommand(postDeleteCmd())
	postsCmd.AddCommand(postListScheduledCmd())
	rootCmd.AddCommand(postsCmd)
}

func postListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent posts",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := posts.New(rctx.Client)
			list, err := svc.List(cmd.Context(), rctx.PageID, limit)
			if err != nil {
				return err
			}

			return rctx.Printer.Print(list)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Number of posts to fetch")
	return cmd
}

func postCreateCmd() *cobra.Command {
	var message, link, schedule, tz string
	var photos []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new post",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			ctx := cmd.Context()

			var result *posts.CreateResult
			switch {
			case len(photos) > 1:
				result, err = svc.CreatePhotos(ctx, rctx.PageID, message, photos, schedOpts)
			case len(photos) == 1:
				result, err = svc.CreatePhoto(ctx, rctx.PageID, message, photos[0], schedOpts)
			case link != "":
				result, err = svc.CreateLink(ctx, rctx.PageID, message, link, schedOpts)
			case message != "":
				result, err = svc.CreateText(ctx, rctx.PageID, message, schedOpts)
			default:
				return fmt.Errorf("provide --message, --photo, or --link")
			}

			if err != nil {
				return err
			}

			return rctx.Printer.PrintOne(result)
		},
	}

	cmd.Flags().StringVar(&message, "message", "", "Post message text")
	cmd.Flags().StringArrayVar(&photos, "photo", nil, "Path to photo file (repeatable for multiple images)")
	cmd.Flags().StringVar(&link, "link", "", "URL to share")
	cmd.Flags().StringVar(&schedule, "schedule", "", "Schedule post for future publishing (e.g. \"2026-03-20 14:00\")")
	cmd.Flags().StringVar(&tz, "tz", "", "Timezone for --schedule (e.g. \"Asia/Yangon\"), defaults to local")
	return cmd
}

func postUpdateCmd() *cobra.Command {
	return postModifyCmd("update", "Update a post's message")
}

func postEditCmd() *cobra.Command {
	return postModifyCmd("edit", "Edit a post's message")
}

func postModifyCmd(use, short string) *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   use + " <post-id>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			if message == "" {
				return fmt.Errorf("--message is required")
			}

			svc := posts.New(rctx.Client)
			if err := svc.Update(cmd.Context(), args[0], message); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Updated post %s", args[0]))
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "New message text")
	return cmd
}

func postDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <post-id>",
		Short: "Delete a post",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := posts.New(rctx.Client)
			if err := svc.Delete(cmd.Context(), args[0]); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Deleted post %s", args[0]))
			return nil
		},
	}
}

func postListScheduledCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list-scheduled",
		Short: "List scheduled (unpublished) posts",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := posts.New(rctx.Client)
			list, err := svc.ListScheduled(cmd.Context(), rctx.PageID, limit)
			if err != nil {
				return err
			}

			if len(list) == 0 {
				rctx.Printer.OK("No scheduled posts")
				return nil
			}

			return rctx.Printer.Print(list)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Number of scheduled posts to fetch")
	return cmd
}

func parseScheduleTime(datetime, tz string) (time.Time, error) {
	loc := time.Local
	if tz != "" {
		var err error
		loc, err = time.LoadLocation(tz)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid timezone %q: %w", tz, err)
		}
	}
	t, err := time.ParseInLocation("2006-01-02 15:04", datetime, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("expected format \"YYYY-MM-DD HH:MM\": %w", err)
	}
	return t, nil
}
