package cmd_impl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/leads"
)

func init() {
	leadsCmd := &cobra.Command{
		Use:     "lead",
		Aliases: []string{"leads"},
		Short:   "Manage lead generation forms and leads",
	}

	leadsCmd.AddCommand(leadCreateFormCmd())
	leadsCmd.AddCommand(leadListCmd())
	rootCmd.AddCommand(leadsCmd)
}

func leadCreateFormCmd() *cobra.Command {
	var jsonStr, filePath string

	cmd := &cobra.Command{
		Use:   "create-form",
		Short: "Create a lead generation form",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			payload, err := readJSONInput(jsonStr, filePath)
			if err != nil {
				return err
			}

			svc := leads.New(rctx.Client)
			id, err := svc.CreateForm(cmd.Context(), rctx.PageID, payload)
			if err != nil {
				return err
			}

			rctx.Printer.OK(fmt.Sprintf("Form created: %s", id))
			return nil
		},
	}

	cmd.Flags().StringVar(&jsonStr, "json", "", "Form definition as JSON string")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON file with form definition")
	return cmd
}

func leadListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list <form-id>",
		Short: "List leads from a form",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := leads.New(rctx.Client)
			list, err := svc.ListLeads(cmd.Context(), args[0], limit)
			if err != nil {
				return err
			}

			if len(list) == 0 {
				rctx.Printer.OK("No leads found")
				return nil
			}
			return rctx.Printer.Print(list)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "Number of leads to fetch")
	return cmd
}
