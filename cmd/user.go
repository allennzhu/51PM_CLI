package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

// UserInfo 用户信息
type UserInfo struct {
	Id       int    `json:"id"`
	NickName string `json:"nick_name"`
}

// UserInfoResponse 用户查询响应
type UserInfoResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Data UserInfo `json:"data"`
	} `json:"data"`
}

var (
	userNickname   string
	userOutputJSON bool
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "用户管理",
	Long:  `51PM用户相关命令。`,
}

var userLookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "通过用户名称查询用户ID",
	Long: `根据用户名称（昵称）查询用户信息，返回用户ID等信息。
此命令用于将用户名称转换为用户ID，供其他命令使用。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if userNickname == "" && len(args) > 0 {
			userNickname = args[0]
		}
		if userNickname == "" {
			return fmt.Errorf("请通过 --name 参数或直接传入用户名称")
		}

		localToken, err := loadToken()
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("nick_name", userNickname)

		reqURL := fmt.Sprintf("%s/manage_api/user/get_user_info_by_nick_name?%s", baseURL, params.Encode())
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			return fmt.Errorf("构造请求失败: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+localToken.AccessToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("请求失败: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取响应失败: %w", err)
		}

		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("Token已过期，请重新执行 login 命令")
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}

		var result UserInfoResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}

		if result.Code != 0 {
			return fmt.Errorf("请求失败[%d]: %s", result.Code, result.Message)
		}

		user := result.Data.Data
		if user.Id == 0 {
			fmt.Printf("未找到名称包含 \"%s\" 的用户\n", userNickname)
			return nil
		}

		if userOutputJSON {
			output, _ := json.MarshalIndent(user, "", "  ")
			fmt.Println(string(output))
			return nil
		}

		// 表格输出
		fmt.Println()
		fmt.Printf("  \033[1m找到 %d 个匹配用户：\033[0m\n\n", 1)
		fmt.Printf("  \033[90m%-8s %-16s\033[0m\n", "ID", "昵称")
		fmt.Printf("  \033[90m%-8s %-16s\033[0m\n", "────", "────────")
		fmt.Printf("  %-8d %-16s\n", user.Id, user.NickName)
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userLookupCmd)

	userLookupCmd.Flags().StringVar(&userNickname, "name", "", "用户名称（昵称）")
	userLookupCmd.Flags().BoolVar(&userOutputJSON, "json", false, "以JSON格式输出")
}
