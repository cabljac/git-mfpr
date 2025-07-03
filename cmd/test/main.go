package main

import (
	"fmt"
	"log"
	"os"

	"github.com/user/git-mfpr/internal/migrate"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test <pr-ref>")
		os.Exit(1)
	}

	// Create a migrator
	m := migrate.New()

	// Set up event handler to display progress
	m.SetEventHandler(func(event migrate.Event) {
		switch event.Type {
		case "info":
			if event.Message != "" {
				fmt.Printf("→ %s\n", event.Message)
			}
		case "success":
			fmt.Printf("✅ %s\n", event.Message)
		case "error":
			fmt.Printf("❌ %s\n", event.Message)
		case "command":
			if event.Message != "" {
				fmt.Printf("  %s\n", event.Message)
			}
			if event.Detail != "" {
				fmt.Printf("  $ %s\n", event.Detail)
			}
		}
	})

	// Run migration in dry-run mode
	err := m.MigratePR(os.Args[1], migrate.Options{
		DryRun: true,
	})
	
	if err != nil {
		log.Fatal(err)
	}
}