package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"mei/internal/config"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "環境変数を管理します",
}

var envAddCmd = &cobra.Command{
	Use:   "add [key]",
	Short: "環境変数を追加します",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("ホームディレクトリの取得に失敗しました: %w", err)
		}

		envFileSource := filepath.Join(homeDir, ".env", key)

		// .envファイルが存在しない場合はエラー
		if _, err := os.Stat(envFileSource); os.IsNotExist(err) {
			return fmt.Errorf("envファイルが見つかりません: %s", envFileSource)
		}

		// .envファイルの内容を読み込む
		content, err := os.ReadFile(envFileSource)
		if err != nil {
			return fmt.Errorf(".envファイルの読み込みに失敗しました: %w", err)
		}

		// カレントディレクトリを取得
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("カレントディレクトリの取得に失敗しました: %w", err)
		}

		envFileDest := filepath.Join(currentDir, ".env")

		// BlockManagerを使って.envファイルに追記・上書き
		blockManager := config.NewBlockManager(key, string(content), "#")
		if err := blockManager.UpdateFile(envFileDest); err != nil {
			return fmt.Errorf(".envファイルの更新に失敗しました: %w", err)
		}

		fmt.Printf(".envファイルを更新しました: %s\n", envFileDest)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envAddCmd)
}
