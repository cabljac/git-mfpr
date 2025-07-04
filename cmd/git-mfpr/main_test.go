package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/user/git-mfpr/internal/migrate"
)

// Mock migrator for testing
type mockMigrator struct {
	migratePRFunc       func(ctx context.Context, prRef string, opts migrate.Options) error
	migratePRsFunc      func(ctx context.Context, prRefs []string, opts migrate.Options) error
	getPRInfoFunc       func(ctx context.Context, prRef string) (*migrate.PRInfo, error)
	generateBranchFunc  func(pr *migrate.PRInfo) string
	eventHandler        migrate.EventHandler
	setEventHandlerFunc func(handler migrate.EventHandler)
}

func (m *mockMigrator) MigratePR(ctx context.Context, prRef string, opts migrate.Options) error {
	if m.migratePRFunc != nil {
		return m.migratePRFunc(ctx, prRef, opts)
	}
	return nil
}

func (m *mockMigrator) MigratePRs(ctx context.Context, prRefs []string, opts migrate.Options) error {
	if m.migratePRsFunc != nil {
		return m.migratePRsFunc(ctx, prRefs, opts)
	}
	return nil
}

func (m *mockMigrator) GetPRInfo(ctx context.Context, prRef string) (*migrate.PRInfo, error) {
	if m.getPRInfoFunc != nil {
		return m.getPRInfoFunc(ctx, prRef)
	}
	return &migrate.PRInfo{Number: 123}, nil
}

func (m *mockMigrator) GenerateBranchName(pr *migrate.PRInfo) string {
	if m.generateBranchFunc != nil {
		return m.generateBranchFunc(pr)
	}
	return "migrated-123"
}

func (m *mockMigrator) SetEventHandler(handler migrate.EventHandler) {
	m.eventHandler = handler
	if m.setEventHandlerFunc != nil {
		m.setEventHandlerFunc(handler)
	}
}

// Mock UI for testing
type mockUI struct {
	startPRCalls []string
	errors       []error
	events       []migrate.Event
}

func (m *mockUI) StartPR(prRef string) {
	m.startPRCalls = append(m.startPRCalls, prRef)
}

func (m *mockUI) Error(err error) {
	m.errors = append(m.errors, err)
}

func (m *mockUI) Success(message string) {}
func (m *mockUI) Info(message string)    {}
func (m *mockUI) Command(command string) {}

func (m *mockUI) HandleEvent(event migrate.Event) {
	m.events = append(m.events, event)
}

