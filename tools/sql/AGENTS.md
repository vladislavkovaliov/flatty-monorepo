# sql

Go CLI for finding duplicate SQL between Go and NestJS codebases.

## Usage

```
cd tools/sql && go build -o sql .
./sql                     # find duplicate SQL
./sql --help              # flags: --model, --ollama-url
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--model` | `deepseek-coder:6.7b` | Ollama model |
| `--ollama-url` | `http://192.168.1.85:11434` | Ollama server URL |
