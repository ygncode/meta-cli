package cmd_impl

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/messenger"
)

func init() {
	messengerCmd := &cobra.Command{
		Use:   "messenger",
		Short: "Manage Messenger messages",
	}

	messengerCmd.AddCommand(messengerSendCmd())
	messengerCmd.AddCommand(messengerSendTemplateCmd())
	messengerCmd.AddCommand(messengerListCmd())
	messengerCmd.AddCommand(messengerHistoryCmd())
	messengerCmd.AddCommand(messengerConversationsCmd())
	messengerCmd.AddCommand(messengerProfileCmd())
	rootCmd.AddCommand(messengerCmd)
}

func messengerSendCmd() *cobra.Command {
	var psid, message, tag string
	var image, video, audio, file string
	var quickReplies []string

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

			svc := messenger.NewService(rctx.Client)
			ctx := cmd.Context()

			// Determine attachment type
			attachType, attachValue := "", ""
			switch {
			case image != "":
				attachType, attachValue = "image", image
			case video != "":
				attachType, attachValue = "video", video
			case audio != "":
				attachType, attachValue = "audio", audio
			case file != "":
				attachType, attachValue = "file", file
			}

			var mid string
			var storeText string

			switch {
			case attachType != "":
				// Attachment mode
				if strings.HasPrefix(attachValue, "http://") || strings.HasPrefix(attachValue, "https://") {
					mid, err = svc.SendAttachmentURL(ctx, psid, attachType, attachValue)
					storeText = fmt.Sprintf("[%s] %s", attachType, attachValue)
				} else {
					mid, err = svc.SendAttachmentFile(ctx, psid, attachType, attachValue)
					storeText = fmt.Sprintf("[%s] %s", attachType, filepath.Base(attachValue))
				}
			case tag != "":
				if message == "" {
					return fmt.Errorf("--message is required with --tag")
				}
				mid, err = svc.SendTagged(ctx, psid, message, tag)
				storeText = message
			case len(quickReplies) > 0:
				if message == "" {
					return fmt.Errorf("--message is required with --quick-reply")
				}
				mid, err = svc.SendWithQuickReplies(ctx, psid, message, quickReplies)
				storeText = message
			default:
				if message == "" {
					return fmt.Errorf("--message is required")
				}
				mid, err = svc.Send(ctx, psid, message)
				storeText = message
			}

			if err != nil {
				return err
			}

			dbPath := rctx.Config.DBPath
			if dbPath == "" {
				var pathErr error
				dbPath, pathErr = messenger.DefaultDBPath()
				if pathErr != nil {
					return pathErr
				}
			}

			store, err := messenger.OpenStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			if err := store.SaveMessage(&messenger.Message{
				ID:         mid,
				PSID:       psid,
				PageID:     rctx.PageID,
				Text:       storeText,
				Direction:  "out",
				ReceivedAt: time.Now(),
			}); err != nil {
				return fmt.Errorf("save sent message: %w", err)
			}

			rctx.Printer.OK(fmt.Sprintf("Message sent to %s", psid))
			return nil
		},
	}

	cmd.Flags().StringVar(&psid, "psid", "", "Page-scoped user ID")
	cmd.Flags().StringVarP(&message, "message", "m", "", "Message text")
	cmd.Flags().StringVar(&image, "image", "", "Image URL or local file path")
	cmd.Flags().StringVar(&video, "video", "", "Video URL or local file path")
	cmd.Flags().StringVar(&audio, "audio", "", "Audio URL or local file path")
	cmd.Flags().StringVar(&file, "file", "", "File URL or local file path")
	cmd.Flags().StringVar(&tag, "tag", "", "Message tag (HUMAN_AGENT, ACCOUNT_UPDATE, POST_PURCHASE_UPDATE, CONFIRMED_EVENT_UPDATE)")
	cmd.Flags().StringArrayVar(&quickReplies, "quick-reply", nil, "Quick reply option (repeatable)")
	return cmd
}

func messengerSendTemplateCmd() *cobra.Command {
	var psid, jsonStr, filePath string

	cmd := &cobra.Command{
		Use:   "send-template",
		Short: "Send a template message",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			if psid == "" {
				return fmt.Errorf("--psid is required")
			}

			payload, err := readJSONInput(jsonStr, filePath)
			if err != nil {
				return err
			}

			svc := messenger.NewService(rctx.Client)
			mid, err := svc.SendTemplate(cmd.Context(), psid, payload)
			if err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Template sent: %s", mid))
			return nil
		},
	}

	cmd.Flags().StringVar(&psid, "psid", "", "Page-scoped user ID")
	cmd.Flags().StringVar(&jsonStr, "json", "", "Template payload as JSON string")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON file with template payload")
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

