package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Favorites struct {
	Repositories []string `json:"repositories"`
}

func NewFavorites() *Favorites {
	return &Favorites{
		Repositories: make([]string, 0),
	}
}

func (f *Favorites) Add(path string) error {
	// 重複チェック
	for _, repo := range f.Repositories {
		if repo == path {
			return fmt.Errorf("リポジトリは既に登録されています: %s", path)
		}
	}
	f.Repositories = append(f.Repositories, path)
	return nil
}

// GetValidRepositories は有効なGitリポジトリのみを返します
func (f *Favorites) GetValidRepositories() []string {
	var validRepos []string
	for _, repo := range f.Repositories {
		gitDir := filepath.Join(repo, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			validRepos = append(validRepos, repo)
		}
	}
	return validRepos
}

func (f *Favorites) Save() error {
	stateDir := filepath.Join(os.Getenv("HOME"), ".local", "state", "mei")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("ステートディレクトリの作成に失敗しました: %w", err)
	}

	favPath := filepath.Join(stateDir, "favorites.json")
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("JSONのエンコードに失敗しました: %w", err)
	}

	if err := os.WriteFile(favPath, data, 0644); err != nil {
		return fmt.Errorf("お気に入りファイルの保存に失敗しました: %w", err)
	}

	return nil
}

func LoadFavorites() (*Favorites, error) {
	favPath := filepath.Join(os.Getenv("HOME"), ".local", "state", "mei", "favorites.json")
	data, err := os.ReadFile(favPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewFavorites(), nil
		}
		return nil, fmt.Errorf("お気に入りファイルの読み込みに失敗しました: %w", err)
	}

	var fav Favorites
	if err := json.Unmarshal(data, &fav); err != nil {
		return nil, fmt.Errorf("JSONのデコードに失敗しました: %w", err)
	}

	return &fav, nil
} 