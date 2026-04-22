package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// SkillsFS 由 main.go 注入的嵌入式文件系统
var SkillsFS embed.FS

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "安装Copilot/AI Agent技能文件",
	Long:  `将51PM Skill文件安装到本机，使Copilot和其他AI Agent能够调用51pm CLI。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("获取用户目录失败: %w", err)
		}

		skillsBase := filepath.Join(home, ".agents", "skills")
		count := 0

		err = fs.WalkDir(SkillsFS, "skills", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// 只处理 SKILL.md 文件
			if d.IsDir() || d.Name() != "SKILL.md" {
				return nil
			}

			content, err := SkillsFS.ReadFile(path)
			if err != nil {
				return fmt.Errorf("读取嵌入文件 %s 失败: %w", path, err)
			}

			// path 格式: skills/51pm-task/SKILL.md -> 取 51pm-task
			dir := filepath.Dir(path)
			skillName := filepath.Base(dir)

			destDir := filepath.Join(skillsBase, skillName)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}

			destPath := filepath.Join(destDir, "SKILL.md")
			if err := os.WriteFile(destPath, content, 0644); err != nil {
				return fmt.Errorf("写入Skill文件失败: %w", err)
			}

			fmt.Printf("  ✅ %s -> %s\n", skillName, destPath)
			count++
			return nil
		})
		if err != nil {
			return err
		}

		fmt.Printf("\n共安装 %d 个 Skill\n", count)
		fmt.Println("\n现在可以在 VS Code Copilot 中直接说：")
		fmt.Println("   \"查看我的任务\"")
		fmt.Println("   \"项目任务进度怎么样\"")
		fmt.Println("   \"筛选状态为doing的任务\"")
		fmt.Println("\n⚠️  使用前请确保已执行: 51pm login --token <TOKEN>")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
