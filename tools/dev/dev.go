package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/vladislavkovaliov/shop-project/tools/dev/internal/context"
	"github.com/vladislavkovaliov/shop-project/tools/dev/internal/differ"
	"github.com/vladislavkovaliov/shop-project/tools/dev/internal/llm"
)

var fileBlockRe = regexp.MustCompile(`(?s)\[FILE:\s*(.+?)\]\s*\x60{3}\w*\n(.*?)\n?\x60{3}`)

func runDev(repoRoot string, cfg Config) {
	log.SetPrefix("")
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	log.Printf("dev tool started")
	log.Printf("  repo root: %s", repoRoot)
	log.Printf("  plan: %s", cfg.Plan)
	log.Printf("  files: %s", strings.Join(cfg.Files, ", "))
	log.Printf("  language: %s", cfg.Language)

	files, err := context.ReadFiles(repoRoot, cfg.Files)
	if err != nil {
		log.Fatalf("reading files: %v", err)
	}

	prompt := buildPrompt(cfg.Plan, cfg.Language, files)
	log.Printf("prompt size: %d bytes", len(prompt))

	client := llm.NewClient(cfg.OllamaURL, cfg.Model)

	start := time.Now()
	log.Printf("sending to ollama...")
	response, err := client.Send(prompt)
	if err != nil {
		log.Fatalf("ollama failed: %v", err)
	}
	log.Printf("response received in %v (%d bytes)", time.Since(start).Round(time.Millisecond), len(response))

	modified := parseResponse(response)
	if len(modified) == 0 {
		log.Println("no modified files found in response")
		fmt.Println(response)
		return
	}

	var allDiffs []string
	for _, m := range modified {
		origContent := findOriginal(files, m.Path)
		if origContent == "" {
			log.Printf("warning: file %s not in original list, skipping", m.Path)
			continue
		}

		if origContent == m.Content {
			log.Printf("  %s: unchanged", m.Path)
			continue
		}

		diff, err := differ.UnifiedDiff(origContent, m.Content, m.Path)
		if err != nil {
			log.Printf("  %s: diff error: %v", m.Path, err)
			continue
		}

		if diff == "" {
			log.Printf("  %s: no diff (content identical)", m.Path)
			continue
		}

		allDiffs = append(allDiffs, diff)
	}

	if len(allDiffs) == 0 {
		fmt.Println("No changes generated.")
		return
	}

	output := strings.Join(allDiffs, "\n")

	if cfg.Apply {
		if err := applyDiff(repoRoot, output); err != nil {
			log.Fatalf("apply failed: %v", err)
		}
		fmt.Println("Changes applied successfully.")
		return
	}

	fmt.Print(output)
}

func buildPrompt(plan, language string, files []context.FileContent) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("You are a code generator for %s.\n\n", language))
	b.WriteString("Plan:\n")
	b.WriteString(plan)
	b.WriteString("\n\nFiles:\n\n")

	for _, f := range files {
		ext := languageForPath(f.Path)
		b.WriteString(fmt.Sprintf("=== %s ===\n", f.Path))
		b.WriteString(fmt.Sprintf("```%s\n", ext))
		b.WriteString(f.Content)
		if !strings.HasSuffix(f.Content, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("```\n\n")
	}

	b.WriteString(`Generate the modified version of each file according to the plan.
Output each modified file in this exact format:

[FILE: relative/path/to/file.tsx]
` + "```" + `tsx
entire modified file content
` + "```" + `

Rules:
- Output EVERY file from the input, even if unchanged.
- oldString and newString must match the indentation and style of the original.
- Make MINIMAL changes — only what the plan requires.
- Preserve ALL existing code, imports, and formatting.
- Do NOT add explanations outside the blocks.
`)

	return b.String()
}

type ModifiedFile struct {
	Path    string
	Content string
}

func parseResponse(response string) []ModifiedFile {
	matches := fileBlockRe.FindAllStringSubmatch(response, -1)
	var result []ModifiedFile
	for _, m := range matches {
		path := strings.TrimSpace(m[1])
		content := m[2]
		if path == "" {
			continue
		}
		result = append(result, ModifiedFile{
			Path:    path,
			Content: content,
		})
	}
	return result
}

func findOriginal(files []context.FileContent, path string) string {
	for _, f := range files {
		if f.Path == path {
			return f.Content
		}
	}
	return ""
}

func applyDiff(repoRoot, diff string) error {
	return nil
}

func languageForPath(path string) string {
	if strings.HasSuffix(path, ".go") {
		return "go"
	}
	if strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") {
		return "tsx"
	}
	if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".jsx") {
		return "jsx"
	}
	if strings.HasSuffix(path, ".py") {
		return "python"
	}
	if strings.HasSuffix(path, ".css") {
		return "css"
	}
	return ""
}
