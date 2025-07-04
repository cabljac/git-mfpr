package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/user/git-mfpr/internal/migrate"
)

type UI interface {
	StartPR(prRef string)
	HandleEvent(event migrate.Event)
	Error(err error)
	Success(message string)
	Info(message string)
	Command(cmd string)
}

type ConsoleUI struct {
	dryRun bool
}

func New() UI {
	return &ConsoleUI{}
}

func NewWithOptions(dryRun bool) UI {
	return &ConsoleUI{dryRun: dryRun}
}

func (ui *ConsoleUI) StartPR(prRef string) {
	fmt.Printf("\n🔄 Migrating PR %s...\n", prRef)
}

func (ui *ConsoleUI) HandleEvent(event migrate.Event) {
	switch event.Type {
	case "info":
		ui.Info(event.Message)
	case "success":
		ui.Success(event.Message)
	case "error":
		ui.Error(fmt.Errorf("%s", event.Message))
	case "command":
		ui.Command(event.Detail)
	case "step":
		fmt.Printf("→ %s\n", event.Message)
	}
}

func (ui *ConsoleUI) Error(err error) {
	fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
}

func (ui *ConsoleUI) Success(message string) {
	fmt.Printf("✅ %s\n", message)
}

func (ui *ConsoleUI) Info(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

func (ui *ConsoleUI) Command(cmd string) {
	if ui.dryRun {
		fmt.Printf("$ %s (dry-run)\n", cmd)
	} else {
		fmt.Printf("$ %s\n", cmd)
	}
}

func FormatPRInfo(pr *migrate.PRInfo) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("📋 Title: %s", pr.Title))
	lines = append(lines, fmt.Sprintf("👤 Author: %s", pr.Author))
	lines = append(lines, fmt.Sprintf("🔢 Number: #%d", pr.Number))
	lines = append(lines, fmt.Sprintf("🌿 Base Branch: %s", pr.BaseBranch))
	if pr.IsFork {
		lines = append(lines, "🍴 From Fork: Yes")
	}
	return strings.Join(lines, "\n")
}

func FormatCreatePRCommand(pr *migrate.PRInfo, branchName string) string {
	return fmt.Sprintf(`gh pr create --title "%s" \
  --body "Migrated from #%d\nOriginal author: @%s" \
  --base %s`, pr.Title, pr.Number, pr.Author, pr.BaseBranch)
}
