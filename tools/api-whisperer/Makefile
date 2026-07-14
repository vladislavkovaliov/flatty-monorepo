BINARY=api-whisperer

.PHONY: build run-openapi run-graphql run-combined run-mcp clean

build:
	go build -o $(BINARY) ./cmd/api-whisperer

run-openapi: build
	OLLAMA_BASE_URL=$(OLLAMA_BASE_URL) ./$(BINARY) --openapi ./examples/petstore.yaml --chat

run-graphql: build
	OLLAMA_BASE_URL=$(OLLAMA_BASE_URL) ./$(BINARY) --graphql ./examples/schema.graphql --chat

run-combined: build
	OLLAMA_BASE_URL=$(OLLAMA_BASE_URL) ./$(BINARY) --openapi ./examples/petstore.yaml --graphql ./examples/schema.graphql --chat

run-mcp: build
	./$(BINARY) --mcp --openapi ./examples/petstore.yaml

clean:
	rm -f $(BINARY)
