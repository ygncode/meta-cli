package cmd_impl

import (
	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/graph"
	"github.com/ygncode/meta-cli/internal/pages"
)

func init() {
	pagesCmd := &cobra.Command{
		Use:   "pages",
		Short: "Manage Facebook Pages",
	}

	pagesCmd.AddCommand(pagesListCmd())
	rootCmd.AddCommand(pagesCmd)
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
