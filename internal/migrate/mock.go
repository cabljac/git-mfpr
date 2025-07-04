package migrate

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type MockMigrator struct {
	handler EventHandler
}

func NewMockMigrator(handler EventHandler) Migrator {
	return &MockMigrator{handler: handler}
}

func (m *MockMigrator) SetEventHandler(handler EventHandler) {
	m.handler = handler
}

func (m *MockMigrator) MigratePR(ctx context.Context, prRef string, opts Options) error {
	pr, err := m.parsePRRef(prRef)
	if err != nil {
		return err
	}

	m.emit(EventInfo, "Fetching PR information...", "")
	time.Sleep(500 * time.Millisecond)

	m.emit(EventInfo, fmt.Sprintf("Title: %s", pr.Title), "")
	m.emit(EventInfo, fmt.Sprintf("Author: %s", pr.Author), "")
	m.emit(EventInfo, fmt.Sprintf("Base branch: %s", pr.BaseBranch), "")

	if pr.IsFork {
		m.emit(EventInfo, "PR is from a fork", "")
	}

	branchName := opts.BranchName
	if branchName == "" {
		branchName = m.GenerateBranchName(pr)
	}
	m.emit(EventInfo, fmt.Sprintf("Branch: %s", branchName), "")

	time.Sleep(300 * time.Millisecond)

	m.emit(EventCommand, "", fmt.Sprintf("git checkout %s", pr.BaseBranch))
	m.emit(EventCommand, "", "git pull origin main")

	time.Sleep(500 * time.Millisecond)
	m.emit(EventInfo, "Creating local branch...", "")
	m.emit(EventCommand, "", fmt.Sprintf("gh pr checkout %d -b %s", pr.Number, branchName))

	if !opts.NoPush {
		time.Sleep(500 * time.Millisecond)
		m.emit(EventInfo, "Pushing to origin...", "")
		m.emit(EventCommand, "", fmt.Sprintf("git push -u origin %s", branchName))
		m.emit(EventSuccess, "Pushed to origin", "")
	}

	if !opts.NoCreate {
		m.emit(EventInfo, "\nCreate PR with:", "")
		createCmd := fmt.Sprintf(`gh pr create --title "%s" \
  --body "Migrated from #%d\nOriginal author: @%s" \
  --base %s`, pr.Title, pr.Number, pr.Author, pr.BaseBranch)
		m.emit(EventCommand, "", createCmd)
	}

	m.emit(EventSuccess, "Migration complete!", "")
	return nil
}

func (m *MockMigrator) MigratePRs(ctx context.Context, prRefs []string, opts Options) error {
	for _, prRef := range prRefs {
		if err := m.MigratePR(ctx, prRef, opts); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockMigrator) GetPRInfo(ctx context.Context, prRef string) (*PRInfo, error) {
	return m.parsePRRef(prRef)
}

func (m *MockMigrator) GenerateBranchName(pr *PRInfo) string {
	return fmt.Sprintf("migrated-%d", pr.Number)
}

func (m *MockMigrator) emit(eventType EventType, message, detail string) {
	if m.handler != nil {
		m.handler(Event{
			Type:    eventType,
			Message: message,
			Detail:  detail,
		})
	}
}

func (m *MockMigrator) parsePRRef(prRef string) (*PRInfo, error) {
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

	var prNumber int

	if num, err := strconv.Atoi(prRef); err == nil {
		prNumber = num
	} else if strings.Contains(prRef, "#") {
		parts := strings.Split(prRef, "#")
		if len(parts) == 2 {
			if num, err := strconv.Atoi(parts[1]); err == nil {
				prNumber = num
			}
		}
	} else if strings.Contains(prRef, "github.com") {
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
