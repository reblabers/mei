package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"bytes"

	"github.com/spf13/cobra"
)

//go:embed templates/activate.sh.template
var activateTemplate string

type activateTemplateData struct {
	HomeDir string
}

// generateShellScript はシェルスクリプトを生成します
func generateShellScript() (string, error) {
	tmpl, err := template.New("activate").Parse(activateTemplate)
	if err != nil {
		return "", fmt.Errorf("テンプレートの解析に失敗しました: %w", err)
	}

	data := activateTemplateData{
		HomeDir: os.Getenv("HOME"),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	return buf.String(), nil
}

// validateMeiCommand はmeiコマンドの存在を確認します
func validateMeiCommand() error {
	meiPath := filepath.Join(os.Getenv("HOME"), ".local", "bin", "mei")
	if _, err := os.Stat(meiPath); os.IsNotExist(err) {
		return fmt.Errorf("mei コマンドが %s に見つかりません\nインストールを完了してから再度実行してください", meiPath)
	}
	return nil
}

// newShellActivateCmd は各シェル用のactivateコマンドを生成します
func newShellActivateCmd(shell string) *cobra.Command {
	return &cobra.Command{
		Use:   shell,
		Short: fmt.Sprintf("%s用の設定を生成します", shell),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateMeiCommand(); err != nil {
				return err
			}

			script, err := generateShellScript()
			if err != nil {
				return err
			}

			fmt.Print(script)
			return nil
		},
		Example: fmt.Sprintf(`  # %sの設定を追加する場合:
  echo 'eval "$(%s/.local/bin/mei activate %s)"' >> ~/.%src`, 
			shell, os.Getenv("HOME"), shell, shell),
	}
}

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "シェルの設定を生成します",
	// カスタムの使用法メッセージ
	Example: fmt.Sprintf(`  # Zshの場合
  eval "$(%s/.local/bin/mei activate zsh)"

  # Bashの場合
  eval "$(%s/.local/bin/mei activate bash)"

  # .zshrcや.bashrcに追加する場合
  echo 'eval "$(%s/.local/bin/mei activate zsh)"' >> ~/.zshrc  # Zshの場合
  echo 'eval "$(%s/.local/bin/mei activate bash)"' >> ~/.bashrc  # Bashの場合`, os.Getenv("HOME"), os.Getenv("HOME"), os.Getenv("HOME"), os.Getenv("HOME")),
}

var bashActivateCmd = newShellActivateCmd("bash")
var zshActivateCmd = newShellActivateCmd("zsh")

func init() {
	rootCmd.AddCommand(activateCmd)
	activateCmd.AddCommand(bashActivateCmd, zshActivateCmd)
}
