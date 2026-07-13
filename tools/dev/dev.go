package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/vladislavkovaliov/shop-project/tools/dev/internal/context"
	"github.com/vladislavkovaliov/shop-project/tools/dev/internal/differ"
	"github.com/vladislavkovaliov/shop-project/tools/dev/internal/llm"
)

var codeBlockRe = regexp.MustCompile(`(?s)\x60{3}\w*\n(.*?)\n\x60{3}`)

func parseModifiedFiles(response string) []ModifiedFile {
	matches := codeBlockRe.FindAllStringSubmatch(response, -1)
	var result []ModifiedFile
	for _, m := range matches {
		content := strings.TrimSpace(m[1])
		if content == "" {
			continue
		}
		content = cleanContent(content)
		result = append(result, ModifiedFile{
			Path:    "",
			Content: content,
		})
	}
	return result
}

func cleanContent(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) > 0 {
		first := strings.TrimSpace(lines[0])
		if strings.HasPrefix(first, "<file") || strings.HasPrefix(first, "---") {
			lines = lines[1:]
		}
	}
	if len(lines) > 0 {
		last := strings.TrimSpace(lines[len(lines)-1])
		if last == "</file>" || last == "---" {
			lines = lines[:len(lines)-1]
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

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

	modified := parseModifiedFiles(response)
	if len(modified) == 0 {
		return
	}

	var allDiffs []string
	for i, m := range modified {
		if i >= len(files) {
			log.Printf("warning: more modified blocks than input files, skipping")
			break
		}
		origFile := files[i]
		origContent := origFile.Content
		filePath := origFile.Path

		if origContent == m.Content {
			log.Printf("  %s: unchanged", filePath)
			continue
		}

		diff, err := differ.UnifiedDiff(origContent, m.Content, filePath)
		if err != nil {
			log.Printf("  %s: diff error: %v", filePath, err)
			continue
		}

		if diff == "" {
			log.Printf("  %s: no diff (content identical)", filePath)
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

	b.WriteString(fmt.Sprintf("You are an expert %s engineer.\n\n", language))
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

	b.WriteString(`Output the modified file in a code block. No explanations.

` + "```" + `go
modified file content
` + "```" + `
`)

	return b.String()
}

type ModifiedFile struct {
	Path    string
	Content string
}

func applyDiff(repoRoot, diff string) error {
	cmd := exec.Command("git", "apply")
	cmd.Dir = repoRoot
	cmd.Stdin = strings.NewReader(diff)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git apply failed: %w\n%s", err, string(out))
	}
	return nil
}

func languageForPath(path string) string {
	if strings.HasSuffix(path, ".go") {
		return "go"
	}
	if strings.HasSuffix(path, ".tsx") {
		return "tsx"
	}
	if strings.HasSuffix(path, ".ts") {
		return "ts"
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
