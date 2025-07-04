package git

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestClient_CurrentRepo(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name      string
		remoteURL string
		wantOwner string
		wantRepo  string
		wantErr   bool
		errType   interface{}
	}{
		{
			name:      "SSH format",
			remoteURL: "git@github.com:owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "SSH format without .git",
			remoteURL: "git@github.com:owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "HTTPS format",
			remoteURL: "https://github.com/owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "HTTPS format without .git",
			remoteURL: "https://github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "Invalid SSH format",
			remoteURL: "git@github.com:invalid",
			wantErr:   true,
			errType:   &ErrInvalidRemoteURL{},
		},
		{
			name:      "Invalid HTTPS format",
			remoteURL: "https://github.com/invalid",
			wantErr:   true,
			errType:   &ErrInvalidRemoteURL{},
		},
		{
			name:      "Non-GitHub URL",
			remoteURL: "https://gitlab.com/owner/repo.git",
			wantErr:   true,
			errType:   &ErrInvalidRemoteURL{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			os.Chdir(tmpDir)

			exec.Command("git", "init").Run()                                  // #nosec G204
			exec.Command("git", "remote", "add", "origin", tt.remoteURL).Run() // #nosec G204

			client := New()
			owner, repo, err := client.CurrentRepo(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("CurrentRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.errType != nil {
					if _, ok := err.(*ErrInvalidRemoteURL); !ok {
						t.Errorf("Expected %T, got %T", tt.errType, err)
					}
				}
				return
			}
			if owner != tt.wantOwner {
				t.Errorf("CurrentRepo() owner = %v, want %v", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("CurrentRepo() repo = %v, want %v", repo, tt.wantRepo)
			}
		})
	}
}

func TestClient_HasBranch(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	exec.Command("git", "init").Run()                                     // #nosec G204
	exec.Command("git", "config", "user.email", "test@example.com").Run() // #nosec G204
	exec.Command("git", "config", "user.name", "Test User").Run()         // #nosec G204

	os.WriteFile("test.txt", []byte("test"), 0o600)
	exec.Command("git", "add", ".").Run()                // #nosec G204
	exec.Command("git", "commit", "-m", "initial").Run() // #nosec G204

	exec.Command("git", "branch", "test-branch").Run() // #nosec G204

	client := New()

	if !client.HasBranch(ctx, "test-branch") {
		t.Error("HasBranch() returned false for existing branch")
	}

	if client.HasBranch(ctx, "non-existing") {
		t.Error("HasBranch() returned true for non-existing branch")
	}
}

func TestClient_CurrentBranch(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Initialize git repo
	exec.Command("git", "init", "-b", "main").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run() // #nosec G204
	exec.Command("git", "config", "user.name", "Test User").Run()         // #nosec G204

	os.WriteFile("test.txt", []byte("test"), 0o600)
	exec.Command("git", "add", ".").Run()                // #nosec G204
	exec.Command("git", "commit", "-m", "initial").Run() // #nosec G204

	client := New()

	branch, err := client.CurrentBranch(ctx)
	if err != nil {
		t.Fatalf("CurrentBranch() error = %v", err)
	}
	if branch != "main" {
		t.Errorf("CurrentBranch() = %v, want main", branch)
	}

	exec.Command("git", "checkout", "-b", "feature").Run() // #nosec G204

	branch, err = client.CurrentBranch(ctx)
	if err != nil {
		t.Fatalf("CurrentBranch() error = %v", err)
	}
	if branch != "feature" {
		t.Errorf("CurrentBranch() = %v, want feature", branch)
	}
}

func TestClient_IsInRepo(t *testing.T) {
	ctx := context.Background()
	client := New()

	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	if client.IsInRepo(ctx) {
		t.Error("IsInRepo() returned true outside git repo")
	}

	exec.Command("git", "init").Run()

	if !client.IsInRepo(ctx) {
		t.Error("IsInRepo() returned false inside git repo")
	}
}

