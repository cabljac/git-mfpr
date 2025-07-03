package git

import (
	"os"
	"os/exec"
	"testing"
)

// TestClient tests the Git client implementation
func TestClient_CurrentRepo(t *testing.T) {
	tests := []struct {
		name       string
		remoteURL  string
		wantOwner  string
		wantRepo   string
		wantErr    bool
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
		},
		{
			name:      "Invalid HTTPS format",
			remoteURL: "https://github.com/invalid",
			wantErr:   true,
		},
		{
			name:      "Non-GitHub URL",
			remoteURL: "https://gitlab.com/owner/repo.git",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()
			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			os.Chdir(tmpDir)

			// Initialize git repo
			exec.Command("git", "init").Run()
			exec.Command("git", "remote", "add", "origin", tt.remoteURL).Run()

			// Test
			client := New()
			owner, repo, err := client.CurrentRepo()

			if (err != nil) != tt.wantErr {
				t.Errorf("CurrentRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("CurrentRepo() owner = %v, want %v", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("CurrentRepo() repo = %v, want %v", repo, tt.wantRepo)
				}
			}
		})
	}
}

func TestClient_HasBranch(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Initialize git repo
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	
	// Create initial commit
	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create test branch
	exec.Command("git", "branch", "test-branch").Run()

	client := New()

	// Test existing branch
	if !client.HasBranch("test-branch") {
		t.Error("HasBranch() returned false for existing branch")
	}

	// Test non-existing branch
	if client.HasBranch("non-existing") {
		t.Error("HasBranch() returned true for non-existing branch")
	}
}

func TestClient_CurrentBranch(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Initialize git repo
	exec.Command("git", "init", "-b", "main").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	
	// Create initial commit
	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	client := New()

	// Test current branch
	branch, err := client.CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch() error = %v", err)
	}
	if branch != "main" {
		t.Errorf("CurrentBranch() = %v, want main", branch)
	}

	// Switch branch
	exec.Command("git", "checkout", "-b", "feature").Run()
	
	branch, err = client.CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch() error = %v", err)
	}
	if branch != "feature" {
		t.Errorf("CurrentBranch() = %v, want feature", branch)
	}
}

func TestClient_IsInRepo(t *testing.T) {
	client := New()

	// Test in a git repo
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Not in repo yet
	if client.IsInRepo() {
		t.Error("IsInRepo() returned true outside git repo")
	}

	// Initialize repo
	exec.Command("git", "init").Run()

	// Now in repo
	if !client.IsInRepo() {
		t.Error("IsInRepo() returned false inside git repo")
	}
}

func TestClient_Checkout(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Initialize git repo
	exec.Command("git", "init", "-b", "main").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	
	// Create initial commit
	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create test branch
	exec.Command("git", "branch", "test-branch").Run()

	client := New()

	// Test checkout
	err := client.Checkout("test-branch")
	if err != nil {
		t.Fatalf("Checkout() error = %v", err)
	}

	// Verify we're on the right branch
	branch, _ := client.CurrentBranch()
	if branch != "test-branch" {
		t.Errorf("After checkout, current branch = %v, want test-branch", branch)
	}

	// Test checkout non-existing branch
	err = client.Checkout("non-existing")
	if err == nil {
		t.Error("Checkout() expected error for non-existing branch")
	}
}

func TestClient_DeleteBranch(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Initialize git repo
	exec.Command("git", "init", "-b", "main").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	
	// Create initial commit
	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create and checkout test branch
	exec.Command("git", "checkout", "-b", "test-branch").Run()
	exec.Command("git", "checkout", "main").Run()

	client := New()

	// Verify branch exists
	if !client.HasBranch("test-branch") {
		t.Fatal("test-branch should exist")
	}

	// Delete branch
	err := client.DeleteBranch("test-branch")
	if err != nil {
		t.Fatalf("DeleteBranch() error = %v", err)
	}

	// Verify branch is gone
	if client.HasBranch("test-branch") {
		t.Error("test-branch should be deleted")
	}

	// Test deleting non-existing branch
	err = client.DeleteBranch("non-existing")
	if err == nil {
		t.Error("DeleteBranch() expected error for non-existing branch")
	}
}

// TestIntegration_PullPush tests pull and push operations
// This test is skipped by default as it requires network access
func TestIntegration_PullPush(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This test would require a real remote repository
	// For unit tests, we're skipping this
	t.Skip("Pull/Push integration tests require real remote repository")
}