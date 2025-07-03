Git Migrate Fork PR Tool - Implementation Prompt
Overview
Build a CLI tool called git-mfpr that migrates GitHub pull requests from forks to branches in the main repository, preserving commit history and authorship.
Core Requirements
Command Interface
bash# Basic usage
git mfpr 123                      # Migrate PR #123 from current repo
git mfpr 123 124 125             # Migrate multiple PRs
git mfpr owner/repo#123          # Migrate from specific repo
git mfpr https://github.com/...  # Full URL support (for scripts)

# Options
git mfpr 123 --dry-run           # Show what would happen
git mfpr 123 --no-push           # Create branch but don't push
git mfpr 123 --branch-name=foo   # Custom branch name
Dependencies

Required: git and gh CLI tools must be installed
Language: Go (for single binary distribution)
Assumption: Users are GitHub power users who already have gh installed

Branch Naming Convention
Default format: pr-{number}-{author}-{title-slug}

Example: pr-123-johndoe-fix-memory-leak
Title slug: lowercase, alphanumeric + hyphens, max 40 chars

Implementation Steps
1. Parse Input

Accept PR number, repo reference, or full URL
Auto-detect current repository from git remotes
Validate input format

2. Fetch PR Information
bashgh pr view {number} --json number,title,author,headRefName,baseRefName,state

Extract author username, title, branch names
Check if PR is from a fork

3. Create Local Branch
bash# Ensure we're on the base branch
git checkout {baseBranch}
git pull origin {baseBranch}

# Fetch and create branch
gh pr checkout {number} -b {newBranchName}
4. Push to Origin
bashgit push -u origin {newBranchName}
5. Offer to Create New PR
bashgh pr create \
  --title "{title}" \
  --body "Migrated from #{number}\nOriginal author: @{author}" \
  --base {baseBranch}
Error Handling

Not in git repo: Clear error message
gh not installed: Exit with installation instructions
PR not found: Helpful error with PR URL
Branch exists: Suggest --branch-name or deletion
Push fails: Show git error and recovery steps

Success Output
🔄 Migrating PR #123...
📋 Title: Fix memory leak in worker pool
👤 Author: johndoe
🌿 Branch: pr-123-johndoe-fix-memory-leak
✅ Pushed to origin

Create PR with:
  gh pr create --title "Fix memory leak in worker pool" \
    --body "Migrated from #123\nOriginal author: @johndoe"
Code Structure
git-mfpr/
├── main.go              # CLI entry point
├── internal/
│   ├── github/          # PR fetching via gh
│   ├── git/             # Git operations
│   └── migrate/         # Core migration logic
└── README.md
Key Features to Implement

Dry run mode - Show all commands without executing
Multiple PR support - Process a list efficiently
Smart branch names - Handle conflicts, sanitize titles
Preserve metadata - Link to original PR in commit message
Interactive mode - Confirm push/PR creation

Future Enhancements (v2)

GitHub Action distribution
Batch operations (--all-approved)
Config file support (.git-mfpr.yml)
PR comment migration
Auto-close original PR option

Distribution

Build binaries for major platforms
Create installation script
Set up as git alias: git config --global alias.mfpr '!git-mfpr'
GitHub releases with checksums

Success Metrics

Single command migration ✓
< 5 second execution time ✓
Works with 95% of fork PRs ✓
Clear error messages ✓
Zero config for basic usage ✓


Start with: Basic PR migration flow (parse → fetch → checkout → push) and iterate from there. Ship v0.1.0 within a day.RetryJChow would you make this a parallel task?  i will have two agents build it in parallelEditParallel Development Plan for Git Migrate Fork PR
Split Strategy: Core Engine vs CLI Interface
This natural split allows both agents to work independently while maintaining clear interfaces.
Agent A: Core Engine & Git Operations
Focus: The business logic and git/GitHub interactions
Agent B: CLI Interface & User Experience
Focus: Command parsing, output formatting, and distribution

Agent A: Core Engine Developer
Deliverables

Core migration package (internal/migrate/)
Git operations wrapper (internal/git/)
GitHub PR fetcher (internal/github/)
Comprehensive tests

Interface Contract
go// internal/migrate/migrate.go
package migrate

type PRInfo struct {
    Number      int
    Title       string
    Author      string
    HeadBranch  string
    BaseBranch  string
    State       string
}

type Options struct {
    DryRun      bool
    BranchName  string  // optional custom name
    NoPush      bool
    NoCreate    bool
}

