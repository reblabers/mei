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

- `mei project add` (または `mei p add`) - 現在のディレクトリをmeiに登録します
  - `--git-user` オプション - プロジェクト用のGitユーザー名を指定します
- `mei project ls` (または `mei p ls`) - 登録されているプロジェクト一覧を表示します
- `mei project sync` (または `mei p sync`) - 登録されているプロジェクトに必要なファイルをコピーします
  - `.cursor`ディレクトリを各プロジェクトにコピー
  - Gitリポジトリの場合は`.git/info/exclude`ファイルを更新
  - GitUser設定がある場合はGit設定を更新

### 環境変数管理

- `mei env` - 環境変数を管理します
- `mei env add [key]` - 環境変数を追加します

### シェル関連

- `mei shell` - シェル関連のコマンドです
- `mei shell setup [shell]` - シェルの設定を行います

### リポジトリ関連

- `mei repo` (または `mei r`) - リポジトリ関連のコマンドです
- `mei repo setup` - リポジトリの設定を行います
  - `--user` オプション - Gitユーザー名を指定します

### その他

- `mei activate` - シェル初期化スクリプトから呼び出し、mei の環境をアクティブにします。

## 開発

- `.mise.toml`を使用したタスク自動化
  - `mise run build` - ビルド実行
  - `mise run deploy` - ビルドして配置
  - `mise run app` - アプリケーション実行
