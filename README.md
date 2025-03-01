# mei

## セットアップ

1. git clone https://github.com/reblabers/mei
2. mise trust && mise install
3. mise deploy
4. ~/.local/bin/mei shell setup $SHELL

## 利用可能なコマンド

### 基本コマンド

- `mei` - meiのルートコマンド

### プロジェクト管理

- `mei add` - 現在のディレクトリをmeiに登録します

### 環境変数管理

- `mei env` - 環境変数を管理します
- `mei env add [key]` - 環境変数を追加します

### シェル関連

- `mei shell` - シェル関連のコマンドです
- `mei shell setup [shell]` - シェルの設定を行います

### リポジトリ関連

- `mei repo` (または `mei r`) - リポジトリ関連のコマンドです
- `mei repo setup` - リポジトリの設定を行います

### その他

- `mei activate` - シェル初期化スクリプトから呼び出し、mei の環境をアクティブにします。
