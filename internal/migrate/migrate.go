package migrate

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/user/git-mfpr/internal/git"
	"github.com/user/git-mfpr/internal/github"
)

const (
	EventInfo    EventType = "info"
	EventSuccess EventType = "success"
	EventError   EventType = "error"
	EventCommand EventType = "command"
)

type EventType string

type PRInfo = github.PRInfo

type Options struct {
	DryRun     bool
	BranchName string
	NoPush     bool
	NoCreate   bool
}

type Event struct {
	Type    EventType
	Message string
	Detail  string
}

type EventHandler func(Event)

type Migrator interface {
	MigratePR(ctx context.Context, prRef string, opts Options) error

	MigratePRs(ctx context.Context, prRefs []string, opts Options) error

	GetPRInfo(ctx context.Context, prRef string) (*PRInfo, error)
	GenerateBranchName(pr *PRInfo) string

	SetEventHandler(handler EventHandler)
}

type Client struct {
	git     git.Git
	github  github.GitHub
	handler EventHandler
}

func New() Migrator {
	return &Client{
		git:     git.New(),
		github:  github.New(),
		handler: func(Event) {},
	}
}

func (c *Client) SetEventHandler(handler EventHandler) {
	c.handler = handler
}

func (c *Client) emit(eventType EventType, message, detail string) {
	c.handler(Event{
		Type:    eventType,
		Message: message,
		Detail:  detail,
	})
}

func (c *Client) parsePRRef(prRef string) (owner, repo string, number int, err error) {
	if num, err := strconv.Atoi(prRef); err == nil {
		ctx := context.Background()
		owner, repo, err = c.git.CurrentRepo(ctx)
		if err != nil {
			return "", "", 0, fmt.Errorf("not in a git repository or no origin remote: %w", err)
		}
		return owner, repo, num, nil
	}

	if strings.HasPrefix(prRef, "http://") || strings.HasPrefix(prRef, "https://") {
		u, err := url.Parse(prRef)
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid URL: %w", err)
		}

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

	return "", "", 0, &ErrInvalidPRRef{Ref: prRef}
}

func (c *Client) GetPRInfo(ctx context.Context, prRef string) (*PRInfo, error) {
	owner, repo, number, err := c.parsePRRef(prRef)
	if err != nil {
		return nil, err
	}

	return c.github.GetPR(ctx, owner, repo, number)
}

func (c *Client) GenerateBranchName(pr *PRInfo) string {
	// Simple, clear branch name: migrated-<PR-number>
	return fmt.Sprintf("migrated-%d", pr.Number)
}

func slugify(s string) string {
	s = strings.ToLower(s)

	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")

	s = strings.Trim(s, "-")

	if len(s) > 40 {
		s = s[:40]
		s = strings.TrimRight(s, "-")
	}

	return s
}

func (c *Client) validatePRState(pr *PRInfo) error {
	if !pr.IsFork {
		c.emit(EventError, "PR is not from a fork (it's from the same repository)", "")
		return &ErrPRNotFork{Number: pr.Number}
	}
	// GitHub returns state in uppercase, so we need to compare case-insensitively
	if !strings.EqualFold(pr.State, "open") {
		c.emit(EventError, fmt.Sprintf("PR is %s (only open PRs can be migrated)", pr.State), "")
		return &ErrPRClosed{Number: pr.Number, State: pr.State}
	}
	return nil
}

func (c *Client) emitPRInfo(pr *PRInfo) {
	c.emit(EventInfo, fmt.Sprintf("Title: %s", pr.Title), "")
	c.emit(EventInfo, fmt.Sprintf("Author: %s", pr.Author), "")
	c.emit(EventInfo, fmt.Sprintf("Base branch: %s", pr.BaseBranch), "")
	if pr.IsFork {
		c.emit(EventInfo, "PR is from a fork", "")
	} else {
		c.emit(EventInfo, "PR is from the same repository", "")
	}
}

