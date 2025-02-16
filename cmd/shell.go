package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"mei/internal/config"
	"github.com/spf13/cobra"
)

//go:embed templates/custom.sh
var customConfigTemplate string

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "シェル関連のコマンドです",
}

var shellSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "シェルの設定を行います",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir := os.Getenv("HOME")
		zprofilePath := filepath.Join(homeDir, ".zprofile")

		blockManager := config.NewBlockManager("custom", customConfigTemplate, "#")

		if err := blockManager.UpdateFile(zprofilePath); err != nil {
			return fmt.Errorf("設定の更新に失敗しました: %w", err)
		}

		fmt.Println("シェルの設定を更新しました")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.AddCommand(shellSetupCmd)
}
