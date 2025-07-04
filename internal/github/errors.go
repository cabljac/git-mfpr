package github

import "fmt"

type (
	ErrGHNotInstalled struct{}

	ErrPRNotFound struct {
		Number int
		Owner  string
		Repo   string
	}

	ErrPRFetchFailed struct {
		Number int
		Owner  string
		Repo   string
		Detail string
	}

	ErrPRParseFailed struct {
		Detail string
	}

	ErrPRCheckoutFailed struct {
		Number int
		Detail string
	}

	ErrPRCreateFailed struct {
		Detail string
	}
)

func (e ErrGHNotInstalled) Error() string {
	return "gh CLI is not installed. Install it from https://cli.github.com"
}

func (e ErrPRNotFound) Error() string {
	return fmt.Sprintf("PR #%d not found in %s/%s", e.Number, e.Owner, e.Repo)
}

func (e ErrPRFetchFailed) Error() string {
	return fmt.Sprintf("failed to fetch PR #%d from %s/%s: %s", e.Number, e.Owner, e.Repo, e.Detail)
}

func (e ErrPRParseFailed) Error() string {
	return fmt.Sprintf("failed to parse PR data: %s", e.Detail)
}

func (e ErrPRCheckoutFailed) Error() string {
	return fmt.Sprintf("failed to checkout PR #%d: %s", e.Number, e.Detail)
}

func (e ErrPRCreateFailed) Error() string {
	return fmt.Sprintf("failed to create PR: %s", e.Detail)
}
