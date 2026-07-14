package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/vladislavkovaliov/api-whisperer/internal/llm"
	"github.com/vladislavkovaliov/api-whisperer/internal/model"
	"github.com/vladislavkovaliov/api-whisperer/internal/rag"
)

type SpecProvider struct {
	Spec     *model.ApiSpec
	Client   *llm.Client
	LLMModel string
	Pipeline *rag.Pipeline
}

func RunServer(provider *SpecProvider) {
	mcpServer := server.NewMCPServer(
		"api-whisperer",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	mcpServer.AddTool(newQueryAPITool(), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleQueryAPI(req, provider)
	})

	mcpServer.AddTool(newDescribeEndpointTool(), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleDescribeEndpoint(req, provider)
	})

	mcpServer.AddTool(newListEndpointsTool(), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleListEndpoints(req, provider)
	})

	if err := server.ServeStdio(mcpServer); err != nil {
		log.Printf("MCP server error: %v", err)
	}
}

func newQueryAPITool() mcp.Tool {
	return mcp.NewTool("query_api",
		mcp.WithDescription("Ask a natural language question about the API specification"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Your question about the API"),
		),
	)
}

func newDescribeEndpointTool() mcp.Tool {
	return mcp.NewTool("describe_endpoint",
		mcp.WithDescription("Describe a specific API endpoint in detail"),
		mcp.WithString("method",
			mcp.Required(),
			mcp.Description("HTTP method (GET, POST, PUT, DELETE, PATCH)"),
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("URL path (e.g., /pets/{id})"),
		),
	)
}

func newListEndpointsTool() mcp.Tool {
	return mcp.NewTool("list_endpoints",
		mcp.WithDescription("List all API endpoints, optionally filtered by tag"),
		mcp.WithString("tag",
			mcp.Description("Optional: filter by tag"),
		),
	)
}

func handleQueryAPI(req mcp.CallToolRequest, p *SpecProvider) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	ragCtx, err := p.Pipeline.Search(query, 5)
	if err != nil {
		ragCtx = ""
	}

	var b strings.Builder
	b.WriteString("You are api-whisperer, an AI assistant specialized in API specifications.\n")
	b.WriteString("You help developers understand and work with OpenAPI/Swagger and GraphQL APIs.\n")
	b.WriteString("Answer questions about endpoints, schemas, parameters.\n\n")

	b.WriteString(fmt.Sprintf("## API Specification\n\n"))
	b.WriteString(fmt.Sprintf("Title: %s\n", p.Spec.Title))
	b.WriteString(fmt.Sprintf("Version: %s\n", p.Spec.Version))
	b.WriteString(fmt.Sprintf("Source: %s\n", p.Spec.Source))
	b.WriteString(fmt.Sprintf("Endpoints: %d\n", len(p.Spec.Endpoints)))
	b.WriteString(fmt.Sprintf("Schemas: %d\n", len(p.Spec.Schemas)))
	b.WriteString("\n")

	if ragCtx != "" {
		b.WriteString("## Relevant API Context\n\n")
		b.WriteString(ragCtx)
		b.WriteString("\n")
	}

	b.WriteString("## Query\n")
	b.WriteString(query)
	b.WriteString("\n\n")
	b.WriteString("## Response\n")
	b.WriteString("Answer based ONLY on the API specification context provided above.\n")
	b.WriteString("Do NOT invent or assume parameters, endpoints, or schemas that are not explicitly listed.\n")
	b.WriteString("If the spec does not have what the user asks about, say so directly.\n")
	b.WriteString("Be concise but thorough.\n")

	response, err := p.Client.Generate(p.LLMModel, b.String())
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("generate response: %v", err)), nil
	}

	return mcp.NewToolResultText(strings.TrimSpace(response)), nil
}

func handleDescribeEndpoint(req mcp.CallToolRequest, p *SpecProvider) (*mcp.CallToolResult, error) {
	method, err := req.RequireString("method")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	method = strings.ToUpper(method)

	for _, ep := range p.Spec.Endpoints {
		if strings.EqualFold(ep.Method, method) && normalizePath(ep.Path) == normalizePath(path) {
			return describeEndpoint(&ep), nil
		}
	}

	return mcp.NewToolResultError(fmt.Sprintf("endpoint %s %s not found", method, path)), nil
}

func handleListEndpoints(req mcp.CallToolRequest, p *SpecProvider) (*mcp.CallToolResult, error) {
	tag := req.GetString("tag", "")

	var b strings.Builder
	b.WriteString("## Endpoints\n\n")

	for _, ep := range p.Spec.Endpoints {
		if tag != "" {
			hasTag := false
			for _, t := range ep.Tags {
				if strings.EqualFold(t, tag) {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		tags := ""
		if len(ep.Tags) > 0 {
			tags = fmt.Sprintf(" [%s]", strings.Join(ep.Tags, ", "))
		}
		b.WriteString(fmt.Sprintf("- **%s** `%s`", ep.Method, ep.Path))
		if ep.Summary != "" {
			b.WriteString(fmt.Sprintf(" — %s", ep.Summary))
		}
		b.WriteString(tags)
		b.WriteString("\n")
	}

	return mcp.NewToolResultText(b.String()), nil
}

func describeEndpoint(ep *model.Endpoint) *mcp.CallToolResult {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s %s\n\n", ep.Method, ep.Path))

	if ep.Summary != "" {
		b.WriteString(fmt.Sprintf("**Summary:** %s\n\n", ep.Summary))
	}
	if ep.Description != "" {
		b.WriteString(fmt.Sprintf("**Description:** %s\n\n", ep.Description))
	}

	if len(ep.Tags) > 0 {
		b.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(ep.Tags, ", ")))
	}

	if len(ep.Params) > 0 {
		b.WriteString("### Parameters\n\n")
		b.WriteString("| Name | In | Type | Required | Description |\n")
		b.WriteString("|------|----|------|----------|-------------|\n")
		for _, p := range ep.Params {
			req := "No"
			if p.Required {
				req = "Yes"
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n", p.Name, p.In, p.Type, req, p.Description))
		}
		b.WriteString("\n")
	}

	if ep.RequestBody != nil && ep.RequestBody.Schema != nil {
		b.WriteString("### Request Body\n\n")
		b.WriteString(fmt.Sprintf("Schema: **%s**\n\n", ep.RequestBody.Schema.Name))
		json, _ := json.MarshalIndent(ep.RequestBody.Schema, "", "  ")
		b.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", string(json)))
	}

	if len(ep.Responses) > 0 {
		b.WriteString("### Responses\n\n")
		b.WriteString("| Code | Description | Schema |\n")
		b.WriteString("|------|-------------|--------|\n")
		for _, r := range ep.Responses {
			schemaName := ""
			if r.Schema != nil && r.Schema.Schema != nil {
				schemaName = r.Schema.Schema.Name
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", r.StatusCode, r.Description, schemaName))
		}
		b.WriteString("\n")
	}

	return mcp.NewToolResultText(b.String())
}

func normalizePath(path string) string {
	return strings.TrimRight(path, "/")
}

func listSchemaNames(schemas []model.Schema) string {
	var names []string
	for _, s := range schemas {
		names = append(names, s.Name)
	}
	return strings.Join(names, ", ")
}

func listTags(endpoints []model.Endpoint) string {
	tagSet := make(map[string]bool)
	for _, ep := range endpoints {
		for _, t := range ep.Tags {
			tagSet[t] = true
		}
	}
	var tags []string
	for t := range tagSet {
		tags = append(tags, t)
	}
	return strings.Join(tags, ", ")
}
