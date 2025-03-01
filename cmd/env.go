package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"mei/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

		// 環境変数ファイルのパスを~/.mei/env/に変更
		meiDir := filepath.Join(homeDir, ".mei")
		envFileSource := filepath.Join(meiDir, "env", key)

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
		
		// --saveオプションが指定されている場合、プロジェクト設定にも追加
		save, err := cmd.Flags().GetBool("save")
		if err != nil {
			return fmt.Errorf("saveオプションの取得に失敗しました: %w", err)
		}
		
		if save {
			// ~/.mei/projects.ymlのパスを作成
			projectsFile := filepath.Join(meiDir, "projects.yml")
			
			// プロジェクトファイルが存在しない場合
			if _, err := os.Stat(projectsFile); os.IsNotExist(err) {
				return fmt.Errorf("登録されているプロジェクトはありません")
			}
			
			// プロジェクトリストを読み込む
			data, err := os.ReadFile(projectsFile)
			if err != nil {
				return fmt.Errorf("プロジェクトファイルを読み込めませんでした: %w", err)
			}
			
			var projects []Project
			err = yaml.Unmarshal(data, &projects)
			if err != nil {
				return fmt.Errorf("YAMLの解析に失敗しました: %w", err)
			}
			
			// 現在のディレクトリに一致するプロジェクトを探す
			found := false
			for i, project := range projects {
				if project.Path == currentDir {
					// 既に同じキーが存在するか確認
					keyExists := false
					for _, existingKey := range project.EnvKeys {
						if existingKey == key {
							keyExists = true
							break
						}
					}
					
					// キーが存在しない場合のみ追加
					if !keyExists {
						projects[i].EnvKeys = append(projects[i].EnvKeys, key)
					}
					found = true
					break
				}
			}
			
			if !found {
				return fmt.Errorf("現在のディレクトリはプロジェクトとして登録されていません")
			}
			
			// YAMLとして保存
			data, err = yaml.Marshal(projects)
			if err != nil {
				return fmt.Errorf("YAMLの生成に失敗しました: %w", err)
			}
			
			err = os.WriteFile(projectsFile, data, 0644)
			if err != nil {
				return fmt.Errorf("プロジェクトファイルの保存に失敗しました: %w", err)
			}
			
			fmt.Printf("環境変数キー '%s' をプロジェクト設定に追加しました\n", key)
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envAddCmd)
	envAddCmd.Flags().Bool("save", false, "環境変数キーをプロジェクト設定に保存します")
}
