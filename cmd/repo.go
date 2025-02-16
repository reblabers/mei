package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"mei/internal/config"
	"github.com/spf13/cobra"
)

//go:embed templates/exclude.txt
var excludeConfigTemplate string

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "リポジトリ関連のコマンドです",
}

var repoSetupCmd = &cobra.Command{
	Use:   "setup",
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

		blockManager := config.NewBlockManager("mei", excludeConfigTemplate, "#")

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

var repoFavCmd = &cobra.Command{
	Use:   "fav",
	Short: "お気に入りリポジトリを表示します",
	RunE: func(cmd *cobra.Command, args []string) error {
		favorites, err := config.LoadFavorites()
		if err != nil {
			return fmt.Errorf("お気に入りの読み込みに失敗しました: %w", err)
		}

		validRepos := favorites.GetValidRepositories()
		for _, repo := range validRepos {
			fmt.Println(repo)
		}
		return nil
	},
}

var repoFavAddCmd = &cobra.Command{
	Use:   "add",
	Short: "現在のgitディレクトリをお気に入りに追加します",
	RunE: func(cmd *cobra.Command, args []string) error {
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("カレントディレクトリの取得に失敗しました: %w", err)
		}

		// gitがないディレクトリならエラー
		gitDir := filepath.Join(currentDir, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			return fmt.Errorf("Gitリポジトリが見つかりません")
		}

		favorites, err := config.LoadFavorites()
		if err != nil {
			return fmt.Errorf("お気に入りの読み込みに失敗しました: %w", err)
		}

		if err := favorites.Add(currentDir); err != nil {
			return err
		}

		if err := favorites.Save(); err != nil {
			return fmt.Errorf("お気に入りの保存に失敗しました: %w", err)
		}

		fmt.Printf("お気に入りに追加しました: %s\n", currentDir)
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
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(repoSetupCmd)
	repoCmd.AddCommand(repoFavCmd)
	repoFavCmd.AddCommand(repoFavAddCmd)
	repoSetupCmd.Flags().String("user", "", "Gitユーザー名を指定します")
}
