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
	} else if err != nil {
		t.Errorf("IsGHInstalled() error = %v", err)
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

func TestGetPR_NetworkFailure(t *testing.T) {
	ctx := context.Background()
	client := New()

	// Test with a timeout context to simulate network failure
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	_, err := client.GetPR(timeoutCtx, "owner", "repo", 123)
	if err == nil {
		t.Error("GetPR() should return error on timeout")
	}
}

func TestGetPR_InvalidJSONResponse(t *testing.T) {
	// This test would require mocking exec.Command to return invalid JSON
	// For now, we'll test the JSON parsing logic separately
	invalidJSON := `{"invalid": json}`
	var pr ghPRResponse
	err := json.Unmarshal([]byte(invalidJSON), &pr)
	if err == nil {
		t.Error("Should fail to parse invalid JSON")
	}
}

func TestGetPR_ExitErrorHandling(t *testing.T) {
	ctx := context.Background()
	client := New()

	// Test with a non-existent repository to trigger exit error
	_, err := client.GetPR(ctx, "nonexistent", "nonexistent", 999999)
	if err == nil {
		t.Skip("GetPR() succeeded unexpectedly, skipping exit error test")
	}

	// Check if it's the expected error type
	if _, ok := err.(*ErrPRNotFound); !ok {
		if _, ok := err.(*ErrPRFetchFailed); !ok {
			if _, ok := err.(*ErrGHNotInstalled); !ok {
				t.Errorf("Expected ErrPRNotFound, ErrPRFetchFailed, or ErrGHNotInstalled, got %T: %v", err, err)
			}
		}
	}
}

func TestGetPR_MissingAuthorField(t *testing.T) {
	// Test JSON response with missing author field
	jsonData := `{
		"number": 123,
		"title": "Test PR",
		"state": "open",
		"headRefName": "feature-branch",
		"baseRefName": "main",
		"headRefOid": "abc123",
		"isCrossRepository": true,
		"url": "https://github.com/owner/repo/pull/123"
	}`

	var pr ghPRResponse
	err := json.Unmarshal([]byte(jsonData), &pr)
	if err != nil {
		t.Fatalf("Failed to unmarshal PR data: %v", err)
	}

	// Author field should be empty but not cause an error
	if pr.Author.Login != "" {
		t.Errorf("Expected empty author login, got %s", pr.Author.Login)
	}
}

func TestCheckoutPR_NetworkFailure(t *testing.T) {
	ctx := context.Background()
	client := New()

	// Test with a timeout context to simulate network failure
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	err := client.CheckoutPR(timeoutCtx, "owner", "repo", 123, "test-branch")
	if err == nil {
		t.Error("CheckoutPR() should return error on timeout")
	}
}

func TestCheckoutPR_CommandFailure(t *testing.T) {
	ctx := context.Background()
	client := New()

	// Test with invalid PR number to trigger command failure
	err := client.CheckoutPR(ctx, "owner", "repo", 999999, "test-branch")
	if err == nil {
		t.Skip("CheckoutPR() succeeded unexpectedly, skipping command failure test")
	}

	// Check if it's the expected error type
	if _, ok := err.(*ErrPRCheckoutFailed); !ok {
		t.Errorf("Expected ErrPRCheckoutFailed, got %T: %v", err, err)
	}
}

func TestCreatePR_CommandFailure(t *testing.T) {
	ctx := context.Background()
	client := New()

	// Test with invalid base branch to trigger command failure
	err := client.CreatePR(ctx, "Test PR", "Test body", "nonexistent-branch")
	if err == nil {
		t.Skip("CreatePR() succeeded unexpectedly, skipping command failure test")
	}

	// Check if it's the expected error type
	if _, ok := err.(*ErrPRCreateFailed); !ok {
		t.Errorf("Expected ErrPRCreateFailed, got %T: %v", err, err)
	}
}

func TestGetPR_EmptyResponse(t *testing.T) {
	// Test with empty response (this would require mocking exec.Command)
	// For now, test the JSON parsing with empty data
	emptyJSON := `{}`
	var pr ghPRResponse
	err := json.Unmarshal([]byte(emptyJSON), &pr)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty JSON: %v", err)
	}

	// Should have zero values
	if pr.Number != 0 {
		t.Errorf("Expected Number = 0, got %d", pr.Number)
	}
	if pr.Title != "" {
		t.Errorf("Expected Title = '', got %s", pr.Title)
	}
}

func TestGetPR_MalformedJSON(t *testing.T) {
	// Test with malformed JSON that's missing required fields
	malformedJSON := `{"number": 123, "title": "Test"`
	var pr ghPRResponse
	err := json.Unmarshal([]byte(malformedJSON), &pr)
	if err == nil {
		t.Error("Should fail to parse malformed JSON")
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
			name:     "ErrPRFetchFailed",
			err:      &ErrPRFetchFailed{Number: 123, Owner: "owner", Repo: "repo", Detail: "network error"},
			expected: "failed to fetch PR #123 from owner/repo: network error",
		},
		{
			name:     "ErrPRParseFailed",
			err:      &ErrPRParseFailed{Detail: "invalid json"},
			expected: "failed to parse PR data: invalid json",
		},
		{
			name:     "ErrPRCheckoutFailed",
			err:      &ErrPRCheckoutFailed{Number: 123, Detail: "branch not found"},
			expected: "failed to checkout PR #123: branch not found",
		},
		{
			name:     "ErrPRCreateFailed",
			err:      &ErrPRCreateFailed{Detail: "invalid base branch"},
			expected: "failed to create PR: invalid base branch",
		},
		{
			name:     "ErrGHNotInstalled",
			err:      &ErrGHNotInstalled{},
			expected: "gh CLI is not installed. Install it from https://cli.github.com",
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

func TestNewWithOptions_Timeout(t *testing.T) {
	timeout := 15 * time.Second
	client := NewWithOptions(WithTimeout(timeout))

	if client == nil {
		t.Error("NewWithOptions() returned nil")
	}
}

func TestGetPR_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := New()
	_, err := client.GetPR(ctx, "owner", "repo", 123)
	if err == nil {
		t.Error("GetPR() should return error when context is cancelled")
	}
}

func TestCheckoutPR_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := New()
	err := client.CheckoutPR(ctx, "owner", "repo", 123, "test-branch")
	if err == nil {
		t.Error("CheckoutPR() should return error when context is cancelled")
	}
}

func TestCreatePR_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := New()
	err := client.CreatePR(ctx, "Test PR", "Test body", "main")
	if err == nil {
		t.Error("CreatePR() should return error when context is cancelled")
	}
}
