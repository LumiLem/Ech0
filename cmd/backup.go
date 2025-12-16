package cmd

import (
	"github.com/spf13/cobra"

	"github.com/lin-snow/ech0/internal/cli"
)

// backupCmd 是备份数据的命令
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "备份数据",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoBackup()
	},
}

// restoreCmd 是恢复数据的命令
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "恢复数据",
	Run: func(cmd *cobra.Command, args []string) {
		// 获取待恢复的备份文件路径
		if len(args) < 1 {
			_ = cmd.Help()
			return
		}

		cli.DoRestore(args[0])
	},
}

// init 函数用于初始化根命令和子命令
func init() {
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
}
