package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/git-mfpr/internal/migrate"
	"github.com/user/git-mfpr/internal/ui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	dryRun     bool
	noPush     bool
	noCreate   bool
	branchName string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "git-mfpr [PR numbers or URLs]",
		Short: "Migrate GitHub pull requests from forks to branches",
		Long: `git-mfpr migrates GitHub pull requests from forks to branches in the main repository,
preserving commit history and authorship.

Examples:
  git mfpr 123                      # Migrate PR #123 from current repo
  git mfpr 123 124 125             # Migrate multiple PRs
  git mfpr owner/repo#123          # Migrate from specific repo
  git mfpr https://github.com/...  # Full URL support`,
		Args: cobra.MinimumNArgs(1),
		Run:  run,
	}

	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would happen without executing")
	rootCmd.Flags().BoolVar(&noPush, "no-push", false, "Create branch but don't push")
	rootCmd.Flags().BoolVar(&noCreate, "no-create", false, "Don't offer to create new PR")
	rootCmd.Flags().StringVar(&branchName, "branch-name", "", "Custom branch name (for single PR only)")

	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(_ *cobra.Command, args []string) {
	ui := ui.NewWithOptions(dryRun)

	opts := migrate.Options{
		DryRun:     dryRun,
		NoPush:     noPush,
		NoCreate:   noCreate,
		BranchName: branchName,
	}

	if branchName != "" && len(args) > 1 {
		ui.Error(fmt.Errorf("--branch-name can only be used with a single PR"))
		os.Exit(1)
	}

	migrator := migrate.NewMockMigrator(func(event migrate.Event) {
		ui.HandleEvent(event)
	})

	failed := false
	ctx := context.Background()
	for _, prRef := range args {
		ui.StartPR(prRef)

		if err := migrator.MigratePR(ctx, prRef, opts); err != nil {
			ui.Error(err)
			failed = true
			continue
		}
	}

	if failed {
		os.Exit(1)
	}
}
