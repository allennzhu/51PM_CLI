package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// LocalToken 本地保存的token信息
type LocalToken struct {
	AccessToken string `json:"access_token"`
}

var loginToken string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录51PM系统",
	Long: `保存Token用于后续API调用。

使用方式：
  51pm login --token <your_token>    # 直接传入token
  51pm login                         # 交互式输入token

Token可从51PM前端页面登录后，在浏览器开发者工具(F12) -> Application -> Local Storage 中找到 oauthToken 字段。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := loginToken

		if token == "" {
			fmt.Println("请输入Token（可从51PM前端 localStorage 的 oauthToken 获取）:")
			fmt.Print("> ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			token = strings.TrimSpace(input)
		}

		if token == "" {
			return fmt.Errorf("未输入Token，登录取消")
		}

		localToken := LocalToken{
			AccessToken: token,
		}
		if err := saveToken(localToken); err != nil {
			return fmt.Errorf("保存Token失败: %w", err)
		}

		fmt.Println("Token 已保存，可以使用其他命令了。")
		return nil
	},
}

func getTokenDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".51pm_cli")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func getTokenPath() (string, error) {
	dir, err := getTokenDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "token.json"), nil
}

func saveToken(token LocalToken) error {
	path, err := getTokenPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func loadToken() (*LocalToken, error) {
	path, err := getTokenPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("未登录，请先执行 login 命令")
		}
		return nil, err
	}
	var token LocalToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	if token.AccessToken == "" {
		return nil, fmt.Errorf("Token为空，请重新登录")
	}
	return &token, nil
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVar(&loginToken, "token", "", "直接传入Token")
}
