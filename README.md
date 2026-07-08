# flatty-budget-monorepo

## Tools

### review

Code review CLI that analyses staged git changes via Ollama.

```bash
cd tools/review
go build -o review .

./review                          # review staged changes in shop-graphql-nestjs/
./review --dir apps/react-launcher # review staged changes in a specific directory
./review --lang go                 # review Go code (default: TypeScript)
./review --model deepseek-coder:6.7b
./review --ollama-url http://192.168.1.85:11434
./review --help
```

### sql

Finds duplicate SQL between Go and NestJS repositories.

```bash
cd tools/sql
go build -o sql .

./sql
./sql --model deepseek-coder:6.7b
./sql --ollama-url http://192.168.1.85:11434
./sql --help
```
