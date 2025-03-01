package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

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

		// 各プロジェクトに .cursor ディレクトリをコピー
		for _, project := range projects {
			targetDir := filepath.Join(project.Path, ".cursor")
			
			// コピー処理を実行
			err := copyDir(cursorSourceDir, targetDir)
			if err != nil {
				fmt.Printf("%s へのコピーに失敗しました: %v\n", project.Name, err)
				continue
			}
			
			fmt.Printf("%s に .cursor をコピーしました\n", project.Name)
		}
		
		fmt.Println("すべてのプロジェクトの同期が完了しました")
	},
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

func init() {
	projectCmd.AddCommand(projectSyncCmd)
} 