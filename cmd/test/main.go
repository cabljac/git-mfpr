package main

import (
	"context"
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

	m := migrate.New()

	m.SetEventHandler(func(event migrate.Event) {
		switch event.Type {
		case migrate.EventInfo:
			if event.Message != "" {
				fmt.Printf("→ %s\n", event.Message)
			}
		case migrate.EventSuccess:
			fmt.Printf("✅ %s\n", event.Message)
		case migrate.EventError:
			fmt.Printf("❌ %s\n", event.Message)
		case migrate.EventCommand:
			if event.Message != "" {
				fmt.Printf("  %s\n", event.Message)
			}
			if event.Detail != "" {
				fmt.Printf("  $ %s\n", event.Detail)
			}
		}
	})

	ctx := context.Background()
	err := m.MigratePR(ctx, os.Args[1], migrate.Options{
		DryRun: true,
	})
	if err != nil {
		log.Fatal(err)
	}
}
