package git

import (
	"strings"
	"testing"
)

func TestErrNotInRepo_Error(t *testing.T) {
	err := ErrNotInRepo{}
	expected := "not in a git repository"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestErrInvalidRemoteURL_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrInvalidRemoteURL
		expected string
	}{
		{
			name:     "empty URL",
			err:      ErrInvalidRemoteURL{URL: ""},
			expected: "invalid remote URL format: ",
		},
		{
			name:     "malformed URL",
			err:      ErrInvalidRemoteURL{URL: "not-a-valid-url"},
			expected: "invalid remote URL format: not-a-valid-url",
		},
		{
			name:     "missing protocol",
			err:      ErrInvalidRemoteURL{URL: "github.com/user/repo"},
			expected: "invalid remote URL format: github.com/user/repo",
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

func TestErrBranchNotFound_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrBranchNotFound
		expected string
	}{
		{
			name:     "feature branch",
			err:      ErrBranchNotFound{Branch: "feature/new-feature"},
			expected: "branch feature/new-feature not found",
		},
		{
			name:     "main branch",
			err:      ErrBranchNotFound{Branch: "main"},
			expected: "branch main not found",
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

func TestErrCheckoutFailed_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrCheckoutFailed
		expected string
	}{
		{
			name: "uncommitted changes",
			err: ErrCheckoutFailed{
				Branch: "develop",
				Detail: "uncommitted changes in working directory",
			},
			expected: "failed to checkout branch develop: uncommitted changes in working directory",
		},
		{
			name: "branch conflict",
			err: ErrCheckoutFailed{
				Branch: "hotfix/urgent",
				Detail: "branch name conflicts with existing file",
			},
			expected: "failed to checkout branch hotfix/urgent: branch name conflicts with existing file",
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

func TestErrPullFailed_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrPullFailed
		expected string
	}{
		{
			name: "merge conflict",
			err: ErrPullFailed{
				Remote: "origin",
				Branch: "main",
				Detail: "merge conflict detected",
			},
			expected: "failed to pull origin/main: merge conflict detected",
		},
		{
			name: "network error",
			err: ErrPullFailed{
				Remote: "upstream",
				Branch: "develop",
				Detail: "connection timeout",
			},
			expected: "failed to pull upstream/develop: connection timeout",
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

func TestErrPushFailed_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrPushFailed
		expected string
	}{
		{
			name: "permission denied",
			err: ErrPushFailed{
				Remote: "origin",
				Branch: "feature/protected",
				Detail: "permission denied",
			},
			expected: "failed to push to origin/feature/protected: permission denied",
		},
		{
			name: "non-fast-forward",
			err: ErrPushFailed{
				Remote: "origin",
				Branch: "main",
				Detail: "non-fast-forward update",
			},
			expected: "failed to push to origin/main: non-fast-forward update",
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

func TestErrDeleteBranchFailed_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrDeleteBranchFailed
		contains string
	}{
		{
			name: "not fully merged",
			err: ErrDeleteBranchFailed{
				Branch: "feature/incomplete",
				Detail: "branch is not fully merged",
			},
			contains: "failed to delete branch feature/incomplete: branch is not fully merged",
		},
		{
			name: "current branch",
			err: ErrDeleteBranchFailed{
				Branch: "main",
				Detail: "cannot delete current branch",
			},
			contains: "cannot delete current branch",
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

func TestErrGetCurrentBranchFailed_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrGetCurrentBranchFailed
		expected string
	}{
		{
			name:     "detached HEAD",
			err:      ErrGetCurrentBranchFailed{Detail: "HEAD is detached"},
			expected: "failed to get current branch: HEAD is detached",
		},
		{
			name:     "corrupted git",
			err:      ErrGetCurrentBranchFailed{Detail: "corrupted git directory"},
			expected: "failed to get current branch: corrupted git directory",
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

func TestErrGetRemoteURLFailed_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrGetRemoteURLFailed
		expected string
	}{
		{
			name:     "no remote",
			err:      ErrGetRemoteURLFailed{Detail: "no remote 'origin' found"},
			expected: "failed to get remote URL: no remote 'origin' found",
		},
		{
			name:     "invalid config",
			err:      ErrGetRemoteURLFailed{Detail: "invalid git config"},
			expected: "failed to get remote URL: invalid git config",
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

// Test that errors implement the error interface
func TestErrorsImplementErrorInterface(_ *testing.T) {
	var _ error = ErrNotInRepo{}
	var _ error = ErrInvalidRemoteURL{}
	var _ error = ErrBranchNotFound{}
	var _ error = ErrCheckoutFailed{}
	var _ error = ErrPullFailed{}
	var _ error = ErrPushFailed{}
	var _ error = ErrDeleteBranchFailed{}
	var _ error = ErrGetCurrentBranchFailed{}
	var _ error = ErrGetRemoteURLFailed{}
}

// Benchmark error message generation
func BenchmarkErrPullFailed_Error(b *testing.B) {
	err := ErrPullFailed{
		Remote: "origin",
		Branch: "main",
		Detail: "connection refused",
	}

	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkErrPushFailed_Error(b *testing.B) {
	err := ErrPushFailed{
		Remote: "upstream",
		Branch: "feature/benchmark",
		Detail: "authentication failed",
	}

	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
