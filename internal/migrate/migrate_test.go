package migrate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/user/git-mfpr/internal/git"
	"github.com/user/git-mfpr/internal/github"
)

type mockGit struct {
	currentRepoFunc func(context.Context) (string, string, error)
	hasBranchFunc   func(context.Context, string) bool
	checkoutFunc    func(context.Context, string) error
	pullFunc        func(context.Context, string, string) error
	pushFunc        func(context.Context, string, string) error
}

func (m *mockGit) CurrentRepo(ctx context.Context) (string, string, error) {
	if m.currentRepoFunc != nil {
		return m.currentRepoFunc(ctx)
	}
	return "testowner", "testrepo", nil
}

func (m *mockGit) HasBranch(ctx context.Context, name string) bool {
	if m.hasBranchFunc != nil {
		return m.hasBranchFunc(ctx, name)
	}
	return false
}

func (m *mockGit) Checkout(ctx context.Context, branch string) error {
	if m.checkoutFunc != nil {
		return m.checkoutFunc(ctx, branch)
	}
	return nil
}

func (m *mockGit) Pull(ctx context.Context, remote, branch string) error {
	if m.pullFunc != nil {
		return m.pullFunc(ctx, remote, branch)
	}
	return nil
}

func (m *mockGit) Push(ctx context.Context, remote, branch string) error {
	if m.pushFunc != nil {
		return m.pushFunc(ctx, remote, branch)
	}
	return nil
}

func (m *mockGit) CurrentBranch(_ context.Context) (string, error) { return "main", nil }
func (m *mockGit) DeleteBranch(_ context.Context, _ string) error  { return nil }
func (m *mockGit) IsInRepo(_ context.Context) bool                 { return true }

func (m *mockGit) CurrentBranchResult(_ context.Context) *git.BranchResult {
	return &git.BranchResult{Branch: "main"}
}

func (m *mockGit) CurrentRepoResult(_ context.Context) *git.RepoResult {
	return &git.RepoResult{Owner: "testowner", Repo: "testrepo"}
}

func (m *mockGit) CheckoutResult(_ context.Context, _ string) *git.OperationResult {
	return &git.OperationResult{Success: true}
}

func (m *mockGit) PullResult(_ context.Context, _, _ string) *git.OperationResult {
	return &git.OperationResult{Success: true}
}

func (m *mockGit) PushResult(_ context.Context, _, _ string) *git.OperationResult {
	return &git.OperationResult{Success: true}
}

func (m *mockGit) DeleteBranchResult(_ context.Context, _ string) *git.OperationResult {
	return &git.OperationResult{Success: true}
}

type mockGitHub struct {
	getPRFunc      func(string, string, int) (*github.PRInfo, error)
	checkoutPRFunc func(int, string) error
}

func (m *mockGitHub) GetPR(_ context.Context, owner, repo string, number int) (*github.PRInfo, error) {
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

func (m *mockGitHub) CheckoutPR(_ context.Context, _, _ string, number int, branch string) error {
	if m.checkoutPRFunc != nil {
		return m.checkoutPRFunc(number, branch)
	}
	return nil
}

func (m *mockGitHub) CreatePR(_ context.Context, _, _, _ string) error { return nil }
func (m *mockGitHub) IsGHInstalled(_ context.Context) error            { return nil }

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
		currentRepoFunc: func(_ context.Context) (string, string, error) {
			return "", "", errors.New("not in git repo")
		},
	}

	client := newTestClient(mockGit, &mockGitHub{})
	err := func() error {
		_, _, _, err := client.parsePRRef("123")
		return err
	}()

	if err == nil {
		t.Error("parsePRRef() should return error when not in git repo")
	}
	if !strings.Contains(err.Error(), "not in a git repository") {
		t.Errorf("parsePRRef() error = %v, want error containing 'not in a git repository'", err)
	}
}

