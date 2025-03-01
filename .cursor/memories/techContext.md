# 技術コンテキスト

## 使用技術
- Go言語
- Cobraライブラリ（CLIフレームワーク）
- mise（開発環境管理）

## 開発環境セットアップ
1. リポジトリのクローン: `git clone https://github.com/reblabers/mei`
2. mise信頼と依存関係インストール: `mise trust && mise install`
3. デプロイ: `mise deploy`
4. シェル設定: `~/.local/bin/mei shell setup $SHELL`

## 技術的制約
- Go 1.x互換性の維持
- CLIインターフェースの一貫性
- シェル間の互換性（bash, zsh, fish） 