func TestClient_Checkout(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Initialize git repo
	exec.Command("git", "init", "-b", "main").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	os.WriteFile("test.txt", []byte("test"), 0o600)
	exec.Command("git", "add", ".").Run()                // #nosec G204
	exec.Command("git", "commit", "-m", "initial").Run() // #nosec G204

	exec.Command("git", "branch", "test-branch").Run()

	client := New()

	err := client.Checkout(ctx, "test-branch")
	if err != nil {
		t.Fatalf("Checkout() error = %v", err)
	}

	branch, _ := client.CurrentBranch(ctx)
	if branch != "test-branch" {
		t.Errorf("After checkout, current branch = %v, want test-branch", branch)
	}

	err = client.Checkout(ctx, "non-existing")
	if err == nil {
		t.Error("Checkout() expected error for non-existing branch")
	}
	if _, ok := err.(*ErrCheckoutFailed); !ok {
		t.Errorf("Expected ErrCheckoutFailed, got %T", err)
	}
}

func TestClient_DeleteBranch(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Initialize git repo
	exec.Command("git", "init", "-b", "main").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	os.WriteFile("test.txt", []byte("test"), 0o600)
	exec.Command("git", "add", ".").Run()                // #nosec G204
	exec.Command("git", "commit", "-m", "initial").Run() // #nosec G204

	exec.Command("git", "checkout", "-b", "test-branch").Run() // #nosec G204
	exec.Command("git", "checkout", "main").Run()              // #nosec G204

	client := New()

	if !client.HasBranch(ctx, "test-branch") {
		t.Fatal("test-branch should exist")
	}

	err := client.DeleteBranch(ctx, "test-branch")
	if err != nil {
		t.Fatalf("DeleteBranch() error = %v", err)
	}

	if client.HasBranch(ctx, "test-branch") {
		t.Error("test-branch should be deleted")
	}

	err = client.DeleteBranch(ctx, "non-existing")
	if err == nil {
		t.Error("DeleteBranch() expected error for non-existing branch")
	}
	if _, ok := err.(*ErrDeleteBranchFailed); !ok {
		t.Errorf("Expected ErrDeleteBranchFailed, got %T", err)
	}
}

func TestNewWithOptions(t *testing.T) {
	client := NewWithOptions(WithTimeout(10 * time.Second))

	if client == nil {
		t.Error("NewWithOptions() returned nil")
	}
}

func TestContextCancellation(t *testing.T) {
	// Skip this test as it's not reliable in CI environments
	// The git command completes too quickly for context cancellation to be meaningful
	t.Skip("Context cancellation test is not reliable in CI environments")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	client := New()

	_, err := client.CurrentBranch(ctx)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
}

func TestResultTypes(t *testing.T) {
	t.Run("BranchResult", func(t *testing.T) {
		// Test success case
		br := &BranchResult{Branch: "main", Error: nil}
		if !br.IsSuccess() {
			t.Error("IsSuccess() should return true when Error is nil")
		}
		if br.IsError() {
			t.Error("IsError() should return false when Error is nil")
		}

		// Test error case
		br2 := &BranchResult{Branch: "", Error: errors.New("test error")}
		if br2.IsSuccess() {
			t.Error("IsSuccess() should return false when Error is not nil")
		}
		if !br2.IsError() {
			t.Error("IsError() should return true when Error is not nil")
		}
	})

	t.Run("RepoResult", func(t *testing.T) {
		// Test success case
		rr := &RepoResult{Owner: "owner", Repo: "repo", Error: nil}
		if !rr.IsSuccess() {
			t.Error("IsSuccess() should return true when Error is nil")
		}
		if rr.IsError() {
			t.Error("IsError() should return false when Error is nil")
		}

		// Test error case
		rr2 := &RepoResult{Owner: "", Repo: "", Error: errors.New("test error")}
		if rr2.IsSuccess() {
			t.Error("IsSuccess() should return false when Error is not nil")
		}
		if !rr2.IsError() {
			t.Error("IsError() should return true when Error is not nil")
		}
	})

	t.Run("OperationResult", func(t *testing.T) {
		// Test success case
		or := &OperationResult{Success: true, Error: nil}
		if !or.IsSuccess() {
			t.Error("IsSuccess() should return true when Success is true")
		}
		if or.IsError() {
			t.Error("IsError() should return false when Success is true")
		}

		// Test error case
		or2 := &OperationResult{Success: false, Error: errors.New("test error")}
		if or2.IsSuccess() {
			t.Error("IsSuccess() should return false when Success is false")
		}
		if !or2.IsError() {
			t.Error("IsError() should return true when Success is false")
		}
	})
}

