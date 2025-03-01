package cmd

import (
	"github.com/spf13/cobra"
)

// projectCmd はプロジェクト関連のコマンドを表します
var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"p"},
	Short:   "プロジェクト関連のコマンド",
}

func init() {
	rootCmd.AddCommand(projectCmd)
} 