func TestRunMigration(t *testing.T) {
	// Save original values
	origDryRun := dryRun
	origNoPush := noPush
	origNoCreate := noCreate
	origBranchName := branchName

	// Restore after test
	defer func() {
		dryRun = origDryRun
		noPush = origNoPush
		noCreate = origNoCreate
		branchName = origBranchName
	}()

	tests := []struct {
		name          string
		args          []string
		dryRun        bool
		noPush        bool
		noCreate      bool
		branchName    string
		migratePRFunc func(ctx context.Context, prRef string, opts migrate.Options) error
		expectError   bool
		expectErrMsg  string
	}{
		{
			name:       "single PR success",
			args:       []string{"123"},
			dryRun:     false,
			noPush:     false,
			noCreate:   false,
			branchName: "",
			migratePRFunc: func(ctx context.Context, prRef string, opts migrate.Options) error {
				return nil
			},
			expectError: false,
		},
		{
			name:       "single PR with custom branch",
			args:       []string{"123"},
			dryRun:     false,
			noPush:     false,
			noCreate:   false,
			branchName: "custom-branch",
			migratePRFunc: func(ctx context.Context, prRef string, opts migrate.Options) error {
				if opts.BranchName != "custom-branch" {
					t.Errorf("Expected branch name 'custom-branch', got %s", opts.BranchName)
				}
				return nil
			},
			expectError: false,
		},
		{
			name:         "multiple PRs with custom branch should fail",
			args:         []string{"123", "124"},
			dryRun:       false,
			noPush:       false,
			noCreate:     false,
			branchName:   "custom-branch",
			expectError:  true,
			expectErrMsg: "--branch-name can only be used with a single PR",
		},
		{
			name:       "PR migration failure",
			args:       []string{"123"},
			dryRun:     false,
			noPush:     false,
			noCreate:   false,
			branchName: "",
			migratePRFunc: func(ctx context.Context, prRef string, opts migrate.Options) error {
				return errors.New("migration failed")
			},
			expectError:  true,
			expectErrMsg: "one or more migrations failed",
		},
		{
			name:       "dry run mode",
			args:       []string{"123"},
			dryRun:     true,
			noPush:     false,
			noCreate:   false,
			branchName: "",
			migratePRFunc: func(ctx context.Context, prRef string, opts migrate.Options) error {
				if !opts.DryRun {
					t.Error("Expected dry run mode to be enabled")
				}
				return nil
			},
			expectError: false,
		},
		{
			name:       "no-push and no-create flags",
			args:       []string{"123"},
			dryRun:     false,
			noPush:     true,
			noCreate:   true,
			branchName: "",
			migratePRFunc: func(ctx context.Context, prRef string, opts migrate.Options) error {
				if !opts.NoPush {
					t.Error("Expected no-push to be enabled")
				}
				if !opts.NoCreate {
					t.Error("Expected no-create to be enabled")
				}
				return nil
			},
			expectError: false,
		},
		{
			name:       "multiple PRs with partial failure",
			args:       []string{"123", "124", "125"},
			dryRun:     false,
			noPush:     false,
			noCreate:   false,
			branchName: "",
			migratePRFunc: func(ctx context.Context, prRef string, opts migrate.Options) error {
				if prRef == "124" {
					return errors.New("PR 124 failed")
				}
				return nil
			},
			expectError:  true,
			expectErrMsg: "one or more migrations failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set flags
			dryRun = tt.dryRun
			noPush = tt.noPush
			noCreate = tt.noCreate
			branchName = tt.branchName

			// Create mocks
			mockUI := &mockUI{}
			mockMigrator := &mockMigrator{
				migratePRFunc: tt.migratePRFunc,
			}

			// Run migration
			err := runMigration(tt.args, mockUI, mockMigrator)

			// Check error expectations
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.expectErrMsg != "" && !strings.Contains(err.Error(), tt.expectErrMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.expectErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Verify UI interactions
			// When branch name validation fails, we don't call StartPR
			expectedStartPRCalls := len(tt.args)
			if tt.branchName != "" && len(tt.args) > 1 {
				expectedStartPRCalls = 0
			}
			if len(mockUI.startPRCalls) != expectedStartPRCalls {
				t.Errorf("Expected %d StartPR calls, got %d", expectedStartPRCalls, len(mockUI.startPRCalls))
			}

			// Verify error reporting
			if tt.expectError && tt.migratePRFunc != nil {
				// Count how many PRs should have failed
				failedCount := 0
				for _, prRef := range tt.args {
					if err := tt.migratePRFunc(context.Background(), prRef, migrate.Options{}); err != nil {
						failedCount++
					}
				}
				if len(mockUI.errors) != failedCount {
					t.Errorf("Expected %d errors to be reported to UI, got %d", failedCount, len(mockUI.errors))
				}
			}
		})
	}
}

func TestVersionInfo(t *testing.T) {
	// Test that version information is set correctly
	if version != "dev" {
		t.Errorf("Expected version to be 'dev', got %s", version)
	}

	if commit != "none" {
		t.Errorf("Expected commit to be 'none', got %s", commit)
	}

	if date != "unknown" {
		t.Errorf("Expected date to be 'unknown', got %s", date)
	}
}

func TestEventHandlerIntegration(t *testing.T) {
	// Test that event handler is properly wired up
	mockUI := &mockUI{}
	mockMigrator := &mockMigrator{}

	// Capture the event handler
	var capturedHandler migrate.EventHandler
	mockMigrator.migratePRFunc = func(ctx context.Context, prRef string, opts migrate.Options) error {
		// Emit some events through the handler
		if capturedHandler != nil {
			capturedHandler(migrate.Event{Type: migrate.EventInfo, Message: "Test event"})
			capturedHandler(migrate.Event{Type: migrate.EventSuccess, Message: "Success"})
		}
		return nil
	}
	mockMigrator.setEventHandlerFunc = func(handler migrate.EventHandler) {
		capturedHandler = handler
	}

	// Run migration
	dryRun = false
	noPush = false
	noCreate = false
	branchName = ""

	err := runMigration([]string{"123"}, mockUI, mockMigrator)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify events were handled
	if len(mockUI.events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(mockUI.events))
	}

	if len(mockUI.events) > 0 && mockUI.events[0].Message != "Test event" {
		t.Errorf("Expected first event message to be 'Test event', got %s", mockUI.events[0].Message)
	}
}
