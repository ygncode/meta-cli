package cmd_impl

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/messenger"
	"github.com/ygncode/meta-cli/internal/rag"
)

func init() {
	webhookCmd := &cobra.Command{
		Use:   "webhook",
		Short: "Webhook server for real-time messages",
	}

	webhookCmd.AddCommand(webhookServeCmd())
	rootCmd.AddCommand(webhookCmd)
}

func webhookServeCmd() *cobra.Command {
	var port int
	var verifyToken string
	var ragDir string
	var ragThreshold float64

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the webhook HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			appSecret, err := rctx.Store.GetSecret(rctx.Account)
			if err != nil {
				return fmt.Errorf("app secret not found, run: meta-cli auth login")
			}

			if verifyToken == "" {
				verifyToken = os.Getenv("META_VERIFY_TOKEN")
			}
			if verifyToken == "" {
				return fmt.Errorf("--verify-token or META_VERIFY_TOKEN is required")
			}

			dbPath := rctx.Config.DBPath
			if dbPath == "" {
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

			handler := &messenger.WebhookHandler{
				VerifyToken:  verifyToken,
				AppSecret:    appSecret,
				PageID:       rctx.PageID,
				Store:        store,
				Messenger:    messenger.NewService(rctx.Client),
				RAGThreshold: ragThreshold,
			}

			if ragDir == "" {
				ragDir = rctx.Config.RAGDir
			}
			if ragDir != "" {
				docs, err := rag.LoadDir(ragDir)
				if err != nil {
					return fmt.Errorf("load RAG docs: %w", err)
				}
				handler.RAG = rag.Build(docs)
				log.Printf("RAG index loaded: %d documents from %s", len(docs), ragDir)
			}

			if port == 0 {
				port = rctx.Config.WebhookPort
			}

			addr := fmt.Sprintf(":%d", port)
			log.Printf("Webhook server listening on %s", addr)

			return http.ListenAndServe(addr, handler)
		},
	}

	cmd.Flags().IntVar(&port, "port", 0, "Port to listen on (default from config)")
	cmd.Flags().StringVar(&verifyToken, "verify-token", "", "Webhook verify token")
	cmd.Flags().StringVar(&ragDir, "rag-dir", "", "Directory with RAG documents")
	cmd.Flags().Float64Var(&ragThreshold, "rag-threshold", 0.5, "Minimum RAG score for auto-reply")
	return cmd
}