func (c *Client) handleDryRun(pr *PRInfo, branchName string, opts Options) {
	c.emit(EventCommand, "Would execute:", "git checkout "+pr.BaseBranch)
	c.emit(EventCommand, "Would execute:", "git pull origin "+pr.BaseBranch)
	c.emit(EventCommand, "Would execute:", fmt.Sprintf("gh pr checkout %d -b %s", pr.Number, branchName))
	if !opts.NoPush {
		c.emit(EventCommand, "Would execute:", "git push -u origin "+branchName)
	}
	if !opts.NoCreate {
		c.emit(EventInfo, "Would suggest creating PR with:", "")
		c.emit(EventCommand, "", fmt.Sprintf(`gh pr create --title "%s" --body "Migrated from #%d\nOriginal author: @%s" --base %s`,
			pr.Title, pr.Number, pr.Author, pr.BaseBranch))
	}
}

func (c *Client) checkoutAndPullBase(ctx context.Context, pr *PRInfo) error {
	c.emit(EventInfo, fmt.Sprintf("Switching to %s branch...", pr.BaseBranch), "")
	if err := c.git.Checkout(ctx, pr.BaseBranch); err != nil {
		return err
	}
	c.emit(EventInfo, "Pulling latest changes...", "")
	if err := c.git.Pull(ctx, "origin", pr.BaseBranch); err != nil {
		return err
	}
	return nil
}

func (c *Client) pushAndEmit(ctx context.Context, branchName string) error {
	c.emit(EventInfo, "Pushing to origin...", "")
	if err := c.git.Push(ctx, "origin", branchName); err != nil {
		return err
	}
	c.emit(EventSuccess, "Pushed to origin", "")
	return nil
}

func (c *Client) emitCreatePR(pr *PRInfo) {
	c.emit(EventInfo, "", "")
	c.emit(EventInfo, "Create PR with:", "")
	c.emit(EventCommand, "", fmt.Sprintf(`gh pr create --title "%s" --body "Migrated from #%d\nOriginal author: @%s" --base %s`,
		pr.Title, pr.Number, pr.Author, pr.BaseBranch))
}

func (c *Client) MigratePR(ctx context.Context, prRef string, opts Options) error {
	owner, repo, number, err := c.parsePRRef(prRef)
	if err != nil {
		return err
	}

	c.emit(EventInfo, fmt.Sprintf("Migrating PR #%d from %s/%s", number, owner, repo), "")
	c.emit(EventInfo, "Fetching PR information...", "")
	pr, err := c.github.GetPR(ctx, owner, repo, number)
	if err != nil {
		return err
	}

	if err := c.validatePRState(pr); err != nil {
		return err
	}

	c.emitPRInfo(pr)

	branchName := opts.BranchName
	if branchName == "" {
		branchName = c.GenerateBranchName(pr)
	}
	c.emit(EventInfo, fmt.Sprintf("Branch: %s", branchName), "")

	if c.git.HasBranch(ctx, branchName) {
		return &ErrBranchExists{BranchName: branchName}
	}

	if opts.DryRun {
		c.handleDryRun(pr, branchName, opts)
		return nil
	}

	if err := c.checkoutAndPullBase(ctx, pr); err != nil {
		return err
	}

	c.emit(EventInfo, fmt.Sprintf("Checking out PR #%d...", pr.Number), "")
	if err := c.github.CheckoutPR(ctx, owner, repo, pr.Number, branchName); err != nil {
		return err
	}

	if !opts.NoPush {
		if err := c.pushAndEmit(ctx, branchName); err != nil {
			return err
		}
	}

	c.emit(EventSuccess, fmt.Sprintf("Successfully migrated PR #%d", pr.Number), "")

	if !opts.NoCreate && !opts.NoPush {
		c.emitCreatePR(pr)
	}

	return nil
}

func (c *Client) MigratePRs(ctx context.Context, prRefs []string, opts Options) error {
	var errors []string
	successCount := 0

	for _, prRef := range prRefs {
		c.emit(EventInfo, "", "")
		err := c.MigratePR(ctx, prRef, opts)
		if err != nil {
			errors = append(errors, fmt.Sprintf("PR %s: %v", prRef, err))
			c.emit(EventError, fmt.Sprintf("Failed to migrate %s", prRef), err.Error())
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		c.emit(EventInfo, "", "")
		c.emit(EventInfo, fmt.Sprintf("Migrated %d/%d PRs successfully", successCount, len(prRefs)), "")
		return fmt.Errorf("failed to migrate some PRs:\n%s", strings.Join(errors, "\n"))
	}

	c.emit(EventInfo, "", "")
	c.emit(EventSuccess, fmt.Sprintf("Successfully migrated all %d PRs", len(prRefs)), "")
	return nil
}
