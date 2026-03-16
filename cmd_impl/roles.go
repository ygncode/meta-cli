package cmd_impl

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/roles"
)

func init() {
	rolesCmd := &cobra.Command{
		Use:     "role",
		Aliases: []string{"roles"},
		Short:   "Manage page user roles",
	}

	rolesCmd.AddCommand(roleListCmd())
	rolesCmd.AddCommand(roleAssignCmd())
	rolesCmd.AddCommand(roleRemoveCmd())
	rootCmd.AddCommand(rolesCmd)
}

func roleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List users with page access",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := roles.New(rctx.Client)
			list, err := svc.List(cmd.Context(), rctx.PageID)
			if err != nil {
				return err
			}

			if len(list) == 0 {
				rctx.Printer.OK("No assigned users")
				return nil
			}
			return rctx.Printer.Print(list)
		},
	}
}

func roleAssignCmd() *cobra.Command {
	var tasks string

	cmd := &cobra.Command{
		Use:   "assign <user-id>",
		Short: "Assign roles to a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			if tasks == "" {
				return fmt.Errorf("--tasks is required")
			}

			taskList := strings.Split(tasks, ",")
			svc := roles.New(rctx.Client)
			if err := svc.Assign(cmd.Context(), rctx.PageID, args[0], taskList); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Assigned user %s with tasks: %s", args[0], tasks))
			return nil
		},
	}

	cmd.Flags().StringVar(&tasks, "tasks", "", "Comma-separated tasks (MANAGE,CREATE_CONTENT,MODERATE,MESSAGING,ADVERTISE,ANALYZE)")
	return cmd
}

func roleRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <user-id>",
		Short: "Remove a user's page access",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := roles.New(rctx.Client)
			if err := svc.Remove(cmd.Context(), rctx.PageID, args[0]); err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Removed user %s", args[0]))
			return nil
		},
	}
}
