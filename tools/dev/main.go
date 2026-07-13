package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	Plan      string
	Files     []string
	Language  string
	Model     string
	OllamaURL string
	Apply     bool
}

const (
	defaultModel     = "deepseek-coder:6.7b"
	defaultOllamaURL = "http://192.168.1.85:11434"
	defaultLanguage  = "TypeScript"
)

func main() {
	cfg, showHelp := parseConfig(os.Args[1:])

	if showHelp {
		printHelp()
		return
	}

	if cfg.Plan == "" {
		fmt.Fprintln(os.Stderr, "Error: --plan is required")
		os.Exit(1)
	}

	if len(cfg.Files) == 0 {
		fmt.Fprintln(os.Stderr, "Error: --files is required")
		os.Exit(1)
	}

	repoRoot := findRepoRoot()
	runDev(repoRoot, cfg)
}

func parseConfig(args []string) (cfg Config, help bool) {
	cfg = Config{
		Model:     defaultModel,
		OllamaURL: defaultOllamaURL,
		Language:  defaultLanguage,
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--plan":
			if i+1 < len(args) {
				i++
				cfg.Plan = args[i]
			}
		case "--files":
			if i+1 < len(args) {
				i++
				cfg.Files = strings.Split(args[i], ",")
			}
		case "--lang":
			if i+1 < len(args) {
				i++
				cfg.Language = args[i]
			}
		case "--model":
			if i+1 < len(args) {
				i++
				cfg.Model = args[i]
			}
		case "--ollama-url":
			if i+1 < len(args) {
				i++
				cfg.OllamaURL = args[i]
			}
		case "--apply":
			cfg.Apply = true
		case "-h", "--help":
			help = true
			return
		}
	}

	return
}

func printHelp() {
	fmt.Println()
	fmt.Println("  usage:")
	fmt.Println("    dev --plan <description> --files <paths> [flags]")
	fmt.Println()
	fmt.Println("  required:")
	fmt.Println("    --plan <str>       description of changes to make")
	fmt.Println("    --files <paths>    comma-separated file paths")
	fmt.Println()
	fmt.Println("  flags:")
	fmt.Printf("    --lang <lang>      programming language (default: %s)\n", defaultLanguage)
	fmt.Printf("    --model <name>     Ollama model (default: %s)\n", defaultModel)
	fmt.Printf("    --ollama-url <url> Ollama server URL (default: %s)\n", defaultOllamaURL)
	fmt.Println("    --apply            apply diff directly via git apply")
	fmt.Println("    -h, --help         show this help")
	fmt.Println()
}

func findRepoRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "."
	}
	return strings.TrimSpace(string(out))
}
