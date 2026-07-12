package git

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

type DiffResult struct {
	Files []string
	Diff  string
}

func GetStagedDiff(workDir, pathFilter string, exts ...string) (*DiffResult, error) {
	step := logStep(fmt.Sprintf("git: scanning staged changes in %s", pathFilter))
	defer step()

	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--", pathFilter)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --name-only: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	var files []string
	for _, f := range lines {
		if f == "" {
			continue
		}
		if len(exts) == 0 || hasSuffixAny(f, exts) {
			files = append(files, f)
		}
	}

	if len(files) == 0 {
		return &DiffResult{}, nil
	}

	log.Printf("  found %d file(s):", len(files))
	for _, f := range files {
		log.Printf("    • %s", f)
	}

	cmd2 := exec.Command("git", append([]string{"diff", "--cached", "-U10", "--"}, files...)...)
	cmd2.Dir = workDir
	diffOut, err := cmd2.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff content: %w", err)
	}

	diffSize := len(diffOut)
	log.Printf("  diff size: %d bytes (%d lines)", diffSize, bytes.Count(diffOut, []byte{'\n'}))

	return &DiffResult{
		Files: files,
		Diff:  string(diffOut),
	}, nil
}

func IsClean(workDir, pathFilter string) bool {
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--", pathFilter)
	cmd.Dir = workDir
	out, _ := cmd.Output()
	return len(bytes.TrimSpace(out)) == 0
}

func hasSuffixAny(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

func logStep(name string) func() {
	start := time.Now()
	log.Printf("━━━ %s ━━━", name)
	return func() {
		log.Printf("  ✔ done in %v", time.Since(start).Round(time.Millisecond))
	}
}
