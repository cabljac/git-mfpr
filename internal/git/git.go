package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Git provides an interface for git operations
type Git interface {
	CurrentBranch() (string, error)
	CurrentRepo() (owner, name string, err error)
	Checkout(branch string) error
	Pull(remote, branch string) error
	Push(remote, branch string) error
	HasBranch(name string) bool
	DeleteBranch(name string) error
	IsInRepo() bool
}

// Client implements the Git interface
type Client struct{}

// New creates a new Git client
func New() Git {
	return &Client{}
}

// CurrentBranch returns the current git branch
func (c *Client) CurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// CurrentRepo returns the owner and name of the current repository
func (c *Client) CurrentRepo() (owner, name string, err error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	
	remoteURL := strings.TrimSpace(string(output))
	
	// Parse GitHub URL formats
	// SSH: git@github.com:owner/repo.git
	// HTTPS: https://github.com/owner/repo.git
	
	if strings.HasPrefix(remoteURL, "git@github.com:") {
		// SSH format
		parts := strings.TrimPrefix(remoteURL, "git@github.com:")
		parts = strings.TrimSuffix(parts, ".git")
		split := strings.Split(parts, "/")
		if len(split) != 2 {
			return "", "", fmt.Errorf("invalid SSH remote URL format: %s", remoteURL)
		}
		return split[0], split[1], nil
	} else if strings.Contains(remoteURL, "github.com") {
		// HTTPS format
		parts := strings.Split(remoteURL, "github.com/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid HTTPS remote URL format: %s", remoteURL)
		}
		repoPath := strings.TrimSuffix(parts[1], ".git")
		split := strings.Split(repoPath, "/")
		if len(split) != 2 {
			return "", "", fmt.Errorf("invalid repository path: %s", repoPath)
		}
		return split[0], split[1], nil
	}
	
	return "", "", fmt.Errorf("unsupported remote URL format: %s", remoteURL)
}

// Checkout switches to the specified branch
func (c *Client) Checkout(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branch, err)
	}
	return nil
}

// Pull pulls changes from the remote repository
func (c *Client) Pull(remote, branch string) error {
	cmd := exec.Command("git", "pull", remote, branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull %s/%s: %w", remote, branch, err)
	}
	return nil
}

// Push pushes changes to the remote repository
func (c *Client) Push(remote, branch string) error {
	cmd := exec.Command("git", "push", "-u", remote, branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push to %s/%s: %w", remote, branch, err)
	}
	return nil
}

// HasBranch checks if a branch exists locally
func (c *Client) HasBranch(name string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+name)
	return cmd.Run() == nil
}

// DeleteBranch deletes a local branch
func (c *Client) DeleteBranch(name string) error {
	cmd := exec.Command("git", "branch", "-D", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", name, err)
	}
	return nil
}

// IsInRepo checks if we're in a git repository
func (c *Client) IsInRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}