package migrate

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/user/git-mfpr/internal/git"
	"github.com/user/git-mfpr/internal/github"
)

// PRInfo represents pull request information
type PRInfo = github.PRInfo

// Options represents migration options
type Options struct {
	DryRun     bool
	BranchName string // optional custom name
	NoPush     bool
	NoCreate   bool
}

// Event represents a progress event
type Event struct {
	Type    string // "info", "success", "error", "command"
	Message string
	Detail  string
}

// EventHandler is a function that handles events
type EventHandler func(Event)

// Migrator provides PR migration functionality
type Migrator interface {
	// Main entry point
	MigratePR(prRef string, opts Options) error

	// For batch operations
	MigratePRs(prRefs []string, opts Options) error

	// Helper methods
	GetPRInfo(prRef string) (*PRInfo, error)
	GenerateBranchName(pr *PRInfo) string

	// Set event handler
	SetEventHandler(handler EventHandler)
}

// Client implements the Migrator interface
type Client struct {
	git     git.Git
	github  github.GitHub
	handler EventHandler
}

// New creates a new Migrator
func New() Migrator {
	return &Client{
		git:     git.New(),
		github:  github.New(),
		handler: func(Event) {}, // no-op by default
	}
}

// SetEventHandler sets the event handler
func (c *Client) SetEventHandler(handler EventHandler) {
	c.handler = handler
}

// emit sends an event to the handler
func (c *Client) emit(eventType, message, detail string) {
	c.handler(Event{
		Type:    eventType,
		Message: message,
		Detail:  detail,
	})
}

// parsePRRef parses various PR reference formats
func (c *Client) parsePRRef(prRef string) (owner, repo string, number int, err error) {
	// Try to parse as number first
	if num, err := strconv.Atoi(prRef); err == nil {
		// It's a simple number, get current repo
		owner, repo, err = c.git.CurrentRepo()
		if err != nil {
			return "", "", 0, fmt.Errorf("not in a git repository or no origin remote: %w", err)
		}
		return owner, repo, num, nil
	}

	// Try to parse as URL
	if strings.HasPrefix(prRef, "http://") || strings.HasPrefix(prRef, "https://") {
		u, err := url.Parse(prRef)
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid URL: %w", err)
		}

		// Expected format: https://github.com/owner/repo/pull/123
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) != 4 || parts[2] != "pull" {
			return "", "", 0, fmt.Errorf("invalid GitHub PR URL format")
		}

		num, err := strconv.Atoi(parts[3])
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid PR number in URL: %w", err)
		}

		return parts[0], parts[1], num, nil
	}

	// Try to parse as owner/repo#number
	if strings.Contains(prRef, "#") {
		parts := strings.Split(prRef, "#")
		if len(parts) != 2 {
			return "", "", 0, fmt.Errorf("invalid format, expected owner/repo#number")
		}

		num, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid PR number: %w", err)
		}

		repoParts := strings.Split(parts[0], "/")
		if len(repoParts) != 2 {
			return "", "", 0, fmt.Errorf("invalid repo format, expected owner/repo")
		}

		return repoParts[0], repoParts[1], num, nil
	}

	return "", "", 0, fmt.Errorf("unsupported PR reference format: %s", prRef)
}

// GetPRInfo fetches PR information
func (c *Client) GetPRInfo(prRef string) (*PRInfo, error) {
	owner, repo, number, err := c.parsePRRef(prRef)
	if err != nil {
		return nil, err
	}

	return c.github.GetPR(owner, repo, number)
}

// GenerateBranchName generates a branch name for the PR
func (c *Client) GenerateBranchName(pr *PRInfo) string {
	// Format: pr-{number}-{author}-{title-slug}
	titleSlug := slugify(pr.Title)
	branchName := fmt.Sprintf("pr-%d-%s-%s", pr.Number, pr.Author, titleSlug)

	// Ensure branch name is not too long
	if len(branchName) > 80 {
		// Truncate title slug but keep pr number and author
		baseLen := len(fmt.Sprintf("pr-%d-%s-", pr.Number, pr.Author))
		maxSlugLen := 80 - baseLen
		if maxSlugLen > 0 {
			titleSlug = titleSlug[:min(len(titleSlug), maxSlugLen)]
			branchName = fmt.Sprintf("pr-%d-%s-%s", pr.Number, pr.Author, titleSlug)
		}
	}

	return branchName
}

