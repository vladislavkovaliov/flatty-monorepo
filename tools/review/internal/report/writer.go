package report

import (
	"fmt"
	"strings"
	"time"
)

const separator = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

func PrintHeader(files []string) {
	fmt.Println()
	fmt.Println(separator)
	fmt.Println("  Code Review Report")
	fmt.Printf("  %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(separator)
	fmt.Println()

	fmt.Printf("  Changed files (%d):\n", len(files))
	for _, f := range files {
		fmt.Printf("    📄 %s\n", f)
	}
	fmt.Println()
}

func PrintReview(model string, review string) {
	fmt.Println("  ── Model: " + model + " ──")
	fmt.Println()
	fmt.Println(review)
	fmt.Println()
	fmt.Println(separator)
	fmt.Println()
}

func PrintNoChanges() {
	fmt.Println()
	fmt.Println(separator)
	fmt.Println("  No staged changes found")
	fmt.Println(separator)
	fmt.Println()
}

func PrintError(msg string) {
	fmt.Println()
	fmt.Printf("  ❌ Error: %s\n", msg)
	fmt.Println()
}

func FormatErrors(errors []string) string {
	if len(errors) == 0 {
		return ""
	}
	var b strings.Builder
	for _, e := range errors {
		b.WriteString(e)
		b.WriteString("\n")
	}
	return b.String()
}
