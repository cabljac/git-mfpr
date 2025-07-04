package github

import (
	"strings"
	"testing"
)

func TestErrGHNotInstalled_Error(t *testing.T) {
	err := ErrGHNotInstalled{}
	expected := "gh CLI is not installed. Install it from https://cli.github.com"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestErrPRNotFound_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrPRNotFound
		expected string
	}{
		{
			name: "basic PR not found",
			err: ErrPRNotFound{
				Number: 123,
				Owner:  "testowner",
				Repo:   "testrepo",
			},
			expected: "PR #123 not found in testowner/testrepo",
		},
		{
			name: "different PR number",
			err: ErrPRNotFound{
				Number: 456,
				Owner:  "owner2",
				Repo:   "repo2",
			},
			expected: "PR #456 not found in owner2/repo2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Expected error message %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestErrPRFetchFailed_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrPRFetchFailed
		expected string
	}{
		{
			name: "network error",
			err: ErrPRFetchFailed{
				Number: 789,
				Owner:  "myorg",
				Repo:   "myrepo",
				Detail: "network timeout",
			},
			expected: "failed to fetch PR #789 from myorg/myrepo: network timeout",
		},
		{
			name: "auth error",
			err: ErrPRFetchFailed{
				Number: 321,
				Owner:  "private",
				Repo:   "repo",
				Detail: "authentication required",
			},
			expected: "failed to fetch PR #321 from private/repo: authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Expected error message %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestErrPRParseFailed_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrPRParseFailed
		contains string
	}{
		{
			name:     "invalid JSON",
			err:      ErrPRParseFailed{Detail: "invalid JSON format"},
			contains: "failed to parse PR data: invalid JSON format",
		},
		{
			name:     "missing field",
			err:      ErrPRParseFailed{Detail: "missing required field 'number'"},
			contains: "missing required field 'number'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			if !strings.Contains(errMsg, tt.contains) {
				t.Errorf("Expected error message to contain %q, got %q", tt.contains, errMsg)
			}
		})
	}
}

func TestErrPRCheckoutFailed_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrPRCheckoutFailed
		expected string
	}{
		{
			name: "branch conflict",
			err: ErrPRCheckoutFailed{
				Number: 555,
				Detail: "branch already exists",
			},
			expected: "failed to checkout PR #555: branch already exists",
		},
		{
			name: "permission denied",
			err: ErrPRCheckoutFailed{
				Number: 999,
				Detail: "permission denied",
			},
			expected: "failed to checkout PR #999: permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Expected error message %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestErrPRCreateFailed_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrPRCreateFailed
		contains string
	}{
		{
			name:     "validation error",
			err:      ErrPRCreateFailed{Detail: "title cannot be empty"},
			contains: "failed to create PR: title cannot be empty",
		},
		{
			name:     "API limit",
			err:      ErrPRCreateFailed{Detail: "API rate limit exceeded"},
			contains: "API rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			if !strings.Contains(errMsg, tt.contains) {
				t.Errorf("Expected error message to contain %q, got %q", tt.contains, errMsg)
			}
		})
	}
}

// Test that errors implement the error interface
func TestErrorsImplementErrorInterface(_ *testing.T) {
	var _ error = ErrGHNotInstalled{}
	var _ error = ErrPRNotFound{}
	var _ error = ErrPRFetchFailed{}
	var _ error = ErrPRParseFailed{}
	var _ error = ErrPRCheckoutFailed{}
	var _ error = ErrPRCreateFailed{}
}

// Benchmark error message generation
func BenchmarkErrPRNotFound_Error(b *testing.B) {
	err := ErrPRNotFound{
		Number: 12345,
		Owner:  "benchowner",
		Repo:   "benchrepo",
	}

	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkErrPRFetchFailed_Error(b *testing.B) {
	err := ErrPRFetchFailed{
		Number: 67890,
		Owner:  "testorg",
		Repo:   "testrepo",
		Detail: "connection refused",
	}

	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
