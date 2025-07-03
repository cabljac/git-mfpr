package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/user/git-mfpr/internal/migrate"
)

// UI handles all user interface output
type UI interface {
	StartPR(prRef string)
	HandleEvent(event migrate.Event)
	Error(err error)
	Success(message string)
	Info(message string)
	Command(cmd string)
}

// ConsoleUI implements UI for console output
type ConsoleUI struct {
	dryRun bool
}

// New creates a new console UI
func New() UI {
	return &ConsoleUI{}
}

// NewWithOptions creates a new console UI with options
func NewWithOptions(dryRun bool) UI {
	return &ConsoleUI{dryRun: dryRun}
}

// StartPR indicates we're starting to process a PR
func (ui *ConsoleUI) StartPR(prRef string) {
	fmt.Printf("\n🔄 Migrating PR %s...\n", prRef)
}

// HandleEvent handles progress events from the migrator
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

// Error displays an error message
func (ui *ConsoleUI) Error(err error) {
	fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
}

// Success displays a success message
func (ui *ConsoleUI) Success(message string) {
	fmt.Printf("✅ %s\n", message)
}

// Info displays an informational message
func (ui *ConsoleUI) Info(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// Command displays a command that would be or was executed
func (ui *ConsoleUI) Command(cmd string) {
	if ui.dryRun {
		fmt.Printf("$ %s (dry-run)\n", cmd)
	} else {
		fmt.Printf("$ %s\n", cmd)
	}
}

// FormatPRInfo formats PR information for display
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

// FormatCreatePRCommand formats the gh pr create command for display
func FormatCreatePRCommand(pr *migrate.PRInfo, branchName string) string {
	return fmt.Sprintf(`gh pr create --title "%s" \
  --body "Migrated from #%d\nOriginal author: @%s" \
  --base %s`, pr.Title, pr.Number, pr.Author, pr.BaseBranch)
}