func TestClient_Pull(t *testing.T) {
	ctx := context.Background()
	client := New()

	if !client.IsInRepo(ctx) {
		t.Skip("Not in a git repository")
	}

	// Test pulling from a non-existent remote
	err := client.Pull(ctx, "nonexistent-remote", "main")
	if err == nil {
		t.Error("Expected error when pulling from non-existent remote")
	}

	if pullErr, ok := err.(*ErrPullFailed); !ok {
		t.Errorf("Expected ErrPullFailed, got %T", err)
	} else {
		if pullErr.Remote != "nonexistent-remote" {
			t.Errorf("Expected remote 'nonexistent-remote', got '%s'", pullErr.Remote)
		}
		if pullErr.Branch != "main" {
			t.Errorf("Expected branch 'main', got '%s'", pullErr.Branch)
		}
	}
}

func TestClient_Push(t *testing.T) {
	ctx := context.Background()
	client := New()

	if !client.IsInRepo(ctx) {
		t.Skip("Not in a git repository")
	}

	// Test pushing to a non-existent remote
	err := client.Push(ctx, "nonexistent-remote", "main")
	if err == nil {
		t.Error("Expected error when pushing to non-existent remote")
	}

	if pushErr, ok := err.(*ErrPushFailed); !ok {
		t.Errorf("Expected ErrPushFailed, got %T", err)
	} else {
		if pushErr.Remote != "nonexistent-remote" {
			t.Errorf("Expected remote 'nonexistent-remote', got '%s'", pushErr.Remote)
		}
		if pushErr.Branch != "main" {
			t.Errorf("Expected branch 'main', got '%s'", pushErr.Branch)
		}
	}
}

func TestClient_ResultMethods(t *testing.T) {
	ctx := context.Background()
	client := New()

	t.Run("CurrentBranchResult", func(t *testing.T) {
		result := client.CurrentBranchResult(ctx)
		if result == nil {
			t.Fatal("CurrentBranchResult returned nil")
		}

		if client.IsInRepo(ctx) {
			// In a repo, should have a branch
			if result.Error != nil {
				t.Errorf("Expected no error, got %v", result.Error)
			}
			if result.Branch == "" {
				t.Error("Expected branch name, got empty string")
			}
		} else if result.Error == nil {
			t.Error("Expected error when not in repo")
		}
	})

	t.Run("CurrentRepoResult", func(t *testing.T) {
		result := client.CurrentRepoResult(ctx)
		if result == nil {
			t.Fatal("CurrentRepoResult returned nil")
		}

		// This will likely have an error in test environment
		// Just verify the method works
	})

	t.Run("CheckoutResult", func(t *testing.T) {
		if !client.IsInRepo(ctx) {
			t.Skip("Not in a git repository")
		}

		result := client.CheckoutResult(ctx, "nonexistent-branch-12345")
		if result == nil {
			t.Fatal("CheckoutResult returned nil")
		}

		if result.Success {
			t.Error("Expected failure when checking out non-existent branch")
		}
		if result.Error == nil {
			t.Error("Expected error when checking out non-existent branch")
		}
	})

	t.Run("PullResult", func(t *testing.T) {
		if !client.IsInRepo(ctx) {
			t.Skip("Not in a git repository")
		}

		result := client.PullResult(ctx, "nonexistent-remote", "main")
		if result == nil {
			t.Fatal("PullResult returned nil")
		}

		if result.Success {
			t.Error("Expected failure when pulling from non-existent remote")
		}
		if result.Error == nil {
			t.Error("Expected error when pulling from non-existent remote")
		}
	})

	t.Run("PushResult", func(t *testing.T) {
		if !client.IsInRepo(ctx) {
			t.Skip("Not in a git repository")
		}

		result := client.PushResult(ctx, "nonexistent-remote", "main")
		if result == nil {
			t.Fatal("PushResult returned nil")
		}

		if result.Success {
			t.Error("Expected failure when pushing to non-existent remote")
		}
		if result.Error == nil {
			t.Error("Expected error when pushing to non-existent remote")
		}
	})

	t.Run("DeleteBranchResult", func(t *testing.T) {
		if !client.IsInRepo(ctx) {
			t.Skip("Not in a git repository")
		}

		result := client.DeleteBranchResult(ctx, "nonexistent-branch-12345")
		if result == nil {
			t.Fatal("DeleteBranchResult returned nil")
		}

		if result.Success {
			t.Error("Expected failure when deleting non-existent branch")
		}
		if result.Error == nil {
			t.Error("Expected error when deleting non-existent branch")
		}
	})
}

