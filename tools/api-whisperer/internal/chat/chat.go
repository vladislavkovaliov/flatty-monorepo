package chat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/vladislavkovaliov/api-whisperer/internal/llm"
	"github.com/vladislavkovaliov/api-whisperer/internal/model"
	"github.com/vladislavkovaliov/api-whisperer/internal/rag"
	"golang.org/x/term"
)

type Session struct {
	spec     *model.ApiSpec
	client   *llm.Client
	llmModel string
	pipeline *rag.Pipeline
	history  []chatMessage
}

type chatMessage struct {
	Role    string
	Content string
}

func New(spec *model.ApiSpec, client *llm.Client, llmModel string, pipeline *rag.Pipeline) *Session {
	return &Session{
		spec:     spec,
		client:   client,
		llmModel: llmModel,
		pipeline: pipeline,
	}
}

func (s *Session) Run(ctx context.Context) error {
	fmt.Println(s.green("\n╭──────────────────────────────────────────╮"))
	fmt.Println(s.green("│       api-whisperer — chat with your API  │"))
	fmt.Println(s.green("╰──────────────────────────────────────────╯"))
	fmt.Println()
	fmt.Printf("Loaded: %s %s (%s)\n", s.spec.Title, s.spec.Version, s.spec.Source)
	fmt.Printf("Endpoints: %d | Schemas: %d | Operations: %d\n",
		len(s.spec.Endpoints), len(s.spec.Schemas), len(s.spec.Operations))
	fmt.Println()
	fmt.Println(s.cyan("Commands: /help  /clear  /spec  /endpoints  /schemas  /quit"))
	fmt.Println()

	if term.IsTerminal(int(os.Stdin.Fd())) {
		s.runTerminal(ctx)
	} else {
		s.runPipe(ctx)
	}

	return nil
}

func (s *Session) runTerminal(ctx context.Context) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(s.cyan("> "))

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if strings.HasPrefix(input, "/") {
			s.handleCommand(input)
			continue
		}

		s.handleMessage(ctx, input)
	}
}

func (s *Session) runPipe(ctx context.Context) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "/") {
			s.handleCommand(line)
			continue
		}
		s.handleMessage(ctx, line)
	}
}

func (s *Session) handleCommand(cmd string) {
	switch {
	case cmd == "/help" || cmd == "/h":
		fmt.Println(s.yellow("Commands:"))
		fmt.Println("  /help       Show this help")
		fmt.Println("  /clear      Clear conversation history")
		fmt.Println("  /spec       Show API spec summary")
		fmt.Println("  /endpoints  List all endpoints")
		fmt.Println("  /schemas    List all schemas")
		fmt.Println("  /quit       Exit")

	case cmd == "/clear":
		s.history = nil
		fmt.Println(s.green("Conversation history cleared."))

	case cmd == "/spec":
		fmt.Printf("Title: %s\n", s.spec.Title)
		fmt.Printf("Version: %s\n", s.spec.Version)
		fmt.Printf("Source: %s\n", s.spec.Source)
		fmt.Printf("Endpoints: %d\n", len(s.spec.Endpoints))
		fmt.Printf("Schemas: %d\n", len(s.spec.Schemas))
		fmt.Printf("Operations: %d\n", len(s.spec.Operations))

	case cmd == "/endpoints":
		if len(s.spec.Endpoints) == 0 {
			fmt.Println(s.yellow("No endpoints found."))
			return
		}
		for _, ep := range s.spec.Endpoints {
			tags := ""
			if len(ep.Tags) > 0 {
				tags = fmt.Sprintf(" [%s]", strings.Join(ep.Tags, ", "))
			}
			fmt.Printf("  %s %s%s\n", s.green(ep.Method), ep.Path, s.cyan(tags))
		}

	case cmd == "/schemas":
		if len(s.spec.Schemas) == 0 {
			fmt.Println(s.yellow("No schemas found."))
			return
		}
		for _, sch := range s.spec.Schemas {
			fmt.Printf("  %s (%s): %d fields\n", s.green(sch.Name), sch.Kind, len(sch.Fields))
			if len(sch.Values) > 0 {
				fmt.Printf("    Values: %s\n", strings.Join(sch.Values, ", "))
			}
		}

	case cmd == "/quit" || cmd == "/q" || cmd == "/exit":
		fmt.Println("Bye!")
		os.Exit(0)

	default:
		fmt.Printf("Unknown command: %s. Type /help for available commands.\n", cmd)
	}
}