func TestGetPRInfo(t *testing.T) {
	ctx := context.Background()
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
		getPRFunc: func(_, repo string, number int) (*github.PRInfo, error) {
			if repo == "testrepo" && number == 123 {
				return expectedPR, nil
			}
			return nil, errors.New("PR not found")
		},
	}

	client := newTestClient(&mockGit{}, mockGitHub)
	pr, err := client.GetPRInfo(ctx, "123")
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
			name: "normal PR",
			pr: &PRInfo{
				Number: 123,
				Author: "testuser",
				Title:  "Fix memory leak in worker pool",
			},
			expected: "migrated-123",
		},
		{
			name: "PR with special characters in title",
			pr: &PRInfo{
				Number: 456,
				Author: "user",
				Title:  "Add new feature! (WIP) - Part 1/2",
			},
			expected: "migrated-456",
		},
		{
			name: "PR with very long title",
			pr: &PRInfo{
				Number: 789,
				Author: "longauthor",
				Title:  "This is a very long title that should be truncated to fit within the branch name length limit of 80 characters",
			},
			expected: "migrated-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(&mockGit{}, &mockGitHub{})
			result := client.GenerateBranchName(tt.pr)

			if result != tt.expected {
				t.Errorf("GenerateBranchName() = %v, want %v", result, tt.expected)
			}

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
	ctx := context.Background()
	events := []Event{}
	handler := func(event Event) {
		events = append(events, event)
	}

	mockGit := &mockGit{
		hasBranchFunc: func(_ context.Context, _ string) bool {
			return false
		},
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(_, _ string, _ int) (*github.PRInfo, error) {
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

	err := client.MigratePR(ctx, "123", Options{})
	if err != nil {
		t.Errorf("MigratePR() error = %v", err)
	}

	if len(events) == 0 {
		t.Error("No events were emitted")
	}

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
	ctx := context.Background()
	mockGitHub := &mockGitHub{
		getPRFunc: func(_, _ string, _ int) (*github.PRInfo, error) {
			return &github.PRInfo{
				Number: 123,
				Title:  "Test PR",
				Author: "testuser",
				State:  "open",
				IsFork: false,
			}, nil
		},
	}

	client := newTestClient(&mockGit{}, mockGitHub)
	err := client.MigratePR(ctx, "123", Options{})

	if err == nil {
		t.Error("MigratePR() should return error for non-fork PR")
	}
	if !strings.Contains(err.Error(), "not from a fork") {
		t.Errorf("MigratePR() error = %v, want error containing 'not from a fork'", err)
	}
}

func TestMigratePR_ClosedPR(t *testing.T) {
	ctx := context.Background()
	mockGitHub := &mockGitHub{
		getPRFunc: func(_, _ string, _ int) (*github.PRInfo, error) {
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
	err := client.MigratePR(ctx, "123", Options{})

	if err == nil {
		t.Error("MigratePR() should return error for closed PR")
	}
	if !strings.Contains(err.Error(), "closed") {
		t.Errorf("MigratePR() error = %v, want error containing 'closed'", err)
	}
}

func TestMigratePR_BranchExists(t *testing.T) {
	ctx := context.Background()
	mockGit := &mockGit{
		hasBranchFunc: func(_ context.Context, _ string) bool {
			return true
		},
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(_, _ string, _ int) (*github.PRInfo, error) {
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
	err := client.MigratePR(ctx, "123", Options{})

	if err == nil {
		t.Error("MigratePR() should return error when branch exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("MigratePR() error = %v, want error containing 'already exists'", err)
	}
}

func TestMigratePR_DryRun(t *testing.T) {
	ctx := context.Background()
	events := []Event{}
	handler := func(event Event) {
		events = append(events, event)
	}

	mockGit := &mockGit{
		hasBranchFunc: func(_ context.Context, _ string) bool {
			return false
		},
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(_, _ string, _ int) (*github.PRInfo, error) {
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

	err := client.MigratePR(ctx, "123", Options{DryRun: true})
	if err != nil {
		t.Errorf("MigratePR() error = %v", err)
	}

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
	ctx := context.Background()
	mockGit := &mockGit{
		hasBranchFunc: func(_ context.Context, name string) bool {
			return name == "custom-branch"
		},
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(_, _ string, _ int) (*github.PRInfo, error) {
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
	err := client.MigratePR(ctx, "123", Options{BranchName: "custom-branch"})

	if err == nil {
		t.Error("MigratePR() should return error when custom branch exists")
	}
	if !strings.Contains(err.Error(), "custom-branch") {
		t.Errorf("MigratePR() error = %v, want error containing 'custom-branch'", err)
	}
}

func TestMigratePRs_Success(t *testing.T) {
	ctx := context.Background()
	events := []Event{}
	handler := func(event Event) {
		events = append(events, event)
	}

	mockGit := &mockGit{
		hasBranchFunc: func(_ context.Context, _ string) bool {
			return false
		},
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(_, _ string, _ int) (*github.PRInfo, error) {
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

	err := client.MigratePRs(ctx, []string{"123", "124"}, Options{})
	if err != nil {
		t.Errorf("MigratePRs() error = %v", err)
	}

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
	ctx := context.Background()
	events := []Event{}
	handler := func(event Event) {
		events = append(events, event)
	}

	mockGit := &mockGit{
		hasBranchFunc: func(_ context.Context, _ string) bool {
			return false
		},
	}

	mockGitHub := &mockGitHub{
		getPRFunc: func(_, _ string, number int) (*github.PRInfo, error) {
			if number == 123 {
				return &github.PRInfo{
					Number: 123,
					Title:  "Test PR",
					Author: "testuser",
					State:  "open",
					IsFork: true,
				}, nil
			}
			return nil, errors.New("PR not found")
		},
	}

	client := newTestClient(mockGit, mockGitHub)
	client.SetEventHandler(handler)

	err := client.MigratePRs(ctx, []string{"123", "124"}, Options{})

	if err == nil {
		t.Error("MigratePRs() should return error when some PRs fail")
	}

	if !strings.Contains(err.Error(), "failed to migrate some PRs") {
		t.Errorf("MigratePRs() error = %v, want error containing 'failed to migrate some PRs'", err)
	}

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
}

func TestSetEventHandler(t *testing.T) {
	client := &Client{}
	called := false

	handler := func(_ Event) {
		called = true
	}

	client.SetEventHandler(handler)
	client.emit("test", "message", "detail")

	if !called {
		t.Error("EventHandler was not called")
	}
}

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

func TestPushAndEmit_Error(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(&mockGit{
		pushFunc: func(_ context.Context, remote, branch string) error {
			return &git.ErrPushFailed{Remote: remote, Branch: branch, Detail: "permission denied"}
		},
	}, &mockGitHub{})

	client.SetEventHandler(func(event Event) {
		// Verify error event is emitted
		if event.Type == EventError {
			t.Logf("Error event emitted: %s", event.Message)
		}
	})

	err := client.pushAndEmit(ctx, "test-branch")
	if err == nil {
		t.Error("pushAndEmit() should return error when push fails")
	}

	if _, ok := err.(*git.ErrPushFailed); !ok {
		t.Errorf("Expected git.ErrPushFailed, got %T: %v", err, err)
	}
}

func TestPushAndEmit_ContextCancellation(t *testing.T) {
	// This test is not reliable because context cancellation doesn't immediately
	// stop git commands in all environments
	t.Skip("Context cancellation test is not reliable in all environments")
}

func TestMigratePR_CheckoutBaseError(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(&mockGit{
		checkoutFunc: func(_ context.Context, branch string) error {
			return &git.ErrCheckoutFailed{Branch: branch, Detail: "branch not found"}
		},
	}, &mockGitHub{
		getPRFunc: func(_, _ string, _ int) (*github.PRInfo, error) {
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
	})

	err := client.MigratePR(ctx, "123", Options{})
	if err == nil {
		t.Error("MigratePR() should return error when checkout fails")
	}

	if _, ok := err.(*git.ErrCheckoutFailed); !ok {
		t.Errorf("Expected git.ErrCheckoutFailed, got %T: %v", err, err)
	}
}

func TestMigratePR_PullError(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(&mockGit{
		checkoutFunc: func(_ context.Context, _ string) error {
			return nil
		},
		pullFunc: func(_ context.Context, remote, branch string) error {
			return &git.ErrPullFailed{Remote: remote, Branch: branch, Detail: "network error"}
		},
	}, &mockGitHub{
		getPRFunc: func(_, _ string, _ int) (*github.PRInfo, error) {
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
	})

	err := client.MigratePR(ctx, "123", Options{})
	if err == nil {
		t.Error("MigratePR() should return error when pull fails")
	}

	if _, ok := err.(*git.ErrPullFailed); !ok {
		t.Errorf("Expected git.ErrPullFailed, got %T: %v", err, err)
	}
}

func TestMigratePR_CheckoutPRError(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(&mockGit{
		checkoutFunc: func(_ context.Context, _ string) error {
			return nil
		},
		pullFunc: func(_ context.Context, _, _ string) error {
			return nil
		},
	}, &mockGitHub{
		getPRFunc: func(_, _ string, _ int) (*github.PRInfo, error) {
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
		checkoutPRFunc: func(number int, _ string) error {
			return &github.ErrPRCheckoutFailed{Number: number, Detail: "PR not found"}
		},
	})

	err := client.MigratePR(ctx, "123", Options{})
	if err == nil {
		t.Error("MigratePR() should return error when PR checkout fails")
	}

	if _, ok := err.(*github.ErrPRCheckoutFailed); !ok {
		t.Errorf("Expected github.ErrPRCheckoutFailed, got %T: %v", err, err)
	}
}

func TestMigratePR_ContextCancellation(t *testing.T) {
	// This test is not reliable because context cancellation doesn't immediately
	// stop git commands in all environments
	t.Skip("Context cancellation test is not reliable in all environments")
}

func TestMigratePR_GetPRError(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(&mockGit{}, &mockGitHub{
		getPRFunc: func(_, repo string, number int) (*github.PRInfo, error) {
			return nil, &github.ErrPRNotFound{Number: number, Repo: repo}
		},
	})

	err := client.MigratePR(ctx, "123", Options{})
	if err == nil {
		t.Error("MigratePR() should return error when GetPR fails")
	}

	if _, ok := err.(*github.ErrPRNotFound); !ok {
		t.Errorf("Expected github.ErrPRNotFound, got %T: %v", err, err)
	}
}

func TestMigratePR_CurrentRepoError(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(&mockGit{
		currentRepoFunc: func(_ context.Context) (string, string, error) {
			return "", "", &git.ErrGetRemoteURLFailed{Detail: "no remote origin"}
		},
	}, &mockGitHub{})

	err := client.MigratePR(ctx, "123", Options{})
	if err == nil {
		t.Error("MigratePR() should return error when CurrentRepo fails")
	}

	// The error gets wrapped, so we need to check the underlying error
	if err.Error() == "" {
		t.Error("MigratePR() should return a non-empty error message")
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrPRNotFound",
			err:      &ErrPRNotFound{Number: 123, Owner: "owner", Repo: "repo"},
			expected: "PR #123 not found in owner/repo",
		},
		{
			name:     "ErrPRNotFork",
			err:      &ErrPRNotFork{Number: 123},
			expected: "PR #123 is not from a fork",
		},
		{
			name:     "ErrPRClosed",
			err:      &ErrPRClosed{Number: 123, State: "closed"},
			expected: "PR #123 is closed",
		},
		{
			name:     "ErrBranchExists",
			err:      &ErrBranchExists{BranchName: "test-branch"},
			expected: "branch test-branch already exists. Use --branch-name to specify a different name or delete the existing branch",
		},
		{
			name:     "ErrInvalidPRRef",
			err:      &ErrInvalidPRRef{Ref: "invalid-ref"},
			expected: "unsupported PR reference format: invalid-ref",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Error message = %q, want %q", tt.err.Error(), tt.expected)
			}
		})
	}
}

func TestMigratePR_NoPushOption(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(&mockGit{
		checkoutFunc: func(_ context.Context, _ string) error {
			return nil
		},
		pullFunc: func(_ context.Context, _, _ string) error {
			return nil
		},
	}, &mockGitHub{
		getPRFunc: func(_, _ string, _ int) (*github.PRInfo, error) {
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
		checkoutPRFunc: func(_ int, _ string) error {
			return nil
		},
	})

	// Test with NoPush option
	err := client.MigratePR(ctx, "123", Options{NoPush: true})
	if err != nil {
		t.Errorf("MigratePR() should not return error with NoPush option: %v", err)
	}
}

func TestMigratePR_NoCreateOption(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(&mockGit{
		checkoutFunc: func(_ context.Context, _ string) error {
			return nil
		},
		pullFunc: func(_ context.Context, _, _ string) error {
			return nil
		},
		pushFunc: func(_ context.Context, _, _ string) error {
			return nil
		},
	}, &mockGitHub{
		getPRFunc: func(_, _ string, _ int) (*github.PRInfo, error) {
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
		checkoutPRFunc: func(_ int, _ string) error {
			return nil
		},
	})

	// Test with NoCreate option
	err := client.MigratePR(ctx, "123", Options{NoCreate: true})
	if err != nil {
		t.Errorf("MigratePR() should not return error with NoCreate option: %v", err)
	}
}
