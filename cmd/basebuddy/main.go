package main

import (
	"basebuddy/internal/config"
	"basebuddy/internal/service"
	"flag"
	"log"
)

func main() {
    // Parse command-line flags
    promptFile := flag.String("prompt", "", "Path to the prompt template file")
    flag.Parse()

    if *promptFile == "" {
        log.Fatal("Prompt file is required")
    }

    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("could not load config: %v", err)
    }

    // Generate documentation
    if err := service.Run(*promptFile, cfg); err != nil {
        log.Fatalf("could not generate docs: %v", err)
    }

    log.Println("Documentation generation complete.")
}
