package cmd

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"mei/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/fs"
)

//go:embed templates/exclude.txt
var excludeConfigTemplate string

//go:embed templates/.cursor
var cursorFS embed.FS

var projectSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "登録されているプロジェクトに必要なファイルをコピーします",
	Run: func(cmd *cobra.Command, args []string) {
		// ~/.mei/projects.ymlのパスを作成
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("ホームディレクトリを取得できませんでした:", err)
			return
		}
		meiDir := filepath.Join(homeDir, ".mei")
		projectsFile := filepath.Join(meiDir, "projects.yml")

		// プロジェクトファイルが存在しない場合
		if _, err := os.Stat(projectsFile); os.IsNotExist(err) {
			fmt.Println("登録されているプロジェクトはありません")
			return
		}

		// プロジェクトリストを読み込む
		data, err := os.ReadFile(projectsFile)
		if err != nil {
			fmt.Println("プロジェクトファイルを読み込めませんでした:", err)
			return
		}

		var projects []Project
		// 古い形式（文字列の配列）からの移行をサポート
		var oldProjects []string
		err = yaml.Unmarshal(data, &oldProjects)
		if err == nil && len(oldProjects) > 0 {
			// 古い形式から新しい形式に変換
			for _, path := range oldProjects {
				projects = append(projects, Project{
					Name:      filepath.Base(path),
					Path:      path,
					GitUser:   "", // 空のGitUser
					CreatedAt: time.Now(), // nilではなく現在時刻を設定
				})
			}
		} else {
			// 新しい形式として読み込み
			err = yaml.Unmarshal(data, &projects)
			if err != nil {
				fmt.Println("YAMLの解析に失敗しました:", err)
				return
			}
		}

		// プロジェクトが登録されていない場合
		if len(projects) == 0 {
			fmt.Println("登録されているプロジェクトはありません")
			return
		}

		// ~/.mei/cursor ディレクトリのパス
		cursorSourceDir := filepath.Join(meiDir, "cursor")
		
		// ~/.mei/cursor ディレクトリが存在するか確認
		if _, err := os.Stat(cursorSourceDir); os.IsNotExist(err) {
			fmt.Println("~/.mei/cursor ディレクトリが存在しません")
			return
		}

		// 各プロジェクトに対して処理を実行
		for _, project := range projects {
			fmt.Printf("プロジェクト %s を同期中...\n", project.Name)
			
			// .cursor ディレクトリをコピー
			targetDir := filepath.Join(project.Path, ".cursor")
			err := copyDir(cursorSourceDir, targetDir)
			if err != nil {
				fmt.Printf("%s へのcursorディレクトリのコピーに失敗しました: %v\n", project.Name, err)
				continue
			}
			
			// repo setup相当の処理を実行
			if err := setupRepo(project); err != nil {
				fmt.Printf("%s のrepo setup処理に失敗しました: %v\n", project.Name, err)
				continue
			}
			
			fmt.Printf("%s の同期が完了しました\n", project.Name)
		}
		
		fmt.Println("すべてのプロジェクトの同期が完了しました")
	},
}

