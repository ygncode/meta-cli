package cmd_impl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/labels"
)

func init() {
	labelsCmd := &cobra.Command{
		Use:     "label",
		Aliases: []string{"labels"},
		Short:   "Manage page custom labels",
	}

	labelsCmd.AddCommand(labelListCmd())
	labelsCmd.AddCommand(labelCreateCmd())
	labelsCmd.AddCommand(labelDeleteCmd())
	labelsCmd.AddCommand(labelAssignCmd())
	labelsCmd.AddCommand(labelRemoveCmd())
	labelsCmd.AddCommand(labelListByUserCmd())
	rootCmd.AddCommand(labelsCmd)
}

func labelListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all custom labels for the page",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := labels.New(rctx.Client)
			list, err := svc.List(cmd.Context(), rctx.PageID)
			if err != nil {
				return err
			}

			if len(list) == 0 {
				rctx.Printer.OK("No labels found")
				return nil
			}
			return rctx.Printer.Print(list)
		},
	}
}

func labelCreateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new custom label",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			if name == "" {
				return fmt.Errorf("--name is required")
			}

			svc := labels.New(rctx.Client)
			id, err := svc.Create(cmd.Context(), rctx.PageID, name)
			if err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Label created: %s", id))
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Label name")
	return cmd
}

func labelDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <label-id>",
		Short: "Delete a custom label",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := labels.New(rctx.Client)
			if err := svc.Delete(cmd.Context(), args[0]); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Deleted label %s", args[0]))
			return nil
		},
	}
}

func labelAssignCmd() *cobra.Command {
	var psid string

	cmd := &cobra.Command{
		Use:   "assign <label-id>",
		Short: "Assign a label to a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			if psid == "" {
				return fmt.Errorf("--psid is required")
			}

			svc := labels.New(rctx.Client)
			if err := svc.Assign(cmd.Context(), args[0], psid); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Label %s assigned to user %s", args[0], psid))
			return nil
		},
	}

	cmd.Flags().StringVar(&psid, "psid", "", "Page-scoped user ID")
	return cmd
}

func labelRemoveCmd() *cobra.Command {
	var psid string

	cmd := &cobra.Command{
		Use:   "remove <label-id>",
		Short: "Remove a label from a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			if psid == "" {
				return fmt.Errorf("--psid is required")
			}

			svc := labels.New(rctx.Client)
			if err := svc.Remove(cmd.Context(), args[0], psid); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Label %s removed from user %s", args[0], psid))
			return nil
		},
	}

	cmd.Flags().StringVar(&psid, "psid", "", "Page-scoped user ID")
	return cmd
}

func labelListByUserCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-by-user <psid>",
		Short: "List labels assigned to a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := labels.New(rctx.Client)
			list, err := svc.ListByUser(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if len(list) == 0 {
				rctx.Printer.OK("No labels found for this user")
				return nil
			}
			return rctx.Printer.Print(list)
		},
	}
}
