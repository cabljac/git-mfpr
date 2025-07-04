package ui

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/user/git-mfpr/internal/migrate"
)

func TestNew(t *testing.T) {
	ui := New()
	if ui == nil {
		t.Fatal("New() returned nil")
	}
	if _, ok := ui.(*ConsoleUI); !ok {
		t.Fatal("New() did not return a *ConsoleUI")
	}
}

func TestNewWithOptions(t *testing.T) {
	tests := []struct {
		name   string
		dryRun bool
	}{
		{"dry run enabled", true},
		{"dry run disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := NewWithOptions(tt.dryRun)
			if ui == nil {
				t.Fatal("NewWithOptions() returned nil")
			}
			consoleUI, ok := ui.(*ConsoleUI)
			if !ok {
				t.Fatal("NewWithOptions() did not return a *ConsoleUI")
			}
			if consoleUI.dryRun != tt.dryRun {
				t.Errorf("Expected dryRun=%v, got %v", tt.dryRun, consoleUI.dryRun)
			}
		})
	}
}

func TestConsoleUI_StartPR(t *testing.T) {
	ui := &ConsoleUI{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ui.StartPR("owner/repo#123")

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	expected := "\n🔄 Migrating PR owner/repo#123...\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestConsoleUI_Error(t *testing.T) {
	ui := &ConsoleUI{}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	testErr := errors.New("test error")
	ui.Error(testErr)

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe: %v", err)
	}
	os.Stderr = oldStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	expected := "❌ Error: test error\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestConsoleUI_Success(t *testing.T) {
	ui := &ConsoleUI{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ui.Success("Operation completed")

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	expected := "✅ Operation completed\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestConsoleUI_Info(t *testing.T) {
	ui := &ConsoleUI{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ui.Info("Processing...")

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	expected := "ℹ️  Processing...\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestConsoleUI_Command(t *testing.T) {
	tests := []struct {
		name     string
		dryRun   bool
		cmd      string
		expected string
	}{
		{
			name:     "normal mode",
			dryRun:   false,
			cmd:      "git pull origin main",
			expected: "$ git pull origin main\n",
		},
		{
			name:     "dry run mode",
			dryRun:   true,
			cmd:      "git push origin feature",
			expected: "$ git push origin feature (dry-run)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := &ConsoleUI{dryRun: tt.dryRun}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			ui.Command(tt.cmd)

			if err := w.Close(); err != nil {
				t.Fatalf("Failed to close pipe: %v", err)
			}
			os.Stdout = oldStdout

			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r); err != nil {
				t.Fatalf("Failed to read from pipe: %v", err)
			}
			output := buf.String()

			if output != tt.expected {
				t.Errorf("Expected output %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestConsoleUI_HandleEvent(t *testing.T) {
	tests := []struct {
		name         string
		event        migrate.Event
		expectedOut  string
		expectedErr  string
		expectStdout bool
	}{
		{
			name: "info event",
			event: migrate.Event{
				Type:    migrate.EventInfo,
				Message: "Fetching PR...",
			},
			expectedOut:  "ℹ️  Fetching PR...\n",
			expectStdout: true,
		},
		{
			name: "success event",
			event: migrate.Event{
				Type:    migrate.EventSuccess,
				Message: "Migration complete",
			},
			expectedOut:  "✅ Migration complete\n",
			expectStdout: true,
		},
		{
			name: "error event",
			event: migrate.Event{
				Type:    migrate.EventError,
				Message: "Failed to fetch PR",
			},
			expectedErr:  "❌ Error: Failed to fetch PR\n",
			expectStdout: false,
		},
		{
			name: "command event",
			event: migrate.Event{
				Type:   migrate.EventCommand,
				Detail: "git checkout main",
			},
			expectedOut:  "$ git checkout main\n",
			expectStdout: true,
		},
		{
			name: "unknown event",
			event: migrate.Event{
				Type:    "unknown",
				Message: "Unknown event",
			},
			expectedOut:  "",
			expectStdout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := &ConsoleUI{}

			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			rOut, wOut, _ := os.Pipe()
			rErr, wErr, _ := os.Pipe()
			os.Stdout = wOut
			os.Stderr = wErr

			ui.HandleEvent(tt.event)

			if err := wOut.Close(); err != nil {
				t.Fatalf("Failed to close stdout pipe: %v", err)
			}
			if err := wErr.Close(); err != nil {
				t.Fatalf("Failed to close stderr pipe: %v", err)
			}
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			var bufOut, bufErr bytes.Buffer
			if _, err := bufOut.ReadFrom(rOut); err != nil {
				t.Fatalf("Failed to read from stdout pipe: %v", err)
			}
			if _, err := bufErr.ReadFrom(rErr); err != nil {
				t.Fatalf("Failed to read from stderr pipe: %v", err)
			}

			if tt.expectStdout {
				output := bufOut.String()
				if output != tt.expectedOut {
					t.Errorf("Expected stdout %q, got %q", tt.expectedOut, output)
				}
			} else {
				output := bufErr.String()
				if output != tt.expectedErr {
					t.Errorf("Expected stderr %q, got %q", tt.expectedErr, output)
				}
			}
		})
	}
}

func TestFormatPRInfo(t *testing.T) {
	tests := []struct {
		name     string
		pr       *migrate.PRInfo
		expected []string
	}{
		{
			name: "regular PR",
			pr: &migrate.PRInfo{
				Title:      "Fix memory leak",
				Author:     "johndoe",
				Number:     123,
				BaseBranch: "main",
				IsFork:     false,
			},
			expected: []string{
				"📋 Title: Fix memory leak",
				"👤 Author: johndoe",
				"🔢 Number: #123",
				"🌿 Base Branch: main",
			},
		},
		{
			name: "PR from fork",
			pr: &migrate.PRInfo{
				Title:      "Add new feature",
				Author:     "janedoe",
				Number:     456,
				BaseBranch: "develop",
				IsFork:     true,
			},
			expected: []string{
				"📋 Title: Add new feature",
				"👤 Author: janedoe",
				"🔢 Number: #456",
				"🌿 Base Branch: develop",
				"🍴 From Fork: Yes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPRInfo(tt.pr)
			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nGot: %s", expected, result)
				}
			}
		})
	}
}

func TestFormatCreatePRCommand(t *testing.T) {
	pr := &migrate.PRInfo{
		Title:      "Fix critical bug",
		Author:     "developer",
		Number:     789,
		BaseBranch: "main",
	}
	branchName := "fix/critical-bug-789"

	result := FormatCreatePRCommand(pr, branchName)

	expected := []string{
		`gh pr create`,
		`--title "Fix critical bug"`,
		`--body "Migrated from #789\nOriginal author: @developer"`,
		`--base main`,
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected command to contain %q, but it didn't.\nGot: %s", exp, result)
		}
	}
}

func BenchmarkFormatPRInfo(b *testing.B) {
	pr := &migrate.PRInfo{
		Title:      "Benchmark PR",
		Author:     "benchuser",
		Number:     999,
		BaseBranch: "main",
		IsFork:     true,
	}

	for i := 0; i < b.N; i++ {
		_ = FormatPRInfo(pr)
	}
}

func BenchmarkHandleEvent(b *testing.B) {
	ui := &ConsoleUI{}
	event := migrate.Event{
		Type:    migrate.EventInfo,
		Message: "Benchmark event",
	}

	// Redirect output to discard
	oldStdout := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = oldStdout }()

	for i := 0; i < b.N; i++ {
		ui.HandleEvent(event)
	}
}

func ExampleFormatPRInfo() {
	pr := &migrate.PRInfo{
		Title:      "Example PR",
		Author:     "exampleuser",
		Number:     100,
		BaseBranch: "main",
		IsFork:     false,
	}

	fmt.Println(FormatPRInfo(pr))
	// Output:
	// 📋 Title: Example PR
	// 👤 Author: exampleuser
	// 🔢 Number: #100
	// 🌿 Base Branch: main
}