// slugify converts a string to a slug format
func slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace non-alphanumeric characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")

	// Remove leading/trailing hyphens
	s = strings.Trim(s, "-")

	// Limit to 40 characters
	if len(s) > 40 {
		s = s[:40]
		// Remove trailing hyphen if we cut in the middle of a word
		s = strings.TrimRight(s, "-")
	}

	return s
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MigratePR migrates a single PR
func (c *Client) MigratePR(prRef string, opts Options) error {
	// Parse PR reference
	owner, repo, number, err := c.parsePRRef(prRef)
	if err != nil {
		return err
	}

	c.emit("info", fmt.Sprintf("Migrating PR #%d from %s/%s", number, owner, repo), "")

	// Fetch PR information
	c.emit("info", "Fetching PR information...", "")
	pr, err := c.github.GetPR(owner, repo, number)
	if err != nil {
		return err
	}

	// Check if PR is from a fork
	if !pr.IsFork {
		c.emit("error", "PR is not from a fork", "")
		return fmt.Errorf("PR #%d is not from a fork", pr.Number)
	}

	// Check PR state
	if pr.State != "open" {
		c.emit("error", fmt.Sprintf("PR is %s", pr.State), "")
		return fmt.Errorf("PR #%d is %s", pr.Number, pr.State)
	}

	c.emit("info", fmt.Sprintf("Title: %s", pr.Title), "")
	c.emit("info", fmt.Sprintf("Author: %s", pr.Author), "")

	// Generate or use custom branch name
	branchName := opts.BranchName
	if branchName == "" {
		branchName = c.GenerateBranchName(pr)
	}
	c.emit("info", fmt.Sprintf("Branch: %s", branchName), "")

	// Check if branch already exists
	if c.git.HasBranch(branchName) {
		return fmt.Errorf("branch %s already exists. Use --branch-name to specify a different name or delete the existing branch", branchName)
	}

	if opts.DryRun {
		c.emit("command", "Would execute:", "git checkout "+pr.BaseBranch)
		c.emit("command", "Would execute:", "git pull origin "+pr.BaseBranch)
		c.emit("command", "Would execute:", fmt.Sprintf("gh pr checkout %d -b %s", pr.Number, branchName))
		if !opts.NoPush {
			c.emit("command", "Would execute:", "git push -u origin "+branchName)
		}
		if !opts.NoCreate {
			c.emit("info", "Would suggest creating PR with:", "")
			c.emit("command", "", fmt.Sprintf(`gh pr create --title "%s" --body "Migrated from #%d\nOriginal author: @%s" --base %s`,
				pr.Title, pr.Number, pr.Author, pr.BaseBranch))
		}
		return nil
	}

	// Switch to base branch
	c.emit("info", fmt.Sprintf("Switching to %s branch...", pr.BaseBranch), "")
	if err := c.git.Checkout(pr.BaseBranch); err != nil {
		return err
	}

	// Pull latest changes
	c.emit("info", "Pulling latest changes...", "")
	if err := c.git.Pull("origin", pr.BaseBranch); err != nil {
		return err
	}

	// Checkout PR
	c.emit("info", fmt.Sprintf("Checking out PR #%d...", pr.Number), "")
	if err := c.github.CheckoutPR(pr.Number, branchName); err != nil {
		return err
	}

	// Push to origin
	if !opts.NoPush {
		c.emit("info", "Pushing to origin...", "")
		if err := c.git.Push("origin", branchName); err != nil {
			return err
		}
		c.emit("success", "Pushed to origin", "")
	}

	// Success!
	c.emit("success", fmt.Sprintf("Successfully migrated PR #%d", pr.Number), "")

	// Suggest PR creation
	if !opts.NoCreate && !opts.NoPush {
		c.emit("info", "", "")
		c.emit("info", "Create PR with:", "")
		c.emit("command", "", fmt.Sprintf(`gh pr create --title "%s" --body "Migrated from #%d\nOriginal author: @%s" --base %s`,
			pr.Title, pr.Number, pr.Author, pr.BaseBranch))
	}

	return nil
}

// MigratePRs migrates multiple PRs
func (c *Client) MigratePRs(prRefs []string, opts Options) error {
	var errors []string
	successCount := 0

	for _, prRef := range prRefs {
		c.emit("info", "", "") // Empty line between PRs
		err := c.MigratePR(prRef, opts)
		if err != nil {
			errors = append(errors, fmt.Sprintf("PR %s: %v", prRef, err))
			c.emit("error", fmt.Sprintf("Failed to migrate %s", prRef), err.Error())
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		c.emit("info", "", "")
		c.emit("info", fmt.Sprintf("Migrated %d/%d PRs successfully", successCount, len(prRefs)), "")
		return fmt.Errorf("failed to migrate some PRs:\n%s", strings.Join(errors, "\n"))
	}

	c.emit("info", "", "")
	c.emit("success", fmt.Sprintf("Successfully migrated all %d PRs", len(prRefs)), "")
	return nil
}
