package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileContent struct {
	Path    string
	Content string
}

func ReadFiles(repoRoot string, files []string) ([]FileContent, error) {
	var result []FileContent

	for _, f := range files {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}

		absPath := f
		if !filepath.IsAbs(f) {
			absPath = filepath.Join(repoRoot, f)
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", f, err)
		}

		result = append(result, FileContent{
			Path:    f,
			Content: string(data),
		})
	}

	return result, nil
}
