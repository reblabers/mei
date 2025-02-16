package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mei/internal/config"
	"github.com/spf13/cobra"
)

//go:embed templates/custom.sh
var customConfigTemplate string

// getCommentPrefix はファイル拡張子に応じたコメント記号を返します
func getCommentPrefix(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".vim", ".vimrc":
		return "\""
	case ".el":
		return ";"
	default:
		return "#"
	}
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "各種セットアップを行います",
}

var shellSetupCmd = &cobra.Command{
	Use:   "shell",
	Short: "シェルの設定を行います",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir := os.Getenv("HOME")
		zprofilePath := filepath.Join(homeDir, ".zprofile")

		blockManager := config.NewBlockManager("custom", customConfigTemplate)
		blockManager.WithCommentPrefix(getCommentPrefix(zprofilePath))

		if err := blockManager.UpdateFile(zprofilePath); err != nil {
			return fmt.Errorf("設定の更新に失敗しました: %w", err)
		}

		fmt.Println("シェルの設定を更新しました")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.AddCommand(shellSetupCmd)
}
