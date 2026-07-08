package sql

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/vladislavkovaliov/shop-project/tools/sql/internal/llm"
)

type Extractor struct {
	repoRoot  string
	ollamaURL string
	model     string
}

func NewExtractor(repoRoot, ollamaURL, model string) *Extractor {
	return &Extractor{
		repoRoot:  repoRoot,
		ollamaURL: ollamaURL,
		model:     model,
	}
}

func (e *Extractor) Run() {
	log.Printf("scanning Go repository...")
	goBlocks, err := ScanGoRepo(e.repoRoot)
	if err != nil {
		log.Printf("error scanning Go: %v", err)
		return
	}
	log.Printf("  found %d SQL blocks", len(goBlocks))

	log.Printf("scanning NestJS repository...")
	tsBlocks, err := ScanTSRepo(e.repoRoot)
	if err != nil {
		log.Printf("error scanning NestJS: %v", err)
		return
	}
	log.Printf("  found %d SQL blocks", len(tsBlocks))

	groups := GroupBlocks(goBlocks, tsBlocks)
	groups = filterExtracted(groups)

	if len(groups) == 0 {
		fmt.Println()
		fmt.Println("  No duplicate SQL found between Go and NestJS repositories.")
		fmt.Println()
		return
	}

	log.Printf("  found %d duplicate SQL group(s) (after filtering already extracted)", len(groups))

	client := llm.NewClient(e.ollamaURL, e.model)

	for i, g := range groups {
		e.promptGroup(i+1, len(groups), g, client)
	}
}

func (e *Extractor) promptGroup(idx, total int, g MatchGroup, client *llm.Client) {
	const sep = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

	sql := g.GoBlock.RawSQL

	fmt.Println()
	fmt.Println(sep)
	fmt.Printf("  Duplicate SQL (%d/%d)\n", idx, total)
	fmt.Println(sep)
	fmt.Println()
	fmt.Printf("  Go:       %s:%d\n", relPath(e.repoRoot, g.GoBlock.FilePath), g.GoBlock.Line)
	fmt.Printf("  NestJS:   %s:%d\n", relPath(e.repoRoot, g.TSBlock.FilePath), g.TSBlock.Line)
	fmt.Println()
	fmt.Println("  SQL:")
	for _, line := range strings.Split(sql, "\n") {
		fmt.Printf("    %s\n", line)
	}
	fmt.Println()

	fmt.Print("  Extract to PostgreSQL function? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" && answer != "yes" {
		fmt.Println("  skipped")
		return
	}

	genSQL := e.generateFunction(sql, client)
	fmt.Println()
	fmt.Println(genSQL)
	fmt.Println()
	fmt.Println("  # Add to: shop-api-golang/restore.sql (before dump complete)")
	fmt.Println()
}

func isAlreadyExtracted(sql string) bool {
	lower := strings.ToLower(strings.TrimSpace(sql))
	return strings.HasPrefix(lower, "select * from get_") ||
		strings.HasPrefix(lower, "select * from user_total_spent") ||
		strings.HasPrefix(lower, "select id, name, email, total_spent from user_total_spent")
}

func filterExtracted(groups []MatchGroup) []MatchGroup {
	var filtered []MatchGroup
	for _, g := range groups {
		sql := strings.TrimSpace(strings.ToLower(g.GoBlock.RawSQL))
		if strings.HasPrefix(sql, "select * from get_") ||
			strings.HasPrefix(sql, "select id, name, email, total_spent from user_total_spent") ||
			strings.HasPrefix(sql, "select * from user_total_spent") {
			continue
		}
		filtered = append(filtered, g)
	}
	return filtered
}

func relPath(root, abs string) string {
	rel := strings.TrimPrefix(abs, root)
	return strings.TrimPrefix(rel, "/")
}

func (e *Extractor) generateFunction(sql string, client *llm.Client) string {
	prompt := fmt.Sprintf(`Generate a PostgreSQL function for this SQL.

SQL:
%s

Rules:
- Use LANGUAGE plpgsql
- Use RETURNS TABLE with appropriate types (BIGINT for int64, INTEGER for int32, NUMERIC(10,2) for money, TEXT for strings, TIMESTAMP for dates)
- Use parameter names starting with p_
- Use $N parameter style
- Return the CREATE OR REPLACE FUNCTION DDL only, no explanation
- If the SQL has $1, $2 etc, use parameters p_1, p_2 with the appropriate types
- If the SQL has no parameters, create a function with no parameters`, sql)

	result, err := client.Send(prompt)
	if err != nil {
		return fmt.Sprintf("-- error generating function: %v", err)
	}

	return strings.TrimSpace(result)
}
