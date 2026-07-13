# dev

Go CLI for generating code changes via Ollama.

Given a plan and file paths, uses a local LLM to generate code changes
and outputs unified diffs.

## Usage

```
cd tools/dev && go build -o dev .
./dev --plan "Add loading state to UserCard" \
      --files "apps/react-settings/src/UserCard.tsx" \
      --lang typescript
./dev --plan "Add validation" \
      --files "apps/go-api/handler.go" \
      --lang go --apply
./dev --help
```

## For AI agents (subagent: codegen)

The `codegen` subagent (`.opencode/agents/codegen.md`) MUST use this CLI tool.

Steps:
1. Build: `cd tools/dev && go build -o dev .`
2. Run: `tools/dev/dev --plan "<plan>" --files "<paths>" --lang <lang>`
3. Collect the diff output

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--plan` | — | Description of changes to make (**required**) |
| `--files` | — | Comma-separated file paths (**required**) |
| `--lang` | `TypeScript` | Programming language |
| `--model` | `deepseek-coder:6.7b` | Ollama model |
| `--ollama-url` | `http://192.168.1.85:11434` | Ollama server URL |
| `--apply` | `false` | Apply diff directly via git apply |
