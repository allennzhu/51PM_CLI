/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"embed"

	"github.com/spf13/51PM_CLI/cmd"
)

//go:embed skills
var skillsFS embed.FS

func main() {
	cmd.SkillsFS = skillsFS
	cmd.Execute()
}
