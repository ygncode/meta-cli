package cmd_impl

import (
	"fmt"

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
	postsCmd.AddCommand(postDeleteCmd())
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
	var message, photo, link string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new post",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := posts.New(rctx.Client)
			ctx := cmd.Context()

			var result *posts.CreateResult
			switch {
			case photo != "":
				result, err = svc.CreatePhoto(ctx, rctx.PageID, message, photo)
			case link != "":
				result, err = svc.CreateLink(ctx, rctx.PageID, message, link)
			case message != "":
				result, err = svc.CreateText(ctx, rctx.PageID, message)
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
	cmd.Flags().StringVar(&photo, "photo", "", "Path to photo file")
	cmd.Flags().StringVar(&link, "link", "", "URL to share")
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
