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
		
		// シェル名の正規化（パスから取得）
		shell = filepath.Base(shell)
		
		// サポートされているシェルの確認
		supportedShells := map[string]bool{
			"bash": true,
			"zsh":  true,
		}
		
		if !supportedShells[shell] {
			return fmt.Errorf("サポートされていないシェルです: %s (サポート: bash, zsh)", shell)
		}

		homeDir := os.Getenv("HOME")
		rcFile := ""
		reloadCmd := ""
		
		// シェル固有の設定
		switch shell {
		case "zsh":
			rcFile = filepath.Join(homeDir, ".zprofile")
			reloadCmd = ". ~/.zprofile"
		case "bash":
			rcFile = filepath.Join(homeDir, ".bash_profile")
			reloadCmd = ". ~/.bash_profile"
		}

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

		if err := blockManager.UpdateFile(rcFile); err != nil {
			return fmt.Errorf("設定の更新に失敗しました: %w", err)
		}

		fmt.Printf("シェルの設定を更新しました\n")
		fmt.Printf("設定を反映するには以下のコマンドを実行してください:\n")
		fmt.Printf("  %s\n", reloadCmd)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.AddCommand(shellSetupCmd)
}
