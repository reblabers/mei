package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"bytes"

	"mei/internal/config"
	"github.com/spf13/cobra"
)

//go:embed templates/custom.sh.template
var customConfigTemplate string

type shellTemplateData struct {
	HomeDir string
	Shell   string
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "シェル関連のコマンドです",
}

var shellSetupCmd = &cobra.Command{
	Use:   "setup [shell]",
	Short: "シェルの設定を行います",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]
		homeDir := os.Getenv("HOME")
		zprofilePath := filepath.Join(homeDir, ".zprofile")

		// テンプレートを処理
		tmpl, err := template.New("custom").Parse(customConfigTemplate)
		if err != nil {
			return fmt.Errorf("テンプレートの解析に失敗しました: %w", err)
		}

		data := shellTemplateData{
			HomeDir: homeDir,
			Shell:   shell,
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
		}

		blockManager := config.NewBlockManager("custom", buf.String(), "#")

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
