package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:   "activate [shell]",
	Short: "Generate a shell script to activate",
	Args: cobra.ExactArgs(1),
	ValidArgs: []string{"zsh", "bash"},
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]
		
		// シェルの種類を検証
		isValidShell := false
		for _, validShell := range cmd.ValidArgs {
			if shell == validShell {
				isValidShell = true
				break
			}
		}
		if !isValidShell {
			return fmt.Errorf("サポートされていないシェルです: %s\n利用可能なシェル: %v", shell, cmd.ValidArgs)
		}

		// meiコマンドの存在確認
		meiPath := filepath.Join(os.Getenv("HOME"), ".local", "bin", "mei")
		if _, err := os.Stat(meiPath); os.IsNotExist(err) {
			return fmt.Errorf("mei コマンドが %s に見つかりません\nインストールを完了してから再度実行してください", meiPath)
		}

		script := generateShellScript(shell)
		fmt.Print(script)
		return nil
	},
	// カスタムの使用法メッセージ
	Example: fmt.Sprintf(`  # Zshの場合
  eval "$(%s/.local/bin/mei activate zsh)"

  # Bashの場合
  eval "$(%s/.local/bin/mei activate bash)"

  # .zshrcや.bashrcに追加する場合
  echo 'eval "$(%s/.local/bin/mei activate zsh)"' >> ~/.zshrc  # Zshの場合
  echo 'eval "$(%s/.local/bin/mei activate bash)"' >> ~/.bashrc  # Bashの場合`, os.Getenv("HOME"), os.Getenv("HOME"), os.Getenv("HOME"), os.Getenv("HOME")),
}

func generateShellScript(shell string) string {
	homeDir := os.Getenv("HOME")
	return fmt.Sprintf(`mei() {
  "%s/.local/bin/mei" "$@"
}
`, homeDir)
}

func init() {
	rootCmd.AddCommand(activateCmd)
}
