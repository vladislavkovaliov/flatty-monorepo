# review

Go CLI for code review via Ollama.

## Usage

```
cd tools/review && go build -o review .
./review                  # review staged changes (default: apps/)
./review --dir apps/go-api      # review staged changes in a specific app
./review --lang go        # review Go code (default: TypeScript)
./review --help           # flags: --dir, --lang, --model, --ollama-url
```

## For AI agents (subagent: review)

The `review` subagent MUST use this CLI tool — do NOT perform manual review.

Steps:
1. Build: `cd tools/review && go build -o review .`
2. Run: `./review --dir <path> --lang <lang>`
3. Return the full output as the review result

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dir` | `apps/` | Directory to review |
| `--lang` | `TypeScript` | Language (go, typescript, javascript, python) |
| `--model` | `qwen2.5-coder:14b` | Ollama model |
| `--ollama-url` | `http://192.168.1.85:11434` | Ollama server URL |
