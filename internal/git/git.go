package git

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

type (
	BranchResult struct {
		Branch string
		Error  error
	}

	RepoResult struct {
		Owner string
		Repo  string
		Error error
	}

	OperationResult struct {
		Success bool
		Error   error
	}
)

func (r *BranchResult) IsSuccess() bool {
	return r.Error == nil
}

func (r *BranchResult) IsError() bool {
	return r.Error != nil
}

func (r *RepoResult) IsSuccess() bool {
	return r.Error == nil
}

func (r *RepoResult) IsError() bool {
	return r.Error != nil
}

func (r *OperationResult) IsSuccess() bool {
	return r.Success
}

func (r *OperationResult) IsError() bool {
	return !r.Success
}

type Git interface {
	CurrentBranch(ctx context.Context) (string, error)
	CurrentRepo(ctx context.Context) (owner, name string, err error)
	Checkout(ctx context.Context, branch string) error
	Pull(ctx context.Context, remote, branch string) error
	Push(ctx context.Context, remote, branch string) error
	HasBranch(ctx context.Context, name string) bool
	DeleteBranch(ctx context.Context, name string) error
	IsInRepo(ctx context.Context) bool

	CurrentBranchResult(ctx context.Context) *BranchResult
	CurrentRepoResult(ctx context.Context) *RepoResult
	CheckoutResult(ctx context.Context, branch string) *OperationResult
	PullResult(ctx context.Context, remote, branch string) *OperationResult
	PushResult(ctx context.Context, remote, branch string) *OperationResult
	DeleteBranchResult(ctx context.Context, name string) *OperationResult
}

type Client struct {
	timeout time.Duration
}

type Option func(*Client)

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

func New() Git {
	return NewWithOptions()
}

func NewWithOptions(opts ...Option) Git {
	client := &Client{
		timeout: 30 * time.Second,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) CurrentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", &ErrGetCurrentBranchFailed{Detail: err.Error()}
	}
	return strings.TrimSpace(string(output)), nil
}

func (c *Client) CurrentRepo(ctx context.Context) (owner, name string, err error) {
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", "", &ErrGetRemoteURLFailed{Detail: err.Error()}
	}

	remoteURL := strings.TrimSpace(string(output))

	if strings.HasPrefix(remoteURL, "git@github.com:") {
		parts := strings.TrimPrefix(remoteURL, "git@github.com:")
		parts = strings.TrimSuffix(parts, ".git")
		split := strings.Split(parts, "/")
		if len(split) != 2 {
			return "", "", &ErrInvalidRemoteURL{URL: remoteURL}
		}
		return split[0], split[1], nil
	} else if strings.Contains(remoteURL, "github.com") {
		parts := strings.Split(remoteURL, "github.com/")
		if len(parts) != 2 {
			return "", "", &ErrInvalidRemoteURL{URL: remoteURL}
		}
		repoPath := strings.TrimSuffix(parts[1], ".git")
		split := strings.Split(repoPath, "/")
		if len(split) != 2 {
			return "", "", &ErrInvalidRemoteURL{URL: remoteURL}
		}
		return split[0], split[1], nil
	}

	return "", "", &ErrInvalidRemoteURL{URL: remoteURL}
}

func (c *Client) Checkout(ctx context.Context, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", branch)
	if err := cmd.Run(); err != nil {
		return &ErrCheckoutFailed{Branch: branch, Detail: err.Error()}
	}
	return nil
}

func (c *Client) Pull(ctx context.Context, remote, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "pull", remote, branch)
	if err := cmd.Run(); err != nil {
		return &ErrPullFailed{Remote: remote, Branch: branch, Detail: err.Error()}
	}
	return nil
}

func (c *Client) Push(ctx context.Context, remote, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "push", "-u", remote, branch)
	if err := cmd.Run(); err != nil {
		return &ErrPushFailed{Remote: remote, Branch: branch, Detail: err.Error()}
	}
	return nil
}

func (c *Client) HasBranch(ctx context.Context, name string) bool {
	cmd := exec.CommandContext(ctx, "git", "show-ref", "--verify", "--quiet", "refs/heads/"+name)
	return cmd.Run() == nil
}

func (c *Client) DeleteBranch(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "git", "branch", "-D", name)
	if err := cmd.Run(); err != nil {
		return &ErrDeleteBranchFailed{Branch: name, Detail: err.Error()}
	}
	return nil
}

func (c *Client) IsInRepo(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func (c *Client) CurrentBranchResult(ctx context.Context) *BranchResult {
	result := &BranchResult{}

	branch, err := c.CurrentBranch(ctx)
	if err != nil {
		result.Error = err
	} else {
		result.Branch = branch
	}

	return result
}

func (c *Client) CurrentRepoResult(ctx context.Context) *RepoResult {
	result := &RepoResult{}

	owner, name, err := c.CurrentRepo(ctx)
	if err != nil {
		result.Error = err
	} else {
		result.Owner = owner
		result.Repo = name
	}

	return result
}

func (c *Client) CheckoutResult(ctx context.Context, branch string) *OperationResult {
	result := &OperationResult{}

	err := c.Checkout(ctx, branch)
	if err != nil {
		result.Error = err
	} else {
		result.Success = true
	}

	return result
}

func (c *Client) PullResult(ctx context.Context, remote, branch string) *OperationResult {
	result := &OperationResult{}

	err := c.Pull(ctx, remote, branch)
	if err != nil {
		result.Error = err
	} else {
		result.Success = true
	}

	return result
}

func (c *Client) PushResult(ctx context.Context, remote, branch string) *OperationResult {
	result := &OperationResult{}

	err := c.Push(ctx, remote, branch)
	if err != nil {
		result.Error = err
	} else {
		result.Success = true
	}

	return result
}

func (c *Client) DeleteBranchResult(ctx context.Context, name string) *OperationResult {
	result := &OperationResult{}

	err := c.DeleteBranch(ctx, name)
	if err != nil {
		result.Error = err
	} else {
		result.Success = true
	}

	return result
}
