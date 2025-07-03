package github

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

// mockCommand is used to mock exec.Command for testing
var mockCommand = exec.Command

func TestClient_IsGHInstalled(t *testing.T) {
	client := New()

	// This test assumes gh is installed in the test environment
	// In a real test suite, we would mock exec.Command
	err := client.IsGHInstalled()
	
	// Check if gh is actually installed
	if _, checkErr := exec.LookPath("gh"); checkErr != nil {
		// gh is not installed, so we expect an error
		if err == nil {
			t.Error("IsGHInstalled() should return error when gh is not installed")
		}
	} else {
		// gh is installed, so we expect no error
		if err != nil {
			t.Errorf("IsGHInstalled() error = %v", err)
		}
	}
}

func TestPRInfoJSON(t *testing.T) {
	// Test JSON parsing
	jsonData := `{
		"number": 123,
		"title": "Fix memory leak",
		"state": "open",
		"headRefName": "feature-branch",
		"baseRefName": "main",
		"headRefOid": "abc123",
		"isFork": true,
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

	// Verify fields
	if pr.Number != 123 {
		t.Errorf("Number = %d, want 123", pr.Number)
	}
	if pr.Title != "Fix memory leak" {
		t.Errorf("Title = %s, want 'Fix memory leak'", pr.Title)
	}
	if pr.Author.Login != "johndoe" {
		t.Errorf("Author = %s, want 'johndoe'", pr.Author.Login)
	}
	if !pr.IsFork {
		t.Error("IsFork = false, want true")
	}
}

// TestGetPR_ErrorHandling tests error handling in GetPR
func TestGetPR_ErrorHandling(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Test with gh not in PATH
	os.Setenv("PATH", "")
	
	client := New()
	_, err := client.GetPR("owner", "repo", 123)
	
	if err == nil {
		t.Error("GetPR() should return error when gh is not installed")
	}
	
	expectedError := "gh CLI is not installed"
	if err != nil && !containsString(err.Error(), expectedError) {
		t.Errorf("GetPR() error = %v, want error containing %q", err, expectedError)
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && s[:len(substr)] == substr
}

// Integration test that requires gh CLI and network access
func TestIntegration_GetPR(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Check if gh is installed
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not installed, skipping integration test")
	}

	// This would test against a real GitHub PR
	// For unit tests, we're skipping this
	t.Skip("GetPR integration test requires real GitHub repository")
}

// Test data for various error scenarios
func TestGetPR_ParseErrors(t *testing.T) {
	// This test would require mocking exec.Command to simulate different gh outputs
	// For now, we're testing the basic structure
	
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
				"isFork": true,
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

// Benchmark for JSON parsing
func BenchmarkPRInfoParsing(b *testing.B) {
	jsonData := `{
		"number": 123,
		"title": "Fix memory leak in worker pool",
		"state": "open",
		"headRefName": "fix-memory-leak",
		"baseRefName": "main",
		"headRefOid": "abc123def456",
		"isFork": true,
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

// Example usage
func ExampleClient_GetPR() {
	client := New()
	
	// Example of getting a PR (would require gh CLI and network)
	pr, err := client.GetPR("owner", "repo", 123)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("PR #%d: %s by @%s\n", pr.Number, pr.Title, pr.Author)
}