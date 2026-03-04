package cmd_impl

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/auth"
	"github.com/ygncode/meta-cli/internal/config"
)

func init() {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}

	authCmd.AddCommand(authLoginCmd())
	authCmd.AddCommand(authStatusCmd())
	authCmd.AddCommand(authRefreshCmd())
	rootCmd.AddCommand(authCmd)
}

func authLoginCmd() *cobra.Command {
	var appID, appSecret string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login with Facebook OAuth",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx := GetCtx(cmd)

			if appID == "" {
				if acct, ok := rctx.Config.Accounts[rctx.Account]; ok && acct.AppID != "" {
					appID = acct.AppID
				}
			}
			if appID == "" {
				return fmt.Errorf("--app-id is required (or set in config)")
			}
			if appSecret == "" {
				return fmt.Errorf("--app-secret is required")
			}

			if err := rctx.Store.SaveSecret(rctx.Account, appSecret); err != nil {
				return fmt.Errorf("save app secret: %w", err)
			}

			loginURL := auth.LoginURL(appID, rctx.Config.GraphAPIVersion)
			fmt.Fprintf(os.Stderr, "Open this URL in your browser:\n\n  %s\n\n", loginURL)
			fmt.Fprint(os.Stderr, "Paste the redirect URL here: ")

			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("read input: %w", err)
			}

			code, err := auth.ExtractCode(strings.TrimSpace(line))
			if err != nil {
				return err
			}

			ctx := cmd.Context()

			shortToken, err := auth.ExchangeCode(ctx, code, appID, appSecret, rctx.Config.GraphAPIVersion)
			if err != nil {
				return fmt.Errorf("exchange code: %w", err)
			}

			longToken, err := auth.ExtendToken(ctx, shortToken, appID, appSecret, rctx.Config.GraphAPIVersion)
			if err != nil {
				return fmt.Errorf("extend token: %w", err)
			}

			pages, err := auth.FetchPageTokens(ctx, longToken, rctx.Config.GraphAPIVersion)
			if err != nil {
				return fmt.Errorf("fetch page tokens: %w", err)
			}

			tokens := &auth.Tokens{
				UserToken: longToken,
				Pages:     pages,
			}

			if err := rctx.Store.SaveTokens(rctx.Account, tokens); err != nil {
				return fmt.Errorf("save tokens: %w", err)
			}

			if rctx.Config.Accounts == nil {
				rctx.Config.Accounts = make(map[string]config.Account)
			}
			rctx.Config.Accounts[rctx.Account] = config.Account{AppID: appID}

			fmt.Fprintf(os.Stderr, "Logged in. Found %d page(s):\n", len(pages))
			for id, pt := range pages {
				fmt.Fprintf(os.Stderr, "  %s  %s\n", id, pt.Name)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "Facebook App ID")
	cmd.Flags().StringVar(&appSecret, "app-secret", "", "Facebook App Secret")
	return cmd
}

func authStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current auth status",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx := GetCtx(cmd)

			tokens, err := rctx.Store.GetTokens(rctx.Account)
			if err != nil {
				rctx.Printer.Err(fmt.Errorf("not logged in"))
				return nil
			}

			type pageInfo struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}

			var pages []pageInfo
			for id, name := range tokens.PageNames() {
				pages = append(pages, pageInfo{ID: id, Name: name})
			}

			if flagJSON {
				return rctx.Printer.Print(pages)
			}

			fmt.Fprintf(os.Stderr, "Account: %s\n", rctx.Account)
			fmt.Fprintf(os.Stderr, "Pages: %d\n", len(pages))
			for _, p := range pages {
				marker := "  "
				if p.ID == rctx.PageID {
					marker = "* "
				}
				fmt.Fprintf(os.Stderr, "  %s%s  %s\n", marker, p.ID, p.Name)
			}
			return nil
		},
	}
}

func authRefreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: "Refresh page tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requireTokens(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()

			pages, err := auth.FetchPageTokens(ctx, rctx.Tokens.UserToken, rctx.Config.GraphAPIVersion)
			if err != nil {
				return fmt.Errorf("refresh page tokens: %w", err)
			}

			rctx.Tokens.Pages = pages
			if err := rctx.Store.SaveTokens(rctx.Account, rctx.Tokens); err != nil {
				return fmt.Errorf("save tokens: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Refreshed %d page token(s)\n", len(pages))
			return nil
		},
	}
}
