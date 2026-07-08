package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vladislavkovaliov/shop-project/tools/review/internal/git"
	"github.com/vladislavkovaliov/shop-project/tools/review/internal/llm"
	"github.com/vladislavkovaliov/shop-project/tools/review/internal/report"
)

func runReview(repoRoot string, cfg Config) {
	log.SetPrefix("")
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	log.Printf("🚀 review tool started")
	log.Printf("  repo root: %s", repoRoot)
	log.Printf("  directory: %s", cfg.Directory)
	log.Printf("  language: %s", cfg.Language)

	exts := extsFromLang(cfg.Language)

	if git.IsClean(repoRoot, cfg.Directory) {
		report.PrintNoChanges()
		return
	}

	startTotal := time.Now()

	result, err := git.GetStagedDiff(repoRoot, cfg.Directory, exts...)
	if err != nil {
		report.PrintError(fmt.Sprintf("git diff failed: %v", err))
		os.Exit(1)
	}

	if len(result.Files) == 0 {
		report.PrintNoChanges()
		return
	}

	report.PrintHeader(result.Files)

	client := llm.NewClient(cfg.OllamaURL, cfg.Model)

	log.Printf("  sending to ollama...")
	review, err := client.Review(result.Diff, cfg.Language)
	if err != nil {
		report.PrintError(fmt.Sprintf("LLM review failed: %v", err))
		os.Exit(1)
	}

	totalElapsed := time.Since(startTotal).Round(time.Millisecond)
	log.Printf("  total time: %v", totalElapsed)

	report.PrintReview(cfg.Model, review)
}