func (s *Session) handleMessage(ctx context.Context, input string) {
	s.history = append(s.history, chatMessage{Role: "user", Content: input})

	ragCtx, err := s.pipeline.Search(input, 5)
	if err != nil {
		ragCtx = ""
	}

	prompt := buildPrompt(s.spec, input, ragCtx, s.history)

	response, err := s.client.Generate(s.llmModel, prompt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	response = strings.TrimSpace(response)
	fmt.Println(response)
	fmt.Println()

	s.history = append(s.history, chatMessage{Role: "assistant", Content: response})
}

func buildPrompt(spec *model.ApiSpec, query, ragCtx string, history []chatMessage) string {
	var b strings.Builder

	b.WriteString("You are api-whisperer, an AI assistant specialized in API specifications.\n")
	b.WriteString("You help developers understand and work with OpenAPI/Swagger and GraphQL APIs.\n")
	b.WriteString("Answer questions about endpoints, schemas, parameters, and generate Go code when asked.\n\n")

	b.WriteString("## API Specification\n\n")
	b.WriteString(fmt.Sprintf("Title: %s\n", spec.Title))
	b.WriteString(fmt.Sprintf("Version: %s\n", spec.Version))
	b.WriteString(fmt.Sprintf("Source: %s\n", spec.Source))
	b.WriteString(fmt.Sprintf("Endpoints: %d\n", len(spec.Endpoints)))
	b.WriteString(fmt.Sprintf("Schemas: %d\n", len(spec.Schemas)))
	if len(spec.Operations) > 0 {
		b.WriteString(fmt.Sprintf("Operations: %d\n", len(spec.Operations)))
	}
	b.WriteString("\n")

	if ragCtx != "" {
		b.WriteString("## Relevant API Context\n\n")
		b.WriteString(ragCtx)
		b.WriteString("\n")
	}

	for i := len(history) - 3; i < len(history); i++ {
		if i < 0 {
			continue
		}
		msg := history[i]
		if msg.Role == "user" {
			b.WriteString(fmt.Sprintf("## User\n%s\n\n", msg.Content))
		} else {
			b.WriteString(fmt.Sprintf("## Assistant\n%s\n\n", msg.Content))
		}
	}

	b.WriteString("## Query\n")
	b.WriteString(query)
	b.WriteString("\n\n")
	b.WriteString("## Response\n")
	b.WriteString("Answer based ONLY on the API specification context provided above.\n")
	b.WriteString("Do NOT invent or assume parameters, endpoints, or schemas that are not explicitly listed.\n")
	b.WriteString("If the spec does not have what the user asks about, say so directly.\n")
	b.WriteString("If the user asks for Typescript code, provide complete, compilable Typescript code.\n")
	b.WriteString("Be concise but thorough.\n")

	return b.String()
}

func writeField(b *strings.Builder, f model.Field, indent string) {
	req := ""
	if f.Required {
		req = " (required)"
	}
	arr := ""
	if f.IsArray {
		arr = "[]"
	}
	b.WriteString(fmt.Sprintf("%s- %s: %s%s%s\n", indent, f.Name, arr, f.Type, req))
}

func SerializeSpec(spec *model.ApiSpec) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("API: %s v%s\n", spec.Title, spec.Version))

	b.WriteString("\n=== ENDPOINTS ===\n")
	for _, ep := range spec.Endpoints {
		b.WriteString(fmt.Sprintf("%s %s\n", ep.Method, ep.Path))
		if ep.Summary != "" {
			b.WriteString(fmt.Sprintf("  Summary: %s\n", ep.Summary))
		}
		if len(ep.Tags) > 0 {
			b.WriteString(fmt.Sprintf("  Tags: %s\n", strings.Join(ep.Tags, ", ")))
		}
		for _, p := range ep.Params {
			b.WriteString(fmt.Sprintf("  Param: %s (%s, %s)\n", p.Name, p.In, p.Type))
		}
		if ep.RequestBody != nil && ep.RequestBody.Schema != nil {
			b.WriteString(fmt.Sprintf("  Body: %s\n", ep.RequestBody.Schema.Name))
			for _, f := range ep.RequestBody.Schema.Fields {
				writeField(&b, f, "    ")
			}
		}
		for _, r := range ep.Responses {
			b.WriteString(fmt.Sprintf("  Response %s: %s\n", r.StatusCode, r.Description))
			if r.Schema != nil && r.Schema.Schema != nil {
				for _, f := range r.Schema.Schema.Fields {
					writeField(&b, f, "    ")
				}
			}
		}
	}

	b.WriteString("\n=== SCHEMAS ===\n")
	for _, s := range spec.Schemas {
		b.WriteString(fmt.Sprintf("%s (%s)\n", s.Name, s.Kind))
		if s.Description != "" {
			b.WriteString(fmt.Sprintf("  Description: %s\n", s.Description))
		}
		for _, f := range s.Fields {
			req := ""
			if f.Required {
				req = " (required)"
			}
			arr := ""
			if f.IsArray {
				arr = "[]"
			}
			b.WriteString(fmt.Sprintf("  - %s: %s%s%s\n", f.Name, arr, f.Type, req))
		}
		if len(s.Values) > 0 {
			b.WriteString(fmt.Sprintf("  Values: %s\n", strings.Join(s.Values, ", ")))
		}
	}

	if len(spec.Operations) > 0 {
		b.WriteString("\n=== OPERATIONS (GraphQL) ===\n")
		for _, op := range spec.Operations {
			b.WriteString(fmt.Sprintf("%s: %s -> %s\n", op.Kind, op.Name, op.Type))
		}
	}

	return b.String()
}

func (s *Session) green(text string) string {
	return "\033[32m" + text + "\033[0m"
}

func (s *Session) cyan(text string) string {
	return "\033[36m" + text + "\033[0m"
}

func (s *Session) yellow(text string) string {
	return "\033[33m" + text + "\033[0m"
}
