package cmd_impl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/blocked"
)

func init() {
	blockedCmd := &cobra.Command{
		Use:   "blocked",
		Short: "Manage blocked users",
	}

	blockedCmd.AddCommand(blockedListCmd())
	blockedCmd.AddCommand(blockedAddCmd())
	blockedCmd.AddCommand(blockedRemoveCmd())
	rootCmd.AddCommand(blockedCmd)
}

func blockedListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List blocked users",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := blocked.New(rctx.Client)
			list, err := svc.List(cmd.Context(), rctx.PageID, limit)
			if err != nil {
				return err
			}

			if len(list) == 0 {
				rctx.Printer.OK("No blocked users")
				return nil
			}
			return rctx.Printer.Print(list)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 25, "Number of blocked users to fetch")
	return cmd
}

func blockedAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <user-id>",
		Short: "Block a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := blocked.New(rctx.Client)
			if err := svc.Block(cmd.Context(), rctx.PageID, args[0]); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Blocked user %s", args[0]))
			return nil
		},
	}
}

func blockedRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <user-id>",
		Short: "Unblock a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := blocked.New(rctx.Client)
			if err := svc.Unblock(cmd.Context(), rctx.PageID, args[0]); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Unblocked user %s", args[0]))
			return nil
		},
	}
}
