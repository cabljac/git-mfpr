package migrate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MockMigrator is a mock implementation for testing the CLI
type MockMigrator struct {
	handler EventHandler
}

// NewMockMigrator creates a new mock migrator
func NewMockMigrator(handler EventHandler) Migrator {
	return &MockMigrator{handler: handler}
}

// SetEventHandler sets the event handler
func (m *MockMigrator) SetEventHandler(handler EventHandler) {
	m.handler = handler
}

// MigratePR simulates migrating a single PR
func (m *MockMigrator) MigratePR(prRef string, opts Options) error {
	// Parse the PR reference
	pr, err := m.parsePRRef(prRef)
	if err != nil {
		return err
	}

	// Simulate fetching PR info
	m.emit("info", "Fetching PR information...", "")
	time.Sleep(500 * time.Millisecond)
	
	m.emit("step", fmt.Sprintf("Title: %s", pr.Title), "")
	m.emit("step", fmt.Sprintf("Author: %s", pr.Author), "")
	m.emit("step", fmt.Sprintf("Base branch: %s", pr.BaseBranch), "")
	
	if pr.IsFork {
		m.emit("step", "PR is from a fork", "")
	}

	// Generate branch name
	branchName := opts.BranchName
	if branchName == "" {
		branchName = m.GenerateBranchName(pr)
	}
	m.emit("step", fmt.Sprintf("Branch: %s", branchName), "")

	// Simulate git operations
	time.Sleep(300 * time.Millisecond)
	
	m.emit("command", "", fmt.Sprintf("git checkout %s", pr.BaseBranch))
	m.emit("command", "", "git pull origin main")
	
	time.Sleep(500 * time.Millisecond)
	m.emit("info", "Creating local branch...", "")
	m.emit("command", "", fmt.Sprintf("gh pr checkout %d -b %s", pr.Number, branchName))
	
	if !opts.NoPush {
		time.Sleep(500 * time.Millisecond)
		m.emit("info", "Pushing to origin...", "")
		m.emit("command", "", fmt.Sprintf("git push -u origin %s", branchName))
		m.emit("success", "Pushed to origin", "")
	}

	if !opts.NoCreate {
		m.emit("info", "\nCreate PR with:", "")
		createCmd := fmt.Sprintf(`gh pr create --title "%s" \
  --body "Migrated from #%d\nOriginal author: @%s" \
  --base %s`, pr.Title, pr.Number, pr.Author, pr.BaseBranch)
		m.emit("command", "", createCmd)
	}

	m.emit("success", "Migration complete!", "")
	return nil
}

// MigratePRs simulates migrating multiple PRs
func (m *MockMigrator) MigratePRs(prRefs []string, opts Options) error {
	for _, prRef := range prRefs {
		if err := m.MigratePR(prRef, opts); err != nil {
			return err
		}
	}
	return nil
}

// GetPRInfo simulates fetching PR information
func (m *MockMigrator) GetPRInfo(prRef string) (*PRInfo, error) {
	return m.parsePRRef(prRef)
}

// GenerateBranchName generates a branch name from PR info
func (m *MockMigrator) GenerateBranchName(pr *PRInfo) string {
	// Sanitize title for branch name
	slug := strings.ToLower(pr.Title)
	slug = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	
	// Limit to 40 characters
	if len(slug) > 40 {
		slug = slug[:40]
		slug = strings.TrimRight(slug, "-")
	}
	
	return fmt.Sprintf("pr-%d-%s-%s", pr.Number, strings.ToLower(pr.Author), slug)
}

// emit sends an event to the handler
func (m *MockMigrator) emit(eventType, message, detail string) {
	if m.handler != nil {
		m.handler(Event{
			Type:    eventType,
			Message: message,
			Detail:  detail,
		})
	}
}

// parsePRRef parses various PR reference formats
func (m *MockMigrator) parsePRRef(prRef string) (*PRInfo, error) {
	// Mock PR data
	mockPRs := map[int]*PRInfo{
		123: {
			Number:     123,
			Title:      "Fix memory leak in worker pool",
			Author:     "johndoe",
			HeadBranch: "fix-memory-leak",
			BaseBranch: "main",
			State:      "open",
			URL:        "https://github.com/owner/repo/pull/123",
			IsFork:     true,
		},
		124: {
			Number:     124,
			Title:      "Add new feature for data processing",
			Author:     "janedoe",
			HeadBranch: "feature-data-processing",
			BaseBranch: "main",
			State:      "open",
			URL:        "https://github.com/owner/repo/pull/124",
			IsFork:     true,
		},
		125: {
			Number:     125,
			Title:      "Update documentation",
			Author:     "alice",
			HeadBranch: "update-docs",
			BaseBranch: "main",
			State:      "open",
			URL:        "https://github.com/owner/repo/pull/125",
			IsFork:     false,
		},
	}

	// Parse different formats
	var prNumber int
	
	// Simple number
	if num, err := strconv.Atoi(prRef); err == nil {
		prNumber = num
	} else if strings.Contains(prRef, "#") {
		// owner/repo#123 format
		parts := strings.Split(prRef, "#")
		if len(parts) == 2 {
			if num, err := strconv.Atoi(parts[1]); err == nil {
				prNumber = num
			}
		}
	} else if strings.Contains(prRef, "github.com") {
		// Full URL format
		re := regexp.MustCompile(`/pull/(\d+)`)
		matches := re.FindStringSubmatch(prRef)
		if len(matches) == 2 {
			if num, err := strconv.Atoi(matches[1]); err == nil {
				prNumber = num
			}
		}
	}

	if prNumber == 0 {
		return nil, fmt.Errorf("invalid PR reference: %s", prRef)
	}

	pr, exists := mockPRs[prNumber]
	if !exists {
		return nil, fmt.Errorf("PR #%d not found", prNumber)
	}

	return pr, nil
}