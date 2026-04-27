package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
)

// LocalToken 本地保存的token信息
type LocalToken struct {
	AccessToken string `json:"access_token"`
}

var loginToken string
var loginBrowser bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录51PM系统",
	Long: `保存Token用于后续API调用。

使用方式：
  51pm login --browser                # 自动打开浏览器登录并获取Token（推荐）
  51pm login --token <your_token>     # 直接传入token
  51pm login                          # 交互式输入token`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := loginToken

		// 浏览器自动登录模式
		if loginBrowser {
			return loginViaBrowser()
		}

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

// loginViaBrowser 通过浏览器自动登录，从 localStorage 抓取 oauthToken
func loginViaBrowser() error {
	// 构建登录页 URL
	loginURL := strings.TrimRight(frontendURL, "/")

	fmt.Println("正在打开浏览器，请在页面中完成登录...")
	fmt.Println("登录页地址:", loginURL)

	// 创建 chromedp 上下文，使用有头浏览器（让用户可见并操作）
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		// 禁用自动化提示栏
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置总超时 5 分钟，给用户足够时间登录
	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	var token string

	fmt.Println("等待登录完成（5分钟内有效）...")

	err := chromedp.Run(ctx,
		// 导航到登录页
		chromedp.Navigate(loginURL),

		// 轮询 localStorage 等待 oauthToken 出现
		chromedp.ActionFunc(func(ctx context.Context) error {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return fmt.Errorf("登录超时，请重试")
				case <-ticker.C:
					var result string
					err := chromedp.Evaluate(`localStorage.getItem('oauthToken') || ''`, &result).Do(ctx)
					if err != nil {
						continue
					}
					result = strings.TrimSpace(result)
					if result != "" {
						token = result
						return nil
					}
				}
			}
		}),
	)
	if err != nil {
		return fmt.Errorf("浏览器登录失败: %w", err)
	}

	if token == "" {
		return fmt.Errorf("未获取到Token")
	}

	// 清理 token（去除可能的引号包裹）
	token = strings.Trim(token, "\"")

	localToken := LocalToken{
		AccessToken: token,
	}
	if err := saveToken(localToken); err != nil {
		return fmt.Errorf("保存Token失败: %w", err)
	}

	fmt.Println("登录成功！Token 已自动保存。")
	return nil
}

// verifyToken 通过调用 API 验证 token 是否有效
func verifyToken(token string) bool {
	reqURL := fmt.Sprintf("%s/manage_api/user/get_user_info_by_nick_name?nick_name=test", strings.TrimRight(baseURL, "/"))
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return resp.StatusCode != http.StatusUnauthorized
}

// ensureToken 确保 token 存在且有效，否则自动触发浏览器登录
func ensureToken() (*LocalToken, error) {
	localToken, err := loadToken()
	if err == nil && verifyToken(localToken.AccessToken) {
		return localToken, nil
	}

	// token 不存在或已过期，自动触发浏览器登录
	if err != nil {
		fmt.Println("未检测到登录信息，正在自动打开浏览器登录...")
	} else {
		fmt.Println("Token 已过期，正在自动打开浏览器重新登录...")
	}

	if loginErr := loginViaBrowser(); loginErr != nil {
		return nil, fmt.Errorf("自动登录失败: %w", loginErr)
	}

	return loadToken()
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
	loginCmd.Flags().BoolVar(&loginBrowser, "browser", false, "通过浏览器自动登录（推荐）")
}
