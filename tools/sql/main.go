package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/vladislavkovaliov/shop-project/tools/sql/internal/sql"
)

type Config struct {
	Model     string
	OllamaURL string
}

const (
	defaultModel     = "deepseek-coder:6.7b"
	defaultOllamaURL = "http://192.168.1.85:11434"
)

func main() {
	cfg, showHelp := parseConfig(os.Args[1:])

	if showHelp {
		printHelp()
		return
	}

	repoRoot := findRepoRoot()
	runSQL(repoRoot, cfg)
}

func parseConfig(args []string) (cfg Config, help bool) {
	cfg = Config{
		Model:     defaultModel,
		OllamaURL: defaultOllamaURL,
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
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
	fmt.Println("    sql [flags]")
	fmt.Println()
	fmt.Println("  flags:")
	fmt.Printf("    --model <name>      Ollama model (default: %s)\n", defaultModel)
	fmt.Printf("    --ollama-url <url>  Ollama server URL (default: %s)\n", defaultOllamaURL)
	fmt.Println("    -h, --help          show this help")
	fmt.Println()
}

func findRepoRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "."
	}
	return strings.TrimSpace(string(out))
}

func runSQL(repoRoot string, cfg Config) {
	extractor := sql.NewExtractor(repoRoot, cfg.OllamaURL, cfg.Model)
	extractor.Run()
}
