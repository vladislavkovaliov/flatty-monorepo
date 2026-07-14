# api-whisperer

Go CLI + MCP tool for chatting with OpenAPI/Swagger and GraphQL API specifications via RAG (Ollama).

## Commands

```bash
make build                         # go build
make run-openapi                   # chat with OpenAPI spec
make run-graphql                   # chat with GraphQL schema
make run-combined                  # chat with both
make run-mcp                       # run as MCP server

# Manual:
./api-whisperer --openapi ./examples/petstore.yaml --chat
./api-whisperer --graphql ./examples/schema.graphql --chat
./api-whisperer --mcp --openapi ./examples/petstore.yaml
```

## Configuration

| Env var | Default | Description |
|---|---|---|
| `OLLAMA_BASE_URL` | `http://localhost:11434` | Ollama server |
| `OLLAMA_MODEL` | `mistral` | LLM model for generation |
| `OLLAMA_EMBEDDING_MODEL` | `nomic-embed-text` | Embedding model |

## Architecture

```
Makefile                   — build / run-openapi / run-graphql / run-combined / run-mcp
cmd/api-whisperer/main.go  — entrypoint, CLI flags, mode dispatch
internal/
  model/types.go            — shared data model (Endpoint, Schema, Field, Operation)
  llm/ollama.go             — HTTP client for Ollama (embed + generate)
  rag/splitter.go           — recursive character text splitter
  rag/vectorstore.go        — in-memory vector store + cosine similarity
  rag/pipeline.go           — RAG Pipeline: split + embed один раз при старте, embed(query) + search на каждый запрос
  openapi/parser.go         — OpenAPI 3.x + Swagger 2.0 parser (kin-openapi)
  graphql/parser.go         — GraphQL SDL + introspection parser (gqlparser)
  chat/chat.go              — interactive CLI chat with RAG + LLM
  mcp/server.go             — MCP stdio server for opencode integration
```

## CLI Chat

```
> какие эндпоинты работают с пользователями?
> покажи схему User
> сгенерируй Go структуру для Pet
> сделай Go HTTP клиент для POST /pets
```

### Chat commands

| Command | Description |
|---|---|
| `/help` | Show available commands |
| `/clear` | Clear conversation history |
| `/spec` | Show API spec summary |
| `/endpoints` | List all endpoints |
| `/schemas` | List all schemas |
| `/quit` | Exit |

## MCP Tools (opencode.json)

```json
{
  "mcp": {
    "api-whisperer": {
      "type": "local",
      "command": "./api-whisperer",
      "args": ["--mcp", "--openapi", "./examples/petstore.yaml"]
    }
  }
}
```

### Tools

| Tool | Params | Description |
|---|---|---|
| `query_api` | `query: string` | Ask a natural language question about the API |
| `describe_endpoint` | `method: string, path: string` | Describe a specific endpoint |
| `list_endpoints` | `tag?: string` | List all endpoints, optionally filtered by tag |
