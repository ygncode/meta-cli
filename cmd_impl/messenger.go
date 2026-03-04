package cmd_impl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/messenger"
)

func init() {
	messengerCmd := &cobra.Command{
		Use:   "messenger",
		Short: "Manage Messenger messages",
	}

	messengerCmd.AddCommand(messengerSendCmd())
	messengerCmd.AddCommand(messengerListCmd())
	rootCmd.AddCommand(messengerCmd)
}

func messengerSendCmd() *cobra.Command {
	var psid, message string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a message to a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			if psid == "" {
				return fmt.Errorf("--psid is required")
			}
			if message == "" {
				return fmt.Errorf("--message is required")
			}

			svc := messenger.NewService(rctx.Client)
			if err := svc.Send(cmd.Context(), psid, message); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Message sent to %s", psid))
			return nil
		},
	}

	cmd.Flags().StringVar(&psid, "psid", "", "Page-scoped user ID")
	cmd.Flags().StringVarP(&message, "message", "m", "", "Message text")
	return cmd
}

func messengerListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stored messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx := GetCtx(cmd)

			if rctx.PageID == "" {
				return fmt.Errorf("no page specified, use --page or set default_page in config")
			}

			dbPath := rctx.Config.DBPath
			if dbPath == "" {
				var err error
				dbPath, err = messenger.DefaultDBPath()
				if err != nil {
					return err
				}
			}

			store, err := messenger.OpenStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			msgs, err := store.ListMessages(rctx.PageID, limit)
			if err != nil {
				return err
			}

			return rctx.Printer.Print(msgs)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "Number of messages to fetch")
	return cmd
}