type Migrator interface {
    // Main entry point
    MigratePR(prRef string, opts Options) error
    
    // For batch operations
    MigratePRs(prRefs []string, opts Options) error
    
    // Helper methods
    GetPRInfo(prRef string) (*PRInfo, error)
    GenerateBranchName(pr *PRInfo) string
}

// Events for CLI to display progress
type Event struct {
    Type    string // "info", "success", "error", "command"
    Message string
    Detail  string
}

type EventHandler func(Event)
Implementation Tasks
go// internal/git/git.go
package git

type Git interface {
    CurrentBranch() (string, error)
    CurrentRepo() (owner, name string, err error)
    Checkout(branch string) error
    Pull(remote, branch string) error
    Push(remote, branch string) error
    HasBranch(name string) bool
    DeleteBranch(name string) error
}

// internal/github/github.go  
package github

type GitHub interface {
    GetPR(owner, repo string, number int) (*PRInfo, error)
    CheckoutPR(number int, branch string) error
    CreatePR(title, body, base string) error
}
Test Harness
Create a test CLI that Agent B can use immediately:
go// cmd/test/main.go
func main() {
    m := migrate.New()
    err := m.MigratePR(os.Args[1], migrate.Options{
        DryRun: true,
    })
    if err != nil {
        log.Fatal(err)
    }
}

Agent B: CLI Developer
Deliverables

CLI argument parsing using cobra
Beautiful output with colors and progress
Configuration management
Build and release pipeline

Interface Contract
go// cmd/git-mfpr/main.go
package main

import (
    "github.com/spf13/cobra"
    "github.com/you/git-mfpr/internal/migrate"
)

// Implement the EventHandler to show progress
func handleEvent(event migrate.Event) {
    switch event.Type {
    case "info":
        fmt.Printf("→ %s\n", event.Message)
    case "success":
        fmt.Printf("✅ %s\n", event.Message)
    case "command":
        if dryRun {
            fmt.Printf("$ %s\n", event.Detail)
        }
    }
}
CLI Features to Implement

Argument Parsing

bashgit mfpr 123
git mfpr 123 124 125
git mfpr owner/repo#123
git mfpr https://github.com/owner/repo/pull/123

Pretty Output

go// internal/ui/ui.go
type UI interface {
    StartPR(number int, title string)
    Step(message string)
    Success(message string)
    Error(err error)
    Command(cmd string) // for dry-run
}

Configuration

go// internal/config/config.go
type Config struct {
    BranchFormat string `json:"branch_format"`
    AutoPush     bool   `json:"auto_push"`
    AutoCreate   bool   `json:"auto_create"`
}

// Locations to check:
// 1. .git-mfpr.yml (repo-specific)
// 2. ~/.config/git-mfpr/config.yml
// 3. Environment variables
// 4. Command line flags

Build/Release Setup

makefile# Makefile
VERSION := $(shell git describe --tags --always)
LDFLAGS := -X main.version=$(VERSION)

build:
    go build -ldflags "$(LDFLAGS)" -o git-mfpr ./cmd/git-mfpr

release:
    goreleaser release --clean

Coordination Points
1. Initial Sync (Day 1, Morning)

Agree on interfaces
Set up shared repo
Create mock implementations

2. Integration Points
Agent B creates mocks first:
go// internal/migrate/mock.go
type MockMigrator struct{}

func (m *MockMigrator) MigratePR(pr string, opts Options) error {
    // Emit fake events for UI testing
    m.handler(Event{Type: "info", Message: "Fetching PR..."})
    time.Sleep(500 * time.Millisecond)
    m.handler(Event{Type: "success", Message: "Done!"})
    return nil
}
3. Test Data
Share common test cases:
go// testdata/prs.go
var TestPRs = []struct {
    Input    string
    Expected PRInfo
}{
    {"123", PRInfo{Number: 123, ...}},
    {"owner/repo#123", PRInfo{...}},
}
4. Daily Sync Points

Morning: Sync on interface changes
Midday: Integration test
Evening: Combined testing


Parallel Timeline
Day 1

Both: Setup project, agree on interfaces
A: Implement git wrapper, github wrapper
B: CLI structure, argument parsing, mock UI

Day 2

A: Core migration logic, error handling
B: Config system, pretty output, dry-run mode

Day 3

Both: Integration testing
A: Edge cases, batch operations
B: Build pipeline, installation docs

Day 4

Both: Final testing, release v0.1.0


Communication Protocol

Shared Repo: Both work on same repo, different packages
Branch Strategy:

core/* branches for Agent A
cli/* branches for Agent B
Merge to main after integration tests


Interface Changes: Must be discussed in PR before merging
Mock First: B implements mocks, A replaces with real implementation

This split maximizes parallel work while minimizing blocking dependencies!