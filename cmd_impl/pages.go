package cmd_impl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/config"
	"github.com/ygncode/meta-cli/internal/graph"
	"github.com/ygncode/meta-cli/internal/pages"
)

func init() {
	pagesCmd := &cobra.Command{
		Use:   "pages",
		Short: "Manage Facebook Pages",
	}

	pagesCmd.AddCommand(pagesListCmd())
	pagesCmd.AddCommand(pagesSetDefaultCmd())
	rootCmd.AddCommand(pagesCmd)
}

func pagesSetDefaultCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-default <page-id>",
		Short: "Set a default page for commands",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pageID := args[0]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			cfg.DefaultPage = pageID
			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Default page set to %s\n", pageID)
			return nil
		},
	}
}

func pagesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List pages you manage",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requireTokens(cmd)
			if err != nil {
				return err
			}

			client := graph.New(rctx.Config.GraphAPIVersion, rctx.Tokens.UserToken)
			svc := pages.New(client)

			list, err := svc.List(cmd.Context())
			if err != nil {
				return err
			}

			return rctx.Printer.Print(list)
		},
	}
}
