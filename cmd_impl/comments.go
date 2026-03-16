package cmd_impl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/comments"
	"github.com/ygncode/meta-cli/internal/messenger"
)

func init() {
	commentsCmd := &cobra.Command{
		Use:     "comment",
		Aliases: []string{"comments"},
		Short:   "Manage post comments",
	}

	commentsCmd.AddCommand(commentListCmd())
	commentsCmd.AddCommand(commentReplyCmd())
	commentsCmd.AddCommand(commentUpdateCmd())
	commentsCmd.AddCommand(commentHideCmd())
	commentsCmd.AddCommand(commentUnhideCmd())
	commentsCmd.AddCommand(commentDeleteCmd())
	commentsCmd.AddCommand(commentPrivateReplyCmd())
	rootCmd.AddCommand(commentsCmd)
}

func commentListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list <post-id>",
		Short: "List comments on a post",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := comments.New(rctx.Client)
			list, err := svc.List(cmd.Context(), args[0], limit)
			if err != nil {
				return err
			}

			return rctx.Printer.Print(list)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 25, "Number of comments to fetch")
	return cmd
}

func commentReplyCmd() *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "reply <comment-id>",
		Short: "Reply to a comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			if message == "" {
				return fmt.Errorf("--message is required")
			}

			svc := comments.New(rctx.Client)
			id, err := svc.Reply(cmd.Context(), args[0], message)
			if err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Reply created: %s", id))
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Reply message")
	return cmd
}

func commentUpdateCmd() *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "update <comment-id>",
		Short: "Update a comment's message",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			if message == "" {
				return fmt.Errorf("--message is required")
			}

			svc := comments.New(rctx.Client)
			if err := svc.Update(cmd.Context(), args[0], message); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Updated comment %s", args[0]))
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "New message text")
	return cmd
}

func commentHideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hide <comment-id>",
		Short: "Hide a comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := comments.New(rctx.Client)
			if err := svc.SetHidden(cmd.Context(), args[0], true); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Comment %s hidden", args[0]))
			return nil
		},
	}
}

func commentUnhideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unhide <comment-id>",
		Short: "Unhide a comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := comments.New(rctx.Client)
			if err := svc.SetHidden(cmd.Context(), args[0], false); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Comment %s unhidden", args[0]))
			return nil
		},
	}
}

func commentDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <comment-id>",
		Short: "Delete a comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := comments.New(rctx.Client)
			if err := svc.Delete(cmd.Context(), args[0]); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Deleted comment %s", args[0]))
			return nil
		},
	}
}

func commentPrivateReplyCmd() *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "private-reply <comment-id>",
		Short: "Send a private Messenger reply to a comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			if message == "" {
				return fmt.Errorf("--message is required")
			}

			svc := messenger.NewService(rctx.Client)
			mid, err := svc.SendPrivateReply(cmd.Context(), args[0], message)
			if err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Private reply sent: %s", mid))
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Message text")
	return cmd
}
