# git-mfpr

A CLI tool that migrates GitHub pull requests from forks to branches in the main repository, preserving commit history and authorship.

## Overview

`git-mfpr` (Git Migrate Fork PR) helps maintainers bring pull requests from forked repositories into their main repository as regular branches. This is useful when you want to:

- Continue work on a PR after the original author has abandoned it
- Run CI/CD pipelines that don't work well with forks
- Collaborate more easily on a contribution
- Maintain a cleaner repository structure

## Features

- 🔄 Preserves complete commit history and authorship
- 🚀 Simple one-command migration
- 🎯 Smart branch naming with PR number, author, and title
- 🔍 Dry-run mode to preview actions
- 📦 Batch migration support for multiple PRs
- 🛡️ Safety checks to prevent accidental overwrites

## Installation

### Prerequisites

- Git
- [GitHub CLI (`gh`)](https://cli.github.com/) - must be installed and authenticated

### Install from Source

```bash
go install github.com/user/git-mfpr/cmd/git-mfpr@latest
```

### Download Binary

Download the latest release for your platform from the [releases page](https://github.com/user/git-mfpr/releases).

## Usage

### Basic Usage

```bash
# Migrate PR #123 from the current repository
git-mfpr 123

# Migrate multiple PRs
git-mfpr 123 124 125

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
🌿 Branch: pr-123-johndoe-fix-memory-leak
$ git checkout main (dry-run)
$ git pull origin main (dry-run)
$ gh pr checkout 123 -b pr-123-johndoe-fix-memory-leak (dry-run)
$ git push -u origin pr-123-johndoe-fix-memory-leak (dry-run)
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
3. **Creates Branch**: Generates a branch name like `pr-123-author-title`
4. **Checks Out Code**: Uses `gh pr checkout` to fetch the PR commits
5. **Pushes to Origin**: Pushes the new branch to your repository
6. **Suggests Next Steps**: Provides the command to create a new PR

## Branch Naming

By default, branches are named using the pattern:

```
pr-{number}-{author}-{title-slug}
```

For example:
- `pr-123-johndoe-fix-memory-leak`
- `pr-456-janedoe-add-new-feature`

The title is converted to lowercase, with special characters replaced by hyphens, and limited to 40 characters.

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/user/git-mfpr.git
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