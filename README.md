# git-mfpr

[![CI](https://github.com/cabljac/git-mfpr/actions/workflows/ci.yml/badge.svg)](https://github.com/cabljac/git-mfpr/actions/workflows/ci.yml)
[![codecov](https://codecov.io/github/cabljac/git-mfpr/branch/main/graph/badge.svg?token=OAI9FZQ1BF)](https://codecov.io/github/cabljac/git-mfpr)
[![Release](https://img.shields.io/github/release/cabljac/git-mfpr.svg)](https://github.com/cabljac/git-mfpr/releases/latest)

A CLI tool that migrates GitHub pull requests from forks to branches in the main repository, preserving commit history and authorship.

## Overview

`git-mfpr` (Git Migrate Fork PR) helps maintainers bring pull requests from forked repositories into their main repository as regular branches. This is useful when you want to:

- Continue work on a PR after the original author has abandoned it
- Run CI/CD pipelines that don't work well with forks
- Collaborate more easily on a contribution
- Maintain a cleaner repository structure

## Value Proposition

### Why Use git-mfpr Instead of Manual Commands?

While you could manually run `gh pr checkout` and `git push`, `git-mfpr` provides significant value through:

#### **Safety & Validation**
- **Validates PR is from a fork** (prevents mistakes on same-repo PRs)
- **Checks if branch already exists** (no accidental overwrites)
- **Verifies PR is still open** (no dead PR migrations)
- **Simple branch naming** (consistent, predictable names)

#### **Workflow Automation**
- **One command instead of 3-4 manual steps**
- **Batch processing** (multiple PRs at once)
- **Dry-run mode** (preview actions before executing)
- **Consistent commit messages and PR descriptions**

#### **Team Benefits**
- **Standardized process** across your entire team
- **Reduced onboarding time** (new contributors learn one command)
- **Prevents human error** (forgot to push, wrong branch name, etc.)
- **Documentation** (the process is codified in the tool)

#### **Automation & CI/CD**
- **GitHub Action integration** for automated workflows
- **Scheduled cleanup** of abandoned PRs
- **Non-interactive operation** (works in automated environments)
- **CI/CD pipeline integration** (migrate PRs as part of your workflow)

### Real-World Scenarios

#### **Scenario 1: Abandoned PRs**
```bash
# Instead of manually checking each PR status and running commands
git-mfpr 123 124 125  # Batch migrate multiple abandoned PRs
```

#### **Scenario 2: Team Onboarding**
```bash
# New team member doesn't need to learn the manual process
git-mfpr 123 --dry-run  # Shows them exactly what will happen
```

#### **Scenario 3: Automated Cleanup**
```yaml
# GitHub Action to automatically migrate PRs that pass CI but are abandoned
- uses: yourusername/git-mfpr@v1
  with:
    pr-ref: ${{ needs.ci.outputs.pr-number }}
```

### When It's Worth It

**Perfect for:**
- Managing multiple repos with frequent fork PRs
- Team environments where consistency matters
- Automation needs (CI/CD, scheduled cleanup)
- High-volume PR management
- Reducing cognitive load and preventing mistakes

**May not be worth it for:**
- Occasional use (once a month)
- Single-person projects
- Simple workflows with just a few commands

**Bottom line:** It's a quality-of-life tool that shines in team environments and automation scenarios, not just time-saving for individual use.

## Features

- 🔄 Preserves complete commit history and authorship
- 🚀 Simple one-command migration
- 🎯 Simple, consistent branch naming with PR number
- 🔍 Dry-run mode to preview actions
- 📦 Batch migration support for multiple PRs
- 🛡️ Safety checks to prevent accidental overwrites

## Installation

### Prerequisites

- Git
- [GitHub CLI (`gh`)](https://cli.github.com/) - must be installed and authenticated

### Install from Source

```bash
go install github.com/cabljac/git-mfpr/cmd/git-mfpr@latest
```

### Download Binary

Download the latest release for your platform from the [releases page](https://github.com/cabljac/git-mfpr/releases).

## Usage

### Basic Usage

```bash
# Migrate PR #123 from the current repository
git-mfpr 123

# Migrate multiple PRs
git-mfpr 124 125

# Migrate from a specific repository
git-mfpr owner/repo#123

# Migrate using full GitHub URL
git-mfpr https://github.com/owner/repo/pull/123
```

### Options

```bash
--dry-run              # Preview what would happen without making changes
--no-push              # Create local branch but don't push to origin
--no-create            # Don't offer to create a new PR
--branch-name string   # Use custom branch name (single PR only)
```

### As a CLI

You can use `git-mfpr` directly in your terminal, or as a custom git command:

```sh
# Migrate PR #123 from the current repo (dry run)
git-mfpr 123 --dry-run

# Migrate a PR from a specific repo
git-mfpr owner/repo#456

# Use as a custom git command (if installed as git-mfpr)
git mfpr 789

# With custom branch name
git-mfpr 123 --branch-name=my-feature-branch
```

---

### As a GitHub Action

Add this step to your workflow (e.g., `.github/workflows/migrate.yml`):

```yaml
name: Migrate Fork PRs

on:
  workflow_dispatch:
    inputs:
      pr-ref:
        description: 'PR reference (number, URL, or owner/repo#number)'
        required: true

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Migrate PR from fork to branch
        uses: yourusername/git-mfpr@v1
        with:
          pr-ref: ${{ github.event.inputs.pr-ref }}
          dry-run: 'true'         # optional
          no-push: 'false'        # optional
          no-create: 'false'      # optional
          branch-name: ''         # optional
```

**Replace `yourusername` with your actual GitHub username.**

---

#### Inputs

| Name         | Description                                         | Required | Default  |
|--------------|-----------------------------------------------------|----------|----------|
| pr-ref       | PR reference (number, URL, or owner/repo#number)    | Yes      |          |
| dry-run      | Show what would happen without executing            | No       | false    |
| no-push      | Do not push branch                                  | No       | false    |
| no-create    | Do not create PR                                    | No       | false    |
| branch-name  | Custom branch name                                  | No       |          |

---

### Examples

#### Preview Migration

```bash
git-mfpr 123 --dry-run
```

Output:
```
🔄 Migrating PR #123...
📋 Title: Fix memory leak in worker pool
👤 Author: johndoe
🌿 Branch: migrated-123
$ git checkout main (dry-run)
$ git pull origin main (dry-run)
$ gh pr checkout 123 -b migrated-123 (dry-run)
$ git push -u origin migrated-123 (dry-run)
✅ Migration complete!
```

#### Custom Branch Name

```bash
git-mfpr 123 --branch-name feature/awesome-feature
```

#### Local Branch Only

```bash
git-mfpr 123 --no-push
```

## How It Works

1. **Fetches PR Information**: Uses `gh` CLI to get PR details
2. **Validates**: Ensures PR is from a fork and is still open
3. **Creates Branch**: Generates a branch name like `migrated-123`
4. **Checks Out Code**: Uses `gh pr checkout` to fetch the PR commits
5. **Pushes to Origin**: Pushes the new branch to your repository
6. **Suggests Next Steps**: Provides the command to create a new PR

## Branch Naming

By default, branches are named using the pattern:

```
migrated-{number}
```

For example:
- `migrated-123`
- `migrated-456`

This simple naming convention ensures consistent and predictable branch names.

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/cabljac/git-mfpr.git
cd git-mfpr

# Build
make build

# Run tests
make test

# Install locally
make install
```

### Project Structure

```
git-mfpr/
├── cmd/git-mfpr/        # CLI entry point
├── internal/
│   ├── git/             # Git operations
│   ├── github/          # GitHub API interactions
│   ├── migrate/         # Core migration logic
│   └── ui/              # Terminal UI
├── Makefile
└── README.md
```

## Error Handling

The tool provides clear error messages for common issues:

- **Not in a git repository**: Run the command from within a git repository
- **gh CLI not installed**: Install from https://cli.github.com/
- **PR not found**: Check the PR number and repository
- **Branch already exists**: Use `--branch-name` to specify a different name
- **PR not from fork**: Only PRs from forks need migration

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) for CLI framework
- [GitHub CLI](https://cli.github.com/) for GitHub API access