/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var baseURL string
var frontendURL string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "51pm",
	Short: "51PM项目管理系统CLI工具",
	Long:  `51PM CLI - 用于与51PM项目管理系统交互的命令行工具，支持登录鉴权和任务管理。`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "http://51pm.51aes.com:218", "51PM 后端 API 服务地址")
	rootCmd.PersistentFlags().StringVar(&frontendURL, "frontend-url", "http://51pm.51aes.com:771", "51PM 前端页面地址（用于浏览器登录）")
}
