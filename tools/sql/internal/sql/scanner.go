package sql

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type SQLBlock struct {
	FilePath string
	Line     int
	RawSQL   string
}

var sqlRe = regexp.MustCompile("`([^`]*?(?:SELECT|INSERT|UPDATE|DELETE|WITH)[^`]*?)`")

func ScanGoRepo(root string) ([]SQLBlock, error) {
	dir := filepath.Join(root, "shop-api-golang", "repository")
	return scanDir(dir, ".go")
}

func ScanTSRepo(root string) ([]SQLBlock, error) {
	dir := filepath.Join(root, "shop-graphql-nestjs", "src")
	return scanDir(dir, ".repository.ts")
}

func scanDir(dir, ext string) ([]SQLBlock, error) {
	var blocks []SQLBlock

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ext) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		content := string(data)
		matches := sqlRe.FindAllStringSubmatch(content, -1)
		if matches == nil {
			return nil
		}

		locs := sqlRe.FindAllStringSubmatchIndex(content, -1)

		for i, m := range matches {
			sql := strings.TrimSpace(m[1])
			if sql == "" {
				continue
			}
			lineNo := bytesToLine(content, locs[i][0])
			blocks = append(blocks, SQLBlock{
				FilePath: path,
				Line:     lineNo,
				RawSQL:   sql,
			})
		}
		return nil
	})

	return blocks, err
}

func bytesToLine(content string, bytePos int) int {
	if bytePos <= 0 {
		return 1
	}
	return strings.Count(content[:bytePos], "\n") + 1
}
