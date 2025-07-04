package migrate

import "fmt"

type (
	ErrPRNotFound struct {
		Number int
		Owner  string
		Repo   string
	}

	ErrPRNotFork struct {
		Number int
	}

	ErrPRClosed struct {
		Number int
		State  string
	}

	ErrBranchExists struct {
		BranchName string
	}

	ErrInvalidPRRef struct {
		Ref string
	}
)

func (e ErrPRNotFound) Error() string {
	return fmt.Sprintf("PR #%d not found in %s/%s", e.Number, e.Owner, e.Repo)
}

func (e ErrPRNotFork) Error() string {
	return fmt.Sprintf("PR #%d is not from a fork", e.Number)
}

func (e ErrPRClosed) Error() string {
	return fmt.Sprintf("PR #%d is %s", e.Number, e.State)
}

func (e ErrBranchExists) Error() string {
	return fmt.Sprintf("branch %s already exists. Use --branch-name to specify a different name or delete the existing branch", e.BranchName)
}

func (e ErrInvalidPRRef) Error() string {
	return fmt.Sprintf("unsupported PR reference format: %s", e.Ref)
}