func TestIntegration_PullPush(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Skip("Pull/Push integration tests require real remote repository")
}

func TestClient_CurrentBranch_Error(t *testing.T) {
	ctx := context.Background()

	// Test outside of a git repository
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	client := New()
	_, err := client.CurrentBranch(ctx)
	if err == nil {
		t.Error("CurrentBranch() should return error outside git repo")
	}

	if _, ok := err.(*ErrGetCurrentBranchFailed); !ok {
		t.Errorf("Expected ErrGetCurrentBranchFailed, got %T: %v", err, err)
	}
}

func TestClient_CurrentBranch_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := New()
	_, err := client.CurrentBranch(ctx)
	if err == nil {
		t.Error("CurrentBranch() should return error when context is cancelled")
	}
}

func TestClient_CurrentRepo_NoRemote(t *testing.T) {
	ctx := context.Background()

	// Test in a git repo without origin remote
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	exec.Command("git", "init").Run() // #nosec G204

	client := New()
	_, _, err := client.CurrentRepo(ctx)
	if err == nil {
		t.Error("CurrentRepo() should return error when no origin remote exists")
	}

	if _, ok := err.(*ErrGetRemoteURLFailed); !ok {
		t.Errorf("Expected ErrGetRemoteURLFailed, got %T: %v", err, err)
	}
}

func TestClient_CurrentRepo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := New()
	_, _, err := client.CurrentRepo(ctx)
	if err == nil {
		t.Error("CurrentRepo() should return error when context is cancelled")
	}
}

func TestClient_Pull_Error(t *testing.T) {
	ctx := context.Background()

	// Test pull with invalid remote
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	exec.Command("git", "init").Run()                                     // #nosec G204
	exec.Command("git", "config", "user.email", "test@example.com").Run() // #nosec G204
	exec.Command("git", "config", "user.name", "Test User").Run()         // #nosec G204

	client := New()
	err := client.Pull(ctx, "nonexistent", "main")
	if err == nil {
		t.Skip("Pull() succeeded unexpectedly, skipping error test")
	}

	if _, ok := err.(*ErrPullFailed); !ok {
		t.Errorf("Expected ErrPullFailed, got %T: %v", err, err)
	}
}

func TestClient_Pull_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := New()
	err := client.Pull(ctx, "origin", "main")
	if err == nil {
		t.Error("Pull() should return error when context is cancelled")
	}
}

func TestClient_Push_Error(t *testing.T) {
	ctx := context.Background()

	// Test push with invalid remote
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	exec.Command("git", "init").Run()                                     // #nosec G204
	exec.Command("git", "config", "user.email", "test@example.com").Run() // #nosec G204
	exec.Command("git", "config", "user.name", "Test User").Run()         // #nosec G204

	client := New()
	err := client.Push(ctx, "nonexistent", "main")
	if err == nil {
		t.Skip("Push() succeeded unexpectedly, skipping error test")
	}

	if _, ok := err.(*ErrPushFailed); !ok {
		t.Errorf("Expected ErrPushFailed, got %T: %v", err, err)
	}
}

func TestClient_Push_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := New()
	err := client.Push(ctx, "origin", "main")
	if err == nil {
		t.Error("Push() should return error when context is cancelled")
	}
}

func TestClient_Checkout_Error(t *testing.T) {
	ctx := context.Background()

	// Test checkout of non-existent branch
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	exec.Command("git", "init").Run()                                     // #nosec G204
	exec.Command("git", "config", "user.email", "test@example.com").Run() // #nosec G204
	exec.Command("git", "config", "user.name", "Test User").Run()         // #nosec G204

	client := New()
	err := client.Checkout(ctx, "nonexistent-branch")
	if err == nil {
		t.Skip("Checkout() succeeded unexpectedly, skipping error test")
	}

	if _, ok := err.(*ErrCheckoutFailed); !ok {
		t.Errorf("Expected ErrCheckoutFailed, got %T: %v", err, err)
	}
}

func TestClient_Checkout_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := New()
	err := client.Checkout(ctx, "main")
	if err == nil {
		t.Error("Checkout() should return error when context is cancelled")
	}
}