func messengerHistoryCmd() *cobra.Command {
	var psid string
	var limit int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "List conversation history with a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx := GetCtx(cmd)

			if rctx.PageID == "" {
				return fmt.Errorf("no page specified, use --page or set default_page in config")
			}
			if psid == "" {
				return fmt.Errorf("--psid is required")
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

			msgs, err := store.RecentMessages(rctx.PageID, psid, limit)
			if err != nil {
				return err
			}

			return rctx.Printer.Print(msgs)
		},
	}

	cmd.Flags().StringVar(&psid, "psid", "", "Page-scoped user ID")
	cmd.Flags().IntVar(&limit, "limit", 20, "Number of messages to fetch")
	return cmd
}

func messengerConversationsCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "conversations",
		Short: "List Messenger conversations from API",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := messenger.NewService(rctx.Client)
			list, err := svc.ListConversations(cmd.Context(), rctx.PageID, limit)
			if err != nil {
				return err
			}

			if len(list) == 0 {
				rctx.Printer.OK("No conversations found")
				return nil
			}
			return rctx.Printer.Print(list)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 25, "Number of conversations to fetch")
	return cmd
}

func messengerProfileCmd() *cobra.Command {
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage Messenger profile settings",
	}

	profileCmd.AddCommand(messengerProfileGetCmd())
	profileCmd.AddCommand(messengerProfileSetGreetingCmd())
	profileCmd.AddCommand(messengerProfileSetGetStartedCmd())
	profileCmd.AddCommand(messengerProfileSetMenuCmd())
	profileCmd.AddCommand(messengerProfileSetIceBreakersCmd())
	profileCmd.AddCommand(messengerProfileDeleteCmd())

	return profileCmd
}

func messengerProfileGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get current Messenger profile settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := messenger.NewService(rctx.Client)
			result, err := svc.GetProfile(cmd.Context())
			if err != nil {
				return err
			}

			return rctx.Printer.PrintOne(result)
		},
	}
}

func messengerProfileSetGreetingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-greeting <text>",
		Short: "Set the Messenger greeting text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := messenger.NewService(rctx.Client)
			if err := svc.SetGreeting(cmd.Context(), args[0]); err != nil {
				return err
			}

			rctx.Printer.OK("Greeting updated")
			return nil
		},
	}
}

func messengerProfileSetGetStartedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-get-started <payload>",
		Short: "Set the Get Started button payload",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := messenger.NewService(rctx.Client)
			if err := svc.SetGetStarted(cmd.Context(), args[0]); err != nil {
				return err
			}

			rctx.Printer.OK("Get Started button updated")
			return nil
		},
	}
}

func messengerProfileSetMenuCmd() *cobra.Command {
	var jsonStr, filePath string

	cmd := &cobra.Command{
		Use:   "set-menu",
		Short: "Set the persistent menu",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			payload, err := readJSONInput(jsonStr, filePath)
			if err != nil {
				return err
			}

			svc := messenger.NewService(rctx.Client)
			if err := svc.SetMenu(cmd.Context(), payload); err != nil {
				return err
			}

			rctx.Printer.OK("Persistent menu updated")
			return nil
		},
	}

	cmd.Flags().StringVar(&jsonStr, "json", "", "Menu definition as JSON string")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON file with menu definition")
	return cmd
}

func messengerProfileSetIceBreakersCmd() *cobra.Command {
	var jsonStr, filePath string

	cmd := &cobra.Command{
		Use:   "set-ice-breakers",
		Short: "Set ice breaker conversation starters",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			payload, err := readJSONInput(jsonStr, filePath)
			if err != nil {
				return err
			}

			svc := messenger.NewService(rctx.Client)
			if err := svc.SetIceBreakers(cmd.Context(), payload); err != nil {
				return err
			}

			rctx.Printer.OK("Ice breakers updated")
			return nil
		},
	}

	cmd.Flags().StringVar(&jsonStr, "json", "", "Ice breakers as JSON string")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON file with ice breakers")
	return cmd
}

func messengerProfileDeleteCmd() *cobra.Command {
	var field string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Messenger profile field",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			if field == "" {
				return fmt.Errorf("--field is required")
			}

			svc := messenger.NewService(rctx.Client)
			if err := svc.DeleteProfileField(cmd.Context(), field); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Deleted profile field: %s", field))
			return nil
		},
	}

	cmd.Flags().StringVar(&field, "field", "", "Profile field to delete (greeting, get_started, persistent_menu, ice_breakers)")
	return cmd
}
