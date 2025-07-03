package migrate

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/user/git-mfpr/internal/git"
	"github.com/user/git-mfpr/internal/github"
)

// Mock implementations for testing
type mockGit struct {
	currentRepoFunc func() (string, string, error)
	hasBranchFunc   func(string) bool
	checkoutFunc    func(string) error
	pullFunc        func(string, string) error
	pushFunc        func(string, string) error
}

func (m *mockGit) CurrentRepo() (string, string, error) {
	if m.currentRepoFunc != nil {
		return m.currentRepoFunc()
	}
	return "testowner", "testrepo", nil
}

func (m *mockGit) HasBranch(name string) bool {
	if m.hasBranchFunc != nil {
		return m.hasBranchFunc(name)
	}
	return false
}

func (m *mockGit) Checkout(branch string) error {
	if m.checkoutFunc != nil {
		return m.checkoutFunc(branch)
	}
	return nil
}

func (m *mockGit) Pull(remote, branch string) error {
	if m.pullFunc != nil {
		return m.pullFunc(remote, branch)
	}
	return nil
}

func (m *mockGit) Push(remote, branch string) error {
	if m.pushFunc != nil {
		return m.pushFunc(remote, branch)
	}
	return nil
}

func (m *mockGit) CurrentBranch() (string, error) { return "main", nil }
func (m *mockGit) DeleteBranch(name string) error { return nil }
func (m *mockGit) IsInRepo() bool                 { return true }

type mockGitHub struct {
	getPRFunc      func(string, string, int) (*github.PRInfo, error)
	checkoutPRFunc func(int, string) error
}

func (m *mockGitHub) GetPR(owner, repo string, number int) (*github.PRInfo, error) {
	if m.getPRFunc != nil {
		return m.getPRFunc(owner, repo, number)
	}
	return &github.PRInfo{
		Number:     123,
		Title:      "Test PR",
		Author:     "testuser",
		HeadBranch: "feature-branch",
		BaseBranch: "main",
		State:      "open",
		URL:        "https://github.com/testowner/testrepo/pull/123",
		IsFork:     true,
	}, nil
}

func (m *mockGitHub) CheckoutPR(number int, branch string) error {
	if m.checkoutPRFunc != nil {
		return m.checkoutPRFunc(number, branch)
	}
	return nil
}

func (m *mockGitHub) CreatePR(title, body, base string) error { return nil }
func (m *mockGitHub) IsGHInstalled() error                    { return nil }

// Test helper to create a test client
func newTestClient(git git.Git, github github.GitHub) *Client {
	return &Client{
		git:     git,
		github:  github,
		handler: func(Event) {},
	}
}

