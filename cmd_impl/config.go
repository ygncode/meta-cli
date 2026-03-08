package cmd_impl

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/config"
)

type configEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func init() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}

	configCmd.AddCommand(configSetCmd())
	configCmd.AddCommand(configGetCmd())
	configCmd.AddCommand(configListCmd())
	rootCmd.AddCommand(configCmd)
}

func configSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := setConfigField(cfg, key, value); err != nil {
				return err
			}

			if err := config.Save(cfg); err != nil {
				return err
			}

			rctx := GetCtx(cmd)
			rctx.Printer.OK(fmt.Sprintf("%s set to %s", key, value))
			return nil
		},
	}
}

func configGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			value, err := getConfigField(cfg, key)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		},
	}
}

func configListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all config values",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			rctx := GetCtx(cmd)
			entries := []configEntry{
				{"default_account", cfg.DefaultAccount},
				{"default_page", cfg.DefaultPage},
				{"graph_api_version", cfg.GraphAPIVersion},
				{"webhook_port", strconv.Itoa(cfg.WebhookPort)},
				{"verify_token", cfg.VerifyToken},
				{"redirect_uri", cfg.RedirectURI},
				{"rag_dir", cfg.RAGDir},
				{"db_path", cfg.DBPath},
				{"debounce_seconds", strconv.Itoa(cfg.DebounceSeconds)},
				{"hooks_endpoint", cfg.HooksEndpoint},
				{"hooks_token", cfg.HooksToken},
				{"auto_reply", strconv.FormatBool(cfg.AutoReply)},
				{"prompt_template", cfg.PromptTemplate},
			}
			return rctx.Printer.Print(entries)
		},
	}
}

func setConfigField(cfg *config.Config, key, value string) error {
	switch key {
	case "default_account":
		cfg.DefaultAccount = value
	case "default_page":
		cfg.DefaultPage = value
	case "graph_api_version":
		cfg.GraphAPIVersion = value
	case "webhook_port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("webhook_port must be an integer: %w", err)
		}
		cfg.WebhookPort = port
	case "verify_token":
		cfg.VerifyToken = value
	case "rag_dir":
		cfg.RAGDir = value
	case "db_path":
		cfg.DBPath = value
	case "debounce_seconds":
		secs, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("debounce_seconds must be an integer: %w", err)
		}
		cfg.DebounceSeconds = secs
	case "hooks_endpoint":
		cfg.HooksEndpoint = value
	case "hooks_token":
		cfg.HooksToken = value
	case "auto_reply":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("auto_reply must be true or false: %w", err)
		}
		cfg.AutoReply = b
	case "prompt_template":
		cfg.PromptTemplate = value
	case "redirect_uri":
		cfg.RedirectURI = value
	default:
		return fmt.Errorf("unknown config key: %s\nvalid keys: default_account, default_page, graph_api_version, webhook_port, verify_token, redirect_uri, rag_dir, db_path, debounce_seconds, hooks_endpoint, hooks_token, auto_reply, prompt_template", key)
	}
	return nil
}

func getConfigField(cfg *config.Config, key string) (string, error) {
	switch key {
	case "default_account":
		return cfg.DefaultAccount, nil
	case "default_page":
		return cfg.DefaultPage, nil
	case "graph_api_version":
		return cfg.GraphAPIVersion, nil
	case "webhook_port":
		return strconv.Itoa(cfg.WebhookPort), nil
	case "verify_token":
		return cfg.VerifyToken, nil
	case "rag_dir":
		return cfg.RAGDir, nil
	case "db_path":
		return cfg.DBPath, nil
	case "debounce_seconds":
		return strconv.Itoa(cfg.DebounceSeconds), nil
	case "hooks_endpoint":
		return cfg.HooksEndpoint, nil
	case "hooks_token":
		return cfg.HooksToken, nil
	case "auto_reply":
		return strconv.FormatBool(cfg.AutoReply), nil
	case "prompt_template":
		return cfg.PromptTemplate, nil
	case "redirect_uri":
		return cfg.RedirectURI, nil
	default:
		return "", fmt.Errorf("unknown config key: %s\nvalid keys: default_account, default_page, graph_api_version, webhook_port, verify_token, redirect_uri, rag_dir, db_path, debounce_seconds, hooks_endpoint, hooks_token, auto_reply, prompt_template", key)
	}
}
