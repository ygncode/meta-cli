package cmd_impl

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/auth"
	"github.com/ygncode/meta-cli/internal/config"
	"github.com/ygncode/meta-cli/internal/graph"
	"github.com/ygncode/meta-cli/internal/output"
)

type RootCtx struct {
	Config  *config.Config
	Store   auth.Store
	Tokens  *auth.Tokens
	Client  *graph.Client
	Printer *output.Printer
	Account string
	PageID  string
}

type ctxKey struct{}

func GetCtx(cmd *cobra.Command) *RootCtx {
	v, ok := cmd.Context().Value(ctxKey{}).(*RootCtx)
	if !ok {
		return &RootCtx{}
	}
	return v
}

var (
	flagJSON    bool
	flagPlain   bool
	flagPage    string
	flagAccount string
)

var rootCmd = &cobra.Command{
	Use:   "meta-cli",
	Short: "CLI for managing Facebook Pages and Messenger",
	Long:  "A command-line tool for managing Facebook Pages, posts, comments, Messenger messages, and RAG-powered auto-replies.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		format := output.FormatTable
		if flagJSON {
			format = output.FormatJSON
		} else if flagPlain {
			format = output.FormatPlain
		}

		account := flagAccount
		if account == "" {
			account = cfg.DefaultAccount
		}

		pageID := flagPage
		if pageID == "" {
			pageID = cfg.DefaultPage
		}

		rctx := &RootCtx{
			Config:  cfg,
			Store:   auth.NewKeyringStore(),
			Printer: output.New(format, os.Stdout),
			Account: account,
			PageID:  pageID,
		}

		cmd.SetContext(withCtx(cmd.Context(), rctx))
		return nil
	},
}

func withCtx(parent context.Context, rctx *RootCtx) context.Context {
	return context.WithValue(parent, ctxKey{}, rctx)
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&flagPlain, "plain", false, "Output as TSV")
	rootCmd.PersistentFlags().StringVar(&flagPage, "page", "", "Page ID to operate on")
	rootCmd.PersistentFlags().StringVar(&flagAccount, "account", "", "Account name")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func requireTokens(cmd *cobra.Command) (*RootCtx, error) {
	rctx := GetCtx(cmd)
	tokens, err := rctx.Store.GetTokens(rctx.Account)
	if err != nil {
		return nil, fmt.Errorf("not logged in, run: meta-cli auth login")
	}
	rctx.Tokens = tokens
	return rctx, nil
}

func requirePageClient(cmd *cobra.Command) (*RootCtx, error) {
	rctx, err := requireTokens(cmd)
	if err != nil {
		return nil, err
	}

	if rctx.PageID == "" {
		return nil, fmt.Errorf("no page specified, use --page or set default_page in config")
	}

	pageToken, ok := rctx.Tokens.PageAccessToken(rctx.PageID)
	if !ok {
		return nil, fmt.Errorf("no token for page %s, run: meta-cli auth login", rctx.PageID)
	}

	rctx.Client = graph.New(rctx.Config.GraphAPIVersion, pageToken)
	return rctx, nil
}
