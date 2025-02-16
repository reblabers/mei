package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"mei/internal/config"
	"github.com/spf13/cobra"
)

//go:embed templates/custom.sh
var customConfigTemplate string

//go:embed templates/exclude.txt
var excludeConfigTemplate string

// getCommentPrefix はファイル拡張子に応じたコメント記号を返します
func getCommentPrefix(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
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

var repoSetupCmd = &cobra.Command{
	Use:   "repo",
	Short: "リポジトリの設定を行います",
	RunE: func(cmd *cobra.Command, args []string) error {
		// カレントディレクトリ名を取得
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("カレントディレクトリの取得に失敗しました: %w", err)
		}

		// gitがないディレクトリならエラー
		gitDir := filepath.Join(currentDir, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			return fmt.Errorf("Gitリポジトリが見つかりません: %w", err)
		}

		// .git/info/excludeファイルのパスを構築
		excludePath := filepath.Join(gitDir, "info", "exclude")

		// info ディレクトリが存在しない場合は作成
		infoDir := filepath.Dir(excludePath)
		if err := os.MkdirAll(infoDir, 0755); err != nil {
			return fmt.Errorf("infoディレクトリの作成に失敗しました: %w", err)
		}

		blockManager := config.NewBlockManager("mei", excludeConfigTemplate)
		blockManager.WithCommentPrefix(getCommentPrefix(excludePath))

		if err := blockManager.UpdateFile(excludePath); err != nil {
			return fmt.Errorf("excludeファイルの更新に失敗しました: %w", err)
		}

		// ユーザー設定が指定されている場合
		user, err := cmd.Flags().GetString("user")
		if err != nil {
			return fmt.Errorf("ユーザー名の取得に失敗しました: %w", err)
		}
		if user != "" {
			// リポジトリのルートディレクトリパスを取得（.gitの親ディレクトリ）
			repoRoot := filepath.Dir(gitDir)
			repoName := filepath.Base(repoRoot)

			// Git設定を実行
			gitCommands := []struct {
				args []string
				desc string
				// エラーを無視するかどうか
				ignoreError bool
			}{
				{[]string{"config", "--local", "user.name", user}, "ユーザー名の設定", false},
				{[]string{"config", "--local", "user.email", user + "@gmail.com"}, "メールアドレスの設定", false},
				// originの削除（存在しない場合のエラーは無視）
				{[]string{"remote", "remove", "origin"}, "既存のoriginの削除", true},
				{[]string{"remote", "add", "origin", fmt.Sprintf("git@%s.github.com:%s/%s.git", user, user, repoName)}, "リモートの設定", false},
			}

			for _, cmd := range gitCommands {
				if err := runGitCommand(gitDir, cmd.args...); err != nil && !cmd.ignoreError {
					return fmt.Errorf("%sに失敗しました: %w", cmd.desc, err)
				}
			}
			fmt.Printf("Gitユーザー設定を更新しました（%s）\n", user)
		}

		fmt.Println("リポジトリの設定を更新しました")
		return nil
	},
}


// runGitCommand はGitコマンドを実行します
func runGitCommand(gitDir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = filepath.Dir(gitDir) // .gitの親ディレクトリで実行
	cmd.Stderr = os.Stderr
	return cmd.Run()
}


func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.AddCommand(shellSetupCmd, repoSetupCmd)
	repoSetupCmd.Flags().String("user", "", "Gitユーザー名を指定します")
}
