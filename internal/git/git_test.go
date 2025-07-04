package git

import (
	"context"
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

			exec.Command("git", "init").Run()
			exec.Command("git", "remote", "add", "origin", tt.remoteURL).Run()

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

	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	exec.Command("git", "branch", "test-branch").Run()

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
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	client := New()

	branch, err := client.CurrentBranch(ctx)
	if err != nil {
		t.Fatalf("CurrentBranch() error = %v", err)
	}
	if branch != "main" {
		t.Errorf("CurrentBranch() = %v, want main", branch)
	}

	exec.Command("git", "checkout", "-b", "feature").Run()

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

	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

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

	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	exec.Command("git", "checkout", "-b", "test-branch").Run()
	exec.Command("git", "checkout", "main").Run()

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

func TestResultTypes(t *testing.T) {
	ctx := context.Background()
	client := New()

	branchResult := client.CurrentBranchResult(ctx)
	if branchResult == nil {
		t.Error("CurrentBranchResult() returned nil")
	}

	repoResult := client.CurrentRepoResult(ctx)
	if repoResult == nil {
		t.Error("CurrentRepoResult() returned nil")
	}

	opResult := client.CheckoutResult(ctx, "main")
	if opResult == nil {
		t.Error("CheckoutResult() returned nil")
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	client := New()

	_, err := client.CurrentBranch(ctx)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
}

func TestIntegration_PullPush(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Skip("Pull/Push integration tests require real remote repository")
}