// setupRepo はプロジェクトに対してrepo setup相当の処理を行います
func setupRepo(project Project) error {
	// gitディレクトリの確認
	gitDir := filepath.Join(project.Path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Gitリポジトリがない場合はスキップ
		fmt.Printf("%s はGitリポジトリではありません。スキップします。\n", project.Name)
		return nil
	}

	// .git/info/excludeファイルのパスを構築
	excludePath := filepath.Join(gitDir, "info", "exclude")

	// info ディレクトリが存在しない場合は作成
	infoDir := filepath.Dir(excludePath)
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		return fmt.Errorf("infoディレクトリの作成に失敗しました: %w", err)
	}

	// excludeファイルを更新
	blockManager := config.NewBlockManager("mei", excludeConfigTemplate, "#")
	if err := blockManager.UpdateFile(excludePath); err != nil {
		return fmt.Errorf("excludeファイルの更新に失敗しました: %w", err)
	}

	// GitUser設定が指定されている場合はGit設定を更新
	if project.GitUser != "" {
		// リポジトリのルートディレクトリパスを取得（.gitの親ディレクトリ）
		repoRoot := filepath.Dir(gitDir)
		repoName := filepath.Base(repoRoot)

		// Git設定を実行
		gitCommands := []struct {
			args []string
			desc string
			ignoreError bool
		}{
			{[]string{"config", "--local", "user.name", project.GitUser}, "ユーザー名の設定", false},
			{[]string{"config", "--local", "user.email", project.GitUser + "@gmail.com"}, "メールアドレスの設定", false},
			// originの削除（存在しない場合のエラーは無視）
			{[]string{"remote", "remove", "origin"}, "既存のoriginの削除", true},
			{[]string{"remote", "add", "origin", fmt.Sprintf("git@%s.github.com:%s/%s.git", project.GitUser, project.GitUser, repoName)}, "リモートの設定", false},
		}

		for _, cmd := range gitCommands {
			if err := runGitCommand(gitDir, cmd.args...); err != nil && !cmd.ignoreError {
				return fmt.Errorf("%sに失敗しました: %w", cmd.desc, err)
			}
		}
		fmt.Printf("%s のGitユーザー設定を更新しました（%s）\n", project.Name, project.GitUser)
	}

	// EnvKeys設定が指定されている場合は環境変数を更新
	if len(project.EnvKeys) > 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("ホームディレクトリの取得に失敗しました: %w", err)
		}

		// 環境変数ファイルのパスを~/.mei/env/に変更
		meiDir := filepath.Join(homeDir, ".mei")
		envFileDest := filepath.Join(project.Path, ".env")
		
		for _, key := range project.EnvKeys {
			envFileSource := filepath.Join(meiDir, "env", key)
			
			// .envファイルが存在しない場合はスキップ
			if _, err := os.Stat(envFileSource); os.IsNotExist(err) {
				fmt.Printf("警告: envファイルが見つかりません: %s\n", envFileSource)
				continue
			}
			
			// .envファイルの内容を読み込む
			content, err := os.ReadFile(envFileSource)
			if err != nil {
				fmt.Printf("警告: .envファイルの読み込みに失敗しました: %v\n", err)
				continue
			}
			
			// BlockManagerを使って.envファイルに追記・上書き
			blockManager := config.NewBlockManager(key, string(content), "#")
			if err := blockManager.UpdateFile(envFileDest); err != nil {
				fmt.Printf("警告: .envファイルの更新に失敗しました: %v\n", err)
				continue
			}
			
			fmt.Printf("%s の環境変数(%s)を更新しました\n", project.Name, key)
		}
	}

	return nil
}

// ディレクトリをコピーする関数
func copyDir(src, dst string) error {
	// 対象ディレクトリを作成（既に存在する場合は何もしない）
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	// ソースディレクトリ内のファイルとディレクトリを取得
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// サブディレクトリの場合は再帰的にコピー
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// ファイルの場合はコピー
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ファイルをコピーする関数
func copyFile(src, dst string) error {
	// ソースファイルを開く
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 宛先ファイルを作成
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// ファイルの内容をコピー
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// ファイルの権限をコピー
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// copyCursorDirectory は.cursorディレクトリをコピーします
func copyCursorDirectory(destRoot string) error {
	return fs.WalkDir(cursorFS, "templates/.cursor", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 相対パスを計算
		relPath, err := filepath.Rel("templates/.cursor", path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destRoot, ".cursor", relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// ファイルの場合
		data, err := cursorFS.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, 0644)
	})
}

// runGitCommand はGitコマンドを実行します
func runGitCommand(gitDir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = filepath.Dir(gitDir) // .gitの親ディレクトリで実行
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	projectCmd.AddCommand(projectSyncCmd)
} 