func TestParsePRRef(t *testing.T) {
	tests := []struct {
		name        string
		prRef       string
		wantOwner   string
		wantRepo    string
		wantNumber  int
		wantErr     bool
		errContains string
	}{
		{
			name:       "simple number",
			prRef:      "123",
			wantOwner:  "testowner",
			wantRepo:   "testrepo",
			wantNumber: 123,
		},
		{
			name:       "owner/repo#number",
			prRef:      "owner/repo#456",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantNumber: 456,
		},
		{
			name:       "full URL",
			prRef:      "https://github.com/owner/repo/pull/789",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantNumber: 789,
		},
		{
			name:        "invalid number",
			prRef:       "abc",
			wantErr:     true,
			errContains: "unsupported PR reference format",
		},
		{
			name:        "invalid owner/repo format",
			prRef:       "invalid#123",
			wantErr:     true,
			errContains: "invalid repo format",
		},
		{
			name:        "invalid URL format",
			prRef:       "https://github.com/owner/repo/issues/123",
			wantErr:     true,
			errContains: "invalid GitHub PR URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(&mockGit{}, &mockGitHub{})
			owner, repo, number, err := client.parsePRRef(tt.prRef)

			if (err != nil) != tt.wantErr {
				t.Errorf("parsePRRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parsePRRef() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if owner != tt.wantOwner {
				t.Errorf("parsePRRef() owner = %v, want %v", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("parsePRRef() repo = %v, want %v", repo, tt.wantRepo)
			}
			if number != tt.wantNumber {
				t.Errorf("parsePRRef() number = %v, want %v", number, tt.wantNumber)
			}
		})
	}
}

func TestParsePRRef_CurrentRepoError(t *testing.T) {
	mockGit := &mockGit{
		currentRepoFunc: func() (string, string, error) {
			return "", "", errors.New("not in git repo")
		},
	}

	client := newTestClient(mockGit, &mockGitHub{})
	_, _, _, err := client.parsePRRef("123")

	if err == nil {
		t.Error("parsePRRef() should return error when not in git repo")
	}
	if !strings.Contains(err.Error(), "not in a git repository") {
		t.Errorf("parsePRRef() error = %v, want error containing 'not in a git repository'", err)
	}
}

func TestGetPRInfo(t *testing.T) {
	expectedPR := &github.PRInfo{
		Number:     123,
		Title:      "Test PR",
		Author:     "testuser",
		HeadBranch: "feature-branch",
		BaseBranch: "main",
		State:      "open",
		URL:        "https://github.com/testowner/testrepo/pull/123",
		IsFork:     true,
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(owner, repo string, number int) (*github.PRInfo, error) {
			if owner == "testowner" && repo == "testrepo" && number == 123 {
				return expectedPR, nil
			}
			return nil, errors.New("PR not found")
		},
	}

	client := newTestClient(&mockGit{}, mockGitHub)
	pr, err := client.GetPRInfo("123")

	if err != nil {
		t.Errorf("GetPRInfo() error = %v", err)
		return
	}

	if pr.Number != expectedPR.Number {
		t.Errorf("GetPRInfo() number = %v, want %v", pr.Number, expectedPR.Number)
	}
	if pr.Title != expectedPR.Title {
		t.Errorf("GetPRInfo() title = %v, want %v", pr.Title, expectedPR.Title)
	}
	if pr.Author != expectedPR.Author {
		t.Errorf("GetPRInfo() author = %v, want %v", pr.Author, expectedPR.Author)
	}
}

func TestGenerateBranchName(t *testing.T) {
	tests := []struct {
		name     string
		pr       *PRInfo
		expected string
	}{
		{
			name: "normal title",
			pr: &PRInfo{
				Number: 123,
				Author: "testuser",
				Title:  "Fix memory leak in worker pool",
			},
			expected: "pr-123-testuser-fix-memory-leak-in-worker-pool",
		},
		{
			name: "title with special characters",
			pr: &PRInfo{
				Number: 456,
				Author: "user",
				Title:  "Add new feature! (WIP) - Part 1/2",
			},
			expected: "pr-456-user-add-new-feature-wip-part-1-2",
		},
		{
			name: "very long title",
			pr: &PRInfo{
				Number: 789,
				Author: "longauthor",
				Title:  "This is a very long title that should be truncated to fit within the branch name length limit of 80 characters",
			},
			expected: "pr-789-longauthor-this-is-a-very-long-title-that-should-be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(&mockGit{}, &mockGitHub{})
			result := client.GenerateBranchName(tt.pr)

			if result != tt.expected {
				t.Errorf("GenerateBranchName() = %v, want %v", result, tt.expected)
			}

			// Ensure branch name is not too long
			if len(result) > 80 {
				t.Errorf("GenerateBranchName() result too long: %d characters", len(result))
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"Fix bug #123", "fix-bug-123"},
		{"Add new feature!", "add-new-feature"},
		{"Update documentation (WIP)", "update-documentation-wip"},
		{"Very long title that should be truncated to fit within the limit", "very-long-title-that-should-be-truncated"},
		{"Multiple   spaces", "multiple-spaces"},
		{"Special@#$%^&*()chars", "special-chars"},
		{"", ""},
		{"---leading-trailing---", "leading-trailing"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := slugify(tt.input)
			if result != tt.expected {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 1, -1},
		{0, 10, 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d,%d", tt.a, tt.b), func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestMigratePR_Success(t *testing.T) {
	events := []Event{}
	handler := func(event Event) {
		events = append(events, event)
	}

	mockGit := &mockGit{
		hasBranchFunc: func(name string) bool {
			return false // Branch doesn't exist
		},
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(owner, repo string, number int) (*github.PRInfo, error) {
			return &github.PRInfo{
				Number:     123,
				Title:      "Test PR",
				Author:     "testuser",
				HeadBranch: "feature-branch",
				BaseBranch: "main",
				State:      "open",
				URL:        "https://github.com/testowner/testrepo/pull/123",
				IsFork:     true,
			}, nil
		},
	}

	client := newTestClient(mockGit, mockGitHub)
	client.SetEventHandler(handler)

	err := client.MigratePR("123", Options{})

	if err != nil {
		t.Errorf("MigratePR() error = %v", err)
	}

	// Check that events were emitted
	if len(events) == 0 {
		t.Error("No events were emitted")
	}

	// Check for specific events
	hasInfoEvent := false
	hasSuccessEvent := false
	for _, event := range events {
		if event.Type == "info" && strings.Contains(event.Message, "Migrating PR #123") {
			hasInfoEvent = true
		}
		if event.Type == "success" && strings.Contains(event.Message, "Successfully migrated PR #123") {
			hasSuccessEvent = true
		}
	}

	if !hasInfoEvent {
		t.Error("Expected info event for migration start")
	}
	if !hasSuccessEvent {
		t.Error("Expected success event for migration completion")
	}
}

func TestMigratePR_NotFork(t *testing.T) {
	mockGitHub := &mockGitHub{
		getPRFunc: func(owner, repo string, number int) (*github.PRInfo, error) {
			return &github.PRInfo{
				Number: 123,
				Title:  "Test PR",
				Author: "testuser",
				State:  "open",
				IsFork: false, // Not a fork
			}, nil
		},
	}

	client := newTestClient(&mockGit{}, mockGitHub)
	err := client.MigratePR("123", Options{})

	if err == nil {
		t.Error("MigratePR() should return error for non-fork PR")
	}
	if !strings.Contains(err.Error(), "not from a fork") {
		t.Errorf("MigratePR() error = %v, want error containing 'not from a fork'", err)
	}
}

func TestMigratePR_ClosedPR(t *testing.T) {
	mockGitHub := &mockGitHub{
		getPRFunc: func(owner, repo string, number int) (*github.PRInfo, error) {
			return &github.PRInfo{
				Number: 123,
				Title:  "Test PR",
				Author: "testuser",
				State:  "closed", // Closed PR
				IsFork: true,
			}, nil
		},
	}

	client := newTestClient(&mockGit{}, mockGitHub)
	err := client.MigratePR("123", Options{})

	if err == nil {
		t.Error("MigratePR() should return error for closed PR")
	}
	if !strings.Contains(err.Error(), "closed") {
		t.Errorf("MigratePR() error = %v, want error containing 'closed'", err)
	}
}

func TestMigratePR_BranchExists(t *testing.T) {
	mockGit := &mockGit{
		hasBranchFunc: func(name string) bool {
			return true // Branch already exists
		},
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(owner, repo string, number int) (*github.PRInfo, error) {
			return &github.PRInfo{
				Number: 123,
				Title:  "Test PR",
				Author: "testuser",
				State:  "open",
				IsFork: true,
			}, nil
		},
	}

	client := newTestClient(mockGit, mockGitHub)
	err := client.MigratePR("123", Options{})

	if err == nil {
		t.Error("MigratePR() should return error when branch exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("MigratePR() error = %v, want error containing 'already exists'", err)
	}
}

func TestMigratePR_DryRun(t *testing.T) {
	events := []Event{}
	handler := func(event Event) {
		events = append(events, event)
	}

	mockGit := &mockGit{
		hasBranchFunc: func(name string) bool {
			return false
		},
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(owner, repo string, number int) (*github.PRInfo, error) {
			return &github.PRInfo{
				Number:     123,
				Title:      "Test PR",
				Author:     "testuser",
				HeadBranch: "feature-branch",
				BaseBranch: "main",
				State:      "open",
				IsFork:     true,
			}, nil
		},
	}

	client := newTestClient(mockGit, mockGitHub)
	client.SetEventHandler(handler)

	err := client.MigratePR("123", Options{DryRun: true})

	if err != nil {
		t.Errorf("MigratePR() error = %v", err)
	}

	// Check for dry run commands
	hasDryRunCommand := false
	for _, event := range events {
		if event.Type == "command" && strings.Contains(event.Message, "Would execute:") {
			hasDryRunCommand = true
			break
		}
	}

	if !hasDryRunCommand {
		t.Error("Expected dry run commands to be emitted")
	}
}

func TestMigratePR_CustomBranchName(t *testing.T) {
	mockGit := &mockGit{
		hasBranchFunc: func(name string) bool {
			return name == "custom-branch" // Only this branch exists
		},
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(owner, repo string, number int) (*github.PRInfo, error) {
			return &github.PRInfo{
				Number: 123,
				Title:  "Test PR",
				Author: "testuser",
				State:  "open",
				IsFork: true,
			}, nil
		},
	}

	client := newTestClient(mockGit, mockGitHub)
	err := client.MigratePR("123", Options{BranchName: "custom-branch"})

	if err == nil {
		t.Error("MigratePR() should return error when custom branch exists")
	}
	if !strings.Contains(err.Error(), "custom-branch") {
		t.Errorf("MigratePR() error = %v, want error containing 'custom-branch'", err)
	}
}

func TestMigratePRs_Success(t *testing.T) {
	events := []Event{}
	handler := func(event Event) {
		events = append(events, event)
	}

	mockGit := &mockGit{
		hasBranchFunc: func(name string) bool {
			return false
		},
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(owner, repo string, number int) (*github.PRInfo, error) {
			return &github.PRInfo{
				Number: 123,
				Title:  "Test PR",
				Author: "testuser",
				State:  "open",
				IsFork: true,
			}, nil
		},
	}

	client := newTestClient(mockGit, mockGitHub)
	client.SetEventHandler(handler)

	err := client.MigratePRs([]string{"123", "124"}, Options{})

	if err != nil {
		t.Errorf("MigratePRs() error = %v", err)
	}

	// Check for success message
	hasSuccessMessage := false
	for _, event := range events {
		if event.Type == "success" && strings.Contains(event.Message, "Successfully migrated all 2 PRs") {
			hasSuccessMessage = true
			break
		}
	}

	if !hasSuccessMessage {
		t.Error("Expected success message for batch migration")
	}
}

func TestMigratePRs_PartialFailure(t *testing.T) {
	events := []Event{}
	handler := func(event Event) {
		events = append(events, event)
	}

	mockGit := &mockGit{
		hasBranchFunc: func(name string) bool {
			return false
		},
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(owner, repo string, number int) (*github.PRInfo, error) {
			if number == 123 {
				return &github.PRInfo{
					Number: 123,
					Title:  "Test PR",
					Author: "testuser",
					State:  "open",
					IsFork: true,
				}, nil
			}
			return nil, errors.New("PR not found") // PR 124 fails
		},
	}

	client := newTestClient(mockGit, mockGitHub)
	client.SetEventHandler(handler)

	err := client.MigratePRs([]string{"123", "124"}, Options{})

	if err == nil {
		t.Error("MigratePRs() should return error when some PRs fail")
	}

	if !strings.Contains(err.Error(), "failed to migrate some PRs") {
		t.Errorf("MigratePRs() error = %v, want error containing 'failed to migrate some PRs'", err)
	}

	// Check for partial success message
	hasPartialSuccess := false
	for _, event := range events {
		if event.Type == "info" && strings.Contains(event.Message, "Migrated 1/2 PRs successfully") {
			hasPartialSuccess = true
			break
		}
	}

	if !hasPartialSuccess {
		t.Error("Expected partial success message")
	}
}

func TestNew(t *testing.T) {
	migrator := New()
	if migrator == nil {
		t.Error("New() returned nil")
	}

	// Test that it implements the Migrator interface
	_, ok := migrator.(Migrator)
	if !ok {
		t.Error("New() does not implement Migrator interface")
	}
}

func TestSetEventHandler(t *testing.T) {
	client := &Client{}
	called := false

	handler := func(event Event) {
		called = true
	}

	client.SetEventHandler(handler)
	client.emit("test", "message", "detail")

	if !called {
		t.Error("EventHandler was not called")
	}
}

// Benchmark tests
func BenchmarkGenerateBranchName(b *testing.B) {
	pr := &PRInfo{
		Number: 123,
		Author: "testuser",
		Title:  "Fix memory leak in worker pool with comprehensive error handling",
	}

	client := newTestClient(&mockGit{}, &mockGitHub{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.GenerateBranchName(pr)
	}
}

func BenchmarkSlugify(b *testing.B) {
	input := "This is a very long title with special characters! @#$%^&*() and numbers 12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slugify(input)
	}
}
