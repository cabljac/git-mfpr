package migrate

import (
	"testing"
)

func TestErrPRNotFound_Error(t *testing.T) {
	err := &ErrPRNotFound{
		Number: 123,
		Owner:  "testowner",
		Repo:   "testrepo",
	}

	expected := "PR #123 not found in testowner/testrepo"
	if err.Error() != expected {
		t.Errorf("ErrPRNotFound.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestErrPRNotFound_EmptyFields(t *testing.T) {
	err := &ErrPRNotFound{
		Number: 0,
		Owner:  "",
		Repo:   "",
	}

	expected := "PR #0 not found in /"
	if err.Error() != expected {
		t.Errorf("ErrPRNotFound.Error() = %q, want %q", err.Error(), expected)
	}
}
