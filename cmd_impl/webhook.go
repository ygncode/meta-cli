package cmd_impl

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/config"
	"github.com/ygncode/meta-cli/internal/daemon"
	"github.com/ygncode/meta-cli/internal/messenger"
)

func init() {
	webhookCmd := &cobra.Command{
		Use:   "webhook",
		Short: "Webhook server for real-time messages",
	}

	webhookCmd.AddCommand(webhookServeCmd())
	webhookCmd.AddCommand(webhookSubscribeCmd())
	webhookCmd.AddCommand(webhookStatusCmd())
	webhookCmd.AddCommand(webhookStopCmd())
	rootCmd.AddCommand(webhookCmd)
}

func webhookServeCmd() *cobra.Command {
	var port int
	var verifyToken string
	var daemonFlag bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the webhook HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if daemonFlag {
				return daemonize()
			}

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

			svc := messenger.NewService(rctx.Client)
			if err := svc.SubscribeWebhook(cmd.Context()); err != nil {
				return fmt.Errorf("subscribe webhook fields: %w", err)
			}
			log.Printf("Subscribed to webhook fields: messages, message_echoes")

			handler := &messenger.WebhookHandler{
				VerifyToken: verifyToken,
				AppSecret:   appSecret,
				PageID:      rctx.PageID,
				Store:       store,
				Messenger:   svc,
			}

			if port == 0 {
				port = rctx.Config.WebhookPort
			}

			addr := fmt.Sprintf(":%d", port)

			// Graceful shutdown
			dir, _ := config.Dir()
			pidPath := daemon.PIDPath(dir)

			srv := &http.Server{Addr: addr, Handler: handler}
			errCh := make(chan error, 1)
			go func() { errCh <- srv.ListenAndServe() }()

			log.Printf("Webhook server listening on %s", addr)

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			select {
			case sig := <-sigCh:
				log.Printf("Received %v, shutting down...", sig)
				ctx, cancel := context.WithTimeout(context.Background(), 5*_second)
				defer cancel()
				srv.Shutdown(ctx)
			case err := <-errCh:
				if err != nil && err != http.ErrServerClosed {
					return err
				}
			}

			daemon.RemovePID(pidPath)
			return nil
		},
	}

	cmd.Flags().IntVar(&port, "port", 0, "Port to listen on (default from config)")
	cmd.Flags().StringVar(&verifyToken, "verify-token", "", "Webhook verify token")
	cmd.Flags().BoolVar(&daemonFlag, "daemon", false, "Run in background")
	return cmd
}

const _second = 1_000_000_000 // time.Second as untyped constant to avoid import

func daemonize() error {
	dir, err := config.Dir()
	if err != nil {
		return err
	}

	pidPath := daemon.PIDPath(dir)
	if pid, err := daemon.ReadPID(pidPath); err == nil && daemon.IsRunning(pid) {
		return fmt.Errorf("webhook server already running (PID %d)", pid)
	}

	logPath := daemon.LogPath(dir)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer logFile.Close()

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	// Filter out --daemon from args
	var args []string
	for _, a := range os.Args[1:] {
		if a != "--daemon" && a != "-daemon" {
			args = append(args, a)
		}
	}

	child := exec.Command(exePath, args...)
	child.Stdout = logFile
	child.Stderr = logFile
	child.Stdin = nil
	child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := child.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	if err := daemon.WritePID(pidPath, child.Process.Pid); err != nil {
		return fmt.Errorf("write PID file: %w", err)
	}

	fmt.Printf("Webhook server started (PID %d)\n", child.Process.Pid)
	fmt.Printf("Log: %s\n", logPath)
	return nil
}

func webhookSubscribeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe page to webhook fields (messages, message_echoes)",
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx, err := requirePageClient(cmd)
			if err != nil {
				return err
			}

			svc := messenger.NewService(rctx.Client)
			if err := svc.SubscribeWebhook(cmd.Context()); err != nil {
				return err
			}

			rctx.Printer.OK("Subscribed to webhook fields: messages, message_echoes")
			return nil
		},
	}
}

func webhookStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check if the webhook server is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := config.Dir()
			if err != nil {
				return err
			}

			pidPath := daemon.PIDPath(dir)
			pid, err := daemon.ReadPID(pidPath)
			if err != nil {
				fmt.Println("Webhook server is not running")
				return nil
			}

			if !daemon.IsRunning(pid) {
				daemon.RemovePID(pidPath)
				fmt.Println("Webhook server is not running (stale PID file cleaned)")
				return nil
			}

			fmt.Printf("Webhook server is running (PID %d)\n", pid)
			return nil
		},
	}
}

func webhookStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the webhook server",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := config.Dir()
			if err != nil {
				return err
			}

			pidPath := daemon.PIDPath(dir)
			pid, err := daemon.ReadPID(pidPath)
			if err != nil {
				return fmt.Errorf("webhook server is not running (no PID file)")
			}

			if !daemon.IsRunning(pid) {
				daemon.RemovePID(pidPath)
				return fmt.Errorf("webhook server is not running (stale PID file cleaned)")
			}

			if err := daemon.StopProcess(pid); err != nil {
				return err
			}

			daemon.RemovePID(pidPath)
			fmt.Printf("Webhook server stopped (PID %d)\n", pid)
			return nil
		},
	}
}
