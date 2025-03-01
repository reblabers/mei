package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var projectLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "登録されているプロジェクト一覧を表示します",
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

		// プロジェクトが登録されていない場合
		if len(projects) == 0 {
			fmt.Println("登録されているプロジェクトはありません")
			return
		}

		// 登録日時の新しい順にソート
		sort.Slice(projects, func(i, j int) bool {
			return projects[i].CreatedAt.After(projects[j].CreatedAt)
		})

		// プロジェクト一覧を表示
		fmt.Println("登録されているプロジェクト一覧:")
		for i, project := range projects {
			fmt.Printf("%d: %s (%s)\n", i+1, project.Name, project.Path)
		}
	},
}

func init() {
	projectCmd.AddCommand(projectLsCmd)
} 