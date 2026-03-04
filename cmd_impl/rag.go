package cmd_impl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ygncode/meta-cli/internal/rag"
)

func init() {
	ragCmd := &cobra.Command{
		Use:   "rag",
		Short: "RAG document search",
	}

	ragCmd.AddCommand(ragIndexCmd())
	ragCmd.AddCommand(ragSearchCmd())
	rootCmd.AddCommand(ragCmd)
}

func ragIndexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "index [directory]",
		Short: "Show index stats for a document directory",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx := GetCtx(cmd)

			dir := rctx.Config.RAGDir
			if len(args) > 0 {
				dir = args[0]
			}
			if dir == "" {
				return fmt.Errorf("provide a directory or set rag_dir in config")
			}

			docs, err := rag.LoadDir(dir)
			if err != nil {
				return err
			}

			idx := rag.Build(docs)
			_ = idx

			type docInfo struct {
				ID    string `json:"id"`
				Path  string `json:"path"`
				Title string `json:"title"`
			}

			var infos []docInfo
			for _, d := range docs {
				infos = append(infos, docInfo{ID: d.ID, Path: d.Path, Title: d.Title})
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Indexed %d chunks from %s\n", len(docs), dir)
			return rctx.Printer.Print(infos)
		},
	}
}

func ragSearchCmd() *cobra.Command {
	var topK int
	var dir string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search documents",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rctx := GetCtx(cmd)

			ragDir := dir
			if ragDir == "" {
				ragDir = rctx.Config.RAGDir
			}
			if ragDir == "" {
				return fmt.Errorf("provide --dir or set rag_dir in config")
			}

			docs, err := rag.LoadDir(ragDir)
			if err != nil {
				return err
			}

			idx := rag.Build(docs)
			results := idx.Search(args[0], topK)

			return rctx.Printer.Print(results)
		},
	}

	cmd.Flags().IntVar(&topK, "top", 5, "Number of results")
	cmd.Flags().StringVar(&dir, "dir", "", "Directory with documents")
	return cmd
}
