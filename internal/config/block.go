package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// BlockManager は設定ブロックの管理を行う構造体です
type BlockManager struct {
	Label         string
	Content       string
	CommentPrefix string
}

// NewBlockManager は新しいBlockManagerを作成します
func NewBlockManager(label string, content string, commentPrefix string) *BlockManager {
	return &BlockManager{
		Label:         label,
		Content:       content,
		CommentPrefix: commentPrefix,
	}
}

// Format はブロックを整形します
func (b *BlockManager) Format() string {
	content := b.Content
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return fmt.Sprintf("%s BEGIN:%s\n%s%s END:%s\n",
		b.CommentPrefix, b.Label,
		content,
		b.CommentPrefix, b.Label)
}

// UpdateFile は指定されたファイルの内容を更新します
func (b *BlockManager) UpdateFile(filepath string) error {
	// ファイルの読み込み
	content, err := os.ReadFile(filepath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
	}

	formattedBlock := b.Format()

	// ファイルが存在しない場合は新規作成
	if os.IsNotExist(err) {
		if err := os.WriteFile(filepath, []byte(formattedBlock), 0644); err != nil {
			return fmt.Errorf("ファイルの作成に失敗しました: %w", err)
		}
		return nil
	}

	// 既存のブロックを検索（コメント記号をエスケープ）
	escapedPrefix := regexp.QuoteMeta(b.CommentPrefix)
	pattern := fmt.Sprintf(`(?s)%s BEGIN:%s.*?%s END:%s\n?`,
		escapedPrefix, b.Label,
		escapedPrefix, b.Label)
	re := regexp.MustCompile(pattern)
	existingContent := string(content)

	var newContent string
	if re.MatchString(existingContent) {
		// 既存のブロックを置換
		newContent = re.ReplaceAllString(existingContent, formattedBlock)
	} else {
		// ファイル末尾に追加（必要に応じて改行を追加）
		if !strings.HasSuffix(existingContent, "\n") {
			existingContent += "\n"
		}
		if !strings.HasSuffix(existingContent, "\n\n") {
			existingContent += "\n"
		}
		newContent = existingContent + formattedBlock
	}

	// ファイルに書き戻し
	if err := os.WriteFile(filepath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗しました: %w", err)
	}

	return nil
}
