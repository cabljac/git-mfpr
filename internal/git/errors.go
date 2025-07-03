package git

import "fmt"

type (
	ErrNotInRepo struct{}

	ErrInvalidRemoteURL struct {
		URL string
	}

	ErrBranchNotFound struct {
		Branch string
	}

	ErrCheckoutFailed struct {
		Branch string
		Detail string
	}

	ErrPullFailed struct {
		Remote string
		Branch string
		Detail string
	}

	ErrPushFailed struct {
		Remote string
		Branch string
		Detail string
	}

	ErrDeleteBranchFailed struct {
		Branch string
		Detail string
	}

	ErrGetCurrentBranchFailed struct {
		Detail string
	}

	ErrGetRemoteURLFailed struct {
		Detail string
	}
)

func (e ErrNotInRepo) Error() string {
	return "not in a git repository"
}

func (e ErrInvalidRemoteURL) Error() string {
	return fmt.Sprintf("invalid remote URL format: %s", e.URL)
}

func (e ErrBranchNotFound) Error() string {
	return fmt.Sprintf("branch %s not found", e.Branch)
}

func (e ErrCheckoutFailed) Error() string {
	return fmt.Sprintf("failed to checkout branch %s: %s", e.Branch, e.Detail)
}

func (e ErrPullFailed) Error() string {
	return fmt.Sprintf("failed to pull %s/%s: %s", e.Remote, e.Branch, e.Detail)
}

func (e ErrPushFailed) Error() string {
	return fmt.Sprintf("failed to push to %s/%s: %s", e.Remote, e.Branch, e.Detail)
}

func (e ErrDeleteBranchFailed) Error() string {
	return fmt.Sprintf("failed to delete branch %s: %s", e.Branch, e.Detail)
}

func (e ErrGetCurrentBranchFailed) Error() string {
	return fmt.Sprintf("failed to get current branch: %s", e.Detail)
}

func (e ErrGetRemoteURLFailed) Error() string {
	return fmt.Sprintf("failed to get remote URL: %s", e.Detail)
}
