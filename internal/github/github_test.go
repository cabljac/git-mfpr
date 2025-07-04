package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

var mockCommand = exec.Command

func TestClient_IsGHInstalled(t *testing.T) {
	ctx := context.Background()
	client := New()

	err := client.IsGHInstalled(ctx)

	if _, checkErr := exec.LookPath("gh"); checkErr != nil {
		if err == nil {
			t.Error("IsGHInstalled() should return error when gh is not installed")
		}
		if _, ok := err.(*ErrGHNotInstalled); !ok {
			t.Errorf("Expected ErrGHNotInstalled, got %T", err)
		}
	} else {
		if err != nil {
			t.Errorf("IsGHInstalled() error = %v", err)
		}
	}
}

func TestClient_IsGHInstalled_WithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := NewWithOptions(WithTimeout(5 * time.Second))
	err := client.IsGHInstalled(ctx)

	if err != nil && !isErrGHNotInstalled(err) {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPRInfoJSON(t *testing.T) {
	jsonData := `{
		"number": 123,
		"title": "Fix memory leak",
		"state": "open",
		"headRefName": "feature-branch",
		"baseRefName": "main",
		"headRefOid": "abc123",
		"isCrossRepository": true,
		"url": "https://github.com/owner/repo/pull/123",
		"author": {
			"login": "johndoe"
		}
	}`

	var pr ghPRResponse
	err := json.Unmarshal([]byte(jsonData), &pr)
	if err != nil {
		t.Fatalf("Failed to unmarshal PR data: %v", err)
	}

	if pr.Number != 123 {
		t.Errorf("Number = %d, want 123", pr.Number)
	}
	if pr.Title != "Fix memory leak" {
		t.Errorf("Title = %s, want 'Fix memory leak'", pr.Title)
	}
	if pr.Author.Login != "johndoe" {
		t.Errorf("Author = %s, want 'johndoe'", pr.Author.Login)
	}
	if !pr.IsCrossRepository {
		t.Error("IsCrossRepository = false, want true")
	}
}

func TestGetPR_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	os.Setenv("PATH", "")

	client := New()
	_, err := client.GetPR(ctx, "owner", "repo", 123)

	if err == nil {
		t.Error("GetPR() should return error when gh is not installed")
	}

	if !isErrGHNotInstalled(err) {
		t.Errorf("Expected ErrGHNotInstalled, got %T: %v", err, err)
	}
}

func TestNewWithOptions(t *testing.T) {
	client := NewWithOptions(WithTimeout(10 * time.Second))

	if client == nil {
		t.Error("NewWithOptions() returned nil")
	}
}

func TestPRResult(t *testing.T) {
	pr := &PRInfo{Number: 123, Title: "Test PR"}
	result := &PRResult{PR: pr, Error: nil}

	if !result.IsSuccess() {
		t.Error("IsSuccess() should return true for successful result")
	}

	if result.IsError() {
		t.Error("IsError() should return false for successful result")
	}

	errResult := &PRResult{PR: nil, Error: &ErrPRNotFound{Number: 123, Owner: "owner", Repo: "repo"}}

	if errResult.IsSuccess() {
		t.Error("IsSuccess() should return false for error result")
	}

	if !errResult.IsError() {
		t.Error("IsError() should return true for error result")
	}
}

func isErrGHNotInstalled(err error) bool {
	_, ok := err.(*ErrGHNotInstalled)
	return ok
}

func TestIntegration_GetPR(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not installed, skipping integration test")
	}

	t.Skip("GetPR integration test requires real GitHub repository")
}

func TestGetPR_ParseErrors(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantErr bool
	}{
		{
			name:    "Invalid JSON",
			output:  "invalid json",
			wantErr: true,
		},
		{
			name: "Valid JSON",
			output: `{
				"number": 123,
				"title": "Test PR",
				"state": "open",
				"headRefName": "feature",
				"baseRefName": "main",
				"headRefOid": "abc123",
				"isCrossRepository": true,
				"url": "https://github.com/owner/repo/pull/123",
				"author": {"login": "user"}
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pr ghPRResponse
			err := json.Unmarshal([]byte(tt.output), &pr)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_CheckoutPR(t *testing.T) {
	ctx := context.Background()
	client := New()

	// This test will only pass if gh CLI is installed and can be mocked
	// Since we can't easily mock exec.Command in this test, we'll skip if gh is not installed
	if err := client.IsGHInstalled(ctx); err != nil {
		t.Skip("gh CLI not installed, skipping CheckoutPR test")
	}

	// We can't actually test this without side effects, so we'll just verify the error handling
	err := client.CheckoutPR(ctx, "testowner", "testrepo", 99999, "test-branch")
	if err == nil {
		t.Error("Expected error for non-existent PR, got nil")
	}

	if _, ok := err.(*ErrPRCheckoutFailed); !ok {
		t.Errorf("Expected ErrPRCheckoutFailed, got %T", err)
	}
}

func TestClient_CreatePR(t *testing.T) {
	ctx := context.Background()
	client := New()

	// This test will only pass if gh CLI is installed
	if err := client.IsGHInstalled(ctx); err != nil {
		t.Skip("gh CLI not installed, skipping CreatePR test")
	}

	// We can't actually create a PR in tests, so we'll just verify the method exists
	// and returns an error when not in a git repo or without proper setup
	err := client.CreatePR(ctx, "Test PR", "Test body", "main")
	if err == nil {
		// If no error, we might be in a real repo with gh auth, which we don't want
		t.Skip("Skipping to avoid creating real PR")
	}

	// Verify it returns the correct error type
	if _, ok := err.(*ErrPRCreateFailed); !ok {
		t.Errorf("Expected ErrPRCreateFailed, got %T", err)
	}
}

func BenchmarkPRInfoParsing(b *testing.B) {
	jsonData := `{
		"number": 123,
		"title": "Fix memory leak in worker pool",
		"state": "open",
		"headRefName": "fix-memory-leak",
		"baseRefName": "main",
		"headRefOid": "abc123def456",
		"isCrossRepository": true,
		"url": "https://github.com/owner/repo/pull/123",
		"author": {
			"login": "johndoe"
		}
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var pr ghPRResponse
		json.Unmarshal([]byte(jsonData), &pr)
	}
}

func ExampleClient_GetPR() {
	ctx := context.Background()
	client := New()

	pr, err := client.GetPR(ctx, "owner", "repo", 123)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("PR #%d: %s by @%s\n", pr.Number, pr.Title, pr.Author)
}

func ExampleNewWithOptions() {
	ctx := context.Background()
	client := NewWithOptions(WithTimeout(10 * time.Second))

	pr, err := client.GetPR(ctx, "owner", "repo", 123)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("PR #%d: %s\n", pr.Number, pr.Title)
}
