package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Project はプロジェクト情報を表す構造体
type Project struct {
	Name      string    `yaml:"name"`       // プロジェクト名（デフォルトはディレクトリ名）
	Path      string    `yaml:"path"`       // プロジェクトのパス
	CreatedAt time.Time `yaml:"created_at"` // 登録日時
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "現在のディレクトリをmeiに登録します",
	Run: func(cmd *cobra.Command, args []string) {
		// 現在のディレクトリを取得
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Println("現在のディレクトリを取得できませんでした:", err)
			return
		}

		// ~/.mei/projects.ymlのパスを作成
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("ホームディレクトリを取得できませんでした:", err)
			return
		}
		meiDir := filepath.Join(homeDir, ".mei")
		projectsFile := filepath.Join(meiDir, "projects.yml")

		// .meiディレクトリが存在しない場合は作成
		if _, err := os.Stat(meiDir); os.IsNotExist(err) {
			err = os.MkdirAll(meiDir, 0755)
			if err != nil {
				fmt.Println(".meiディレクトリを作成できませんでした:", err)
				return
			}
		}

		// 既存のプロジェクトリストを読み込む
		var projects []Project
		if _, err := os.Stat(projectsFile); !os.IsNotExist(err) {
			data, err := os.ReadFile(projectsFile)
			if err != nil {
				fmt.Println("プロジェクトファイルを読み込めませんでした:", err)
				return
			}

			// 古い形式（文字列の配列）からの移行をサポート
			var oldProjects []string
			err = yaml.Unmarshal(data, &oldProjects)
			if err == nil && len(oldProjects) > 0 {
				// 古い形式から新しい形式に変換
				for _, path := range oldProjects {
					projects = append(projects, Project{
						Name:      filepath.Base(path),
						Path:      path,
						CreatedAt: time.Now(),
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
		}

		// 既に登録されているか確認
		for _, project := range projects {
			if project.Path == currentDir {
				fmt.Println("このディレクトリは既に登録されています")
				return
			}
		}

		// 新しいプロジェクトを追加
		newProject := Project{
			Name:      filepath.Base(currentDir),
			Path:      currentDir,
			CreatedAt: time.Now(),
		}
		projects = append(projects, newProject)

		// YAMLとして保存
		data, err := yaml.Marshal(projects)
		if err != nil {
			fmt.Println("YAMLの生成に失敗しました:", err)
			return
		}

		err = os.WriteFile(projectsFile, data, 0644)
		if err != nil {
			fmt.Println("プロジェクトファイルの保存に失敗しました:", err)
			return
		}

		fmt.Printf("プロジェクトを登録しました: %s\n", currentDir)
	},
}

func init() {
	// rootCmd.AddCommand(addCmd) // 古い登録方法
	projectCmd.AddCommand(addCmd) // projectコマンドのサブコマンドとして登録
} 