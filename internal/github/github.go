package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type PRInfo struct {
	Number     int
	Title      string
	Author     string
	HeadBranch string
	BaseBranch string
	State      string
	URL        string
	HeadRefOID string
	IsFork     bool
}

type GitHub interface {
	GetPR(ctx context.Context, owner, repo string, number int) (*PRInfo, error)
	CheckoutPR(ctx context.Context, number int, branch string) error
	CreatePR(ctx context.Context, title, body, base string) error
	IsGHInstalled(ctx context.Context) error
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

func New() GitHub {
	return NewWithOptions()
}

func NewWithOptions(opts ...Option) GitHub {
	client := &Client{
		timeout: 30 * time.Second,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

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

type Result struct {
	Data  interface{}
	Error error
}

type PRResult struct {
	PR    *PRInfo
	Error error
}

func (r *PRResult) IsSuccess() bool {
	return r.Error == nil
}

func (r *PRResult) IsError() bool {
	return r.Error != nil
}

func (c *Client) IsGHInstalled(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "gh", "--version")
	if err := cmd.Run(); err != nil {
		return &ErrGHNotInstalled{}
	}
	return nil
}

func (c *Client) GetPR(ctx context.Context, owner, repo string, number int) (*PRInfo, error) {
	if err := c.IsGHInstalled(ctx); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "gh", "pr", "view", strconv.Itoa(number),
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--json", "number,title,author,headRefName,baseRefName,state,headRefOid,isFork,url")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "no pull requests found") {
				return nil, &ErrPRNotFound{Number: number, Owner: owner, Repo: repo}
			}
			return nil, &ErrPRFetchFailed{Number: number, Owner: owner, Repo: repo, Detail: stderr}
		}
		return nil, &ErrPRFetchFailed{Number: number, Owner: owner, Repo: repo, Detail: err.Error()}
	}

	var pr ghPRResponse
	if err := json.Unmarshal(output, &pr); err != nil {
		return nil, &ErrPRParseFailed{Detail: err.Error()}
	}

	return &PRInfo{
		Number:     pr.Number,
		Title:      pr.Title,
		Author:     pr.Author.Login,
		HeadBranch: pr.HeadRefName,
		BaseBranch: pr.BaseRefName,
		State:      pr.State,
		URL:        pr.URL,
		HeadRefOID: pr.HeadRefOID,
		IsFork:     pr.IsFork,
	}, nil
}

func (c *Client) CheckoutPR(ctx context.Context, number int, branch string) error {
	cmd := exec.CommandContext(ctx, "gh", "pr", "checkout", strconv.Itoa(number), "-b", branch)
	if err := cmd.Run(); err != nil {
		return &ErrPRCheckoutFailed{Number: number, Detail: err.Error()}
	}
	return nil
}

func (c *Client) CreatePR(ctx context.Context, title, body, base string) error {
	cmd := exec.CommandContext(ctx, "gh", "pr", "create",
		"--title", title,
		"--body", body,
		"--base", base)

	if err := cmd.Run(); err != nil {
		return &ErrPRCreateFailed{Detail: err.Error()}
	}
	return nil
}