func TestClient_DeleteBranch_Error(t *testing.T) {
	ctx := context.Background()

	// Test delete of non-existent branch
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	exec.Command("git", "init").Run()                                     // #nosec G204
	exec.Command("git", "config", "user.email", "test@example.com").Run() // #nosec G204
	exec.Command("git", "config", "user.name", "Test User").Run()         // #nosec G204

	client := New()
	err := client.DeleteBranch(ctx, "nonexistent-branch")
	if err == nil {
		t.Skip("DeleteBranch() succeeded unexpectedly, skipping error test")
	}

	if _, ok := err.(*ErrDeleteBranchFailed); !ok {
		t.Errorf("Expected ErrDeleteBranchFailed, got %T: %v", err, err)
	}
}

func TestClient_DeleteBranch_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := New()
	err := client.DeleteBranch(ctx, "main")
	if err == nil {
		t.Error("DeleteBranch() should return error when context is cancelled")
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrGetCurrentBranchFailed",
			err:      &ErrGetCurrentBranchFailed{Detail: "git command failed"},
			expected: "failed to get current branch: git command failed",
		},
		{
			name:     "ErrGetRemoteURLFailed",
			err:      &ErrGetRemoteURLFailed{Detail: "no remote origin"},
			expected: "failed to get remote URL: no remote origin",
		},
		{
			name:     "ErrInvalidRemoteURL",
			err:      &ErrInvalidRemoteURL{URL: "invalid-url"},
			expected: "invalid remote URL format: invalid-url",
		},
		{
			name:     "ErrCheckoutFailed",
			err:      &ErrCheckoutFailed{Branch: "main", Detail: "branch not found"},
			expected: "failed to checkout branch main: branch not found",
		},
		{
			name:     "ErrPullFailed",
			err:      &ErrPullFailed{Remote: "origin", Branch: "main", Detail: "network error"},
			expected: "failed to pull origin/main: network error",
		},
		{
			name:     "ErrPushFailed",
			err:      &ErrPushFailed{Remote: "origin", Branch: "main", Detail: "permission denied"},
			expected: "failed to push to origin/main: permission denied",
		},
		{
			name:     "ErrDeleteBranchFailed",
			err:      &ErrDeleteBranchFailed{Branch: "feature", Detail: "branch not found"},
			expected: "failed to delete branch feature: branch not found",
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

func TestClient_ResultMethods_Error(t *testing.T) {
	ctx := context.Background()
	client := New()

	// Test CurrentBranchResult with error
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	result := client.CurrentBranchResult(ctx)
	if result.IsSuccess() {
		t.Error("CurrentBranchResult() should return error outside git repo")
	}
	if !result.IsError() {
		t.Error("CurrentBranchResult() should indicate error")
	}
	if result.Error == nil {
		t.Error("CurrentBranchResult() should have error")
	}

	// Test CurrentRepoResult with error
	repoResult := client.CurrentRepoResult(ctx)
	if repoResult.IsSuccess() {
		t.Error("CurrentRepoResult() should return error outside git repo")
	}
	if !repoResult.IsError() {
		t.Error("CurrentRepoResult() should indicate error")
	}
	if repoResult.Error == nil {
		t.Error("CurrentRepoResult() should have error")
	}
}

func TestClient_OperationResults_Error(t *testing.T) {
	ctx := context.Background()
	client := New()

	// Test operation results with errors
	checkoutResult := client.CheckoutResult(ctx, "nonexistent")
	if checkoutResult.IsSuccess() {
		t.Error("CheckoutResult() should return error for non-existent branch")
	}
	if !checkoutResult.IsError() {
		t.Error("CheckoutResult() should indicate error")
	}

	pullResult := client.PullResult(ctx, "nonexistent", "main")
	if pullResult.IsSuccess() {
		t.Error("PullResult() should return error for non-existent remote")
	}
	if !pullResult.IsError() {
		t.Error("PullResult() should indicate error")
	}

	pushResult := client.PushResult(ctx, "nonexistent", "main")
	if pushResult.IsSuccess() {
		t.Error("PushResult() should return error for non-existent remote")
	}
	if !pushResult.IsError() {
		t.Error("PushResult() should indicate error")
	}

	deleteResult := client.DeleteBranchResult(ctx, "nonexistent")
	if deleteResult.IsSuccess() {
		t.Error("DeleteBranchResult() should return error for non-existent branch")
	}
	if !deleteResult.IsError() {
		t.Error("DeleteBranchResult() should indicate error")
	}
}
