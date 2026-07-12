package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	Model     string
	OllamaURL string
	Directory string
	Language  string
}

const (
	defaultModel     = "deepseek-coder:6.7b"
	defaultOllamaURL = "http://192.168.1.85:11434"
	defaultDirectory = "apps/"
	defaultLanguage  = "TypeScript"
)

var langConfig = map[string]struct {
	Name string
	Exts []string
}{
	"go":         {"Go", []string{".go"}},
	"typescript": {"TypeScript", []string{".ts", ".tsx"}},
	"javascript": {"JavaScript", []string{".js", ".jsx"}},
	"python":     {"Python", []string{".py"}},
}

func main() {
	cfg, showHelp := parseConfig(os.Args[1:])

	if showHelp {
		printHelp()
		return
	}

	repoRoot := findRepoRoot()
	runReview(repoRoot, cfg)
}

func parseConfig(args []string) (cfg Config, help bool) {
	cfg = Config{
		Model:     defaultModel,
		OllamaURL: defaultOllamaURL,
		Directory: defaultDirectory,
		Language:  defaultLanguage,
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
		case "--dir":
			if i+1 < len(args) {
				i++
				cfg.Directory = args[i]
			}
		case "--lang":
			if i+1 < len(args) {
				i++
				if lang, ok := resolveLang(args[i]); ok {
					cfg.Language = lang.Name
				}
			}
		case "-h", "--help":
			help = true
			return
		}
	}

	return
}

func resolveLang(input string) (*struct {
	Name string
	Exts []string
}, bool) {
	lower := strings.ToLower(input)
	if lc, ok := langConfig[lower]; ok {
		return &lc, true
	}
	return nil, false
}

func extsFromLang(language string) []string {
	for _, lc := range langConfig {
		if strings.EqualFold(lc.Name, language) {
			return lc.Exts
		}
	}
	return nil
}

func printHelp() {
	fmt.Println()
	fmt.Println("  usage:")
	fmt.Println("    review [flags]")
	fmt.Println()
	fmt.Println("  flags:")
	fmt.Printf("    --dir <path>        directory to review (default: %s)\n", defaultDirectory)
	fmt.Printf("    --lang <lang>       programming language (default: %s)\n", defaultLanguage)
	fmt.Println("                         supported: go, typescript, javascript, python")
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
