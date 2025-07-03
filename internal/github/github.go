package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// PRInfo contains pull request information
type PRInfo struct {
	Number      int
	Title       string
	Author      string
	HeadBranch  string
	BaseBranch  string
	State       string
	URL         string
	HeadRefOID  string
	IsFork      bool
}

// GitHub provides an interface for GitHub operations
type GitHub interface {
	GetPR(owner, repo string, number int) (*PRInfo, error)
	CheckoutPR(number int, branch string) error
	CreatePR(title, body, base string) error
	IsGHInstalled() error
}

// Client implements the GitHub interface
type Client struct{}

// New creates a new GitHub client
func New() GitHub {
	return &Client{}
}

// ghPRResponse represents the JSON response from gh pr view
type ghPRResponse struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	State       string `json:"state"`
	HeadRefName string `json:"headRefName"`
	BaseRefName string `json:"baseRefName"`
	HeadRefOID  string `json:"headRefOid"`
	IsFork      bool   `json:"isFork"`
	URL         string `json:"url"`
	Author      struct {
		Login string `json:"login"`
	} `json:"author"`
}

// IsGHInstalled checks if gh CLI is installed
func (c *Client) IsGHInstalled() error {
	cmd := exec.Command("gh", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh CLI is not installed. Install it from https://cli.github.com")
	}
	return nil
}

// GetPR fetches pull request information
func (c *Client) GetPR(owner, repo string, number int) (*PRInfo, error) {
	// Check if gh is installed
	if err := c.IsGHInstalled(); err != nil {
		return nil, err
	}

	// Build the command
	cmd := exec.Command("gh", "pr", "view", strconv.Itoa(number),
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--json", "number,title,author,headRefName,baseRefName,state,headRefOid,isFork,url")
	
	output, err := cmd.Output()
	if err != nil {
		// Check if it's an exit error to provide better error message
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "no pull requests found") {
				return nil, fmt.Errorf("PR #%d not found in %s/%s", number, owner, repo)
			}
			return nil, fmt.Errorf("failed to fetch PR: %s", stderr)
		}
		return nil, fmt.Errorf("failed to fetch PR: %w", err)
	}

	// Parse the JSON response
	var pr ghPRResponse
	if err := json.Unmarshal(output, &pr); err != nil {
		return nil, fmt.Errorf("failed to parse PR data: %w", err)
	}

	// Convert to our PRInfo structure
	return &PRInfo{
		Number:      pr.Number,
		Title:       pr.Title,
		Author:      pr.Author.Login,
		HeadBranch:  pr.HeadRefName,
		BaseBranch:  pr.BaseRefName,
		State:       pr.State,
		URL:         pr.URL,
		HeadRefOID:  pr.HeadRefOID,
		IsFork:      pr.IsFork,
	}, nil
}

// CheckoutPR checks out a pull request to a local branch
func (c *Client) CheckoutPR(number int, branch string) error {
	cmd := exec.Command("gh", "pr", "checkout", strconv.Itoa(number), "-b", branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout PR #%d: %w", number, err)
	}
	return nil
}

// CreatePR creates a new pull request
func (c *Client) CreatePR(title, body, base string) error {
	cmd := exec.Command("gh", "pr", "create",
		"--title", title,
		"--body", body,
		"--base", base)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}
	return nil
}