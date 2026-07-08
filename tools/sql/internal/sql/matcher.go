package sql

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

var whitespaceRe = regexp.MustCompile(`\s+`)
var paramRe = regexp.MustCompile(`\$(\d+)`)
var commentRe = regexp.MustCompile(`(?s)/\*.*?\*/|--[^\n]*`)

type MatchGroup struct {
	Hash    string
	Blocks  []SQLBlock
	GoBlock *SQLBlock
	TSBlock *SQLBlock
}

func Normalize(sql string) string {
	sql = commentRe.ReplaceAllString(sql, "")
	sql = strings.ToLower(sql)
	sql = whitespaceRe.ReplaceAllString(sql, " ")
	sql = paramRe.ReplaceAllString(sql, "?")
	sql = strings.TrimSpace(sql)
	return sql
}

func Hash(sql string) string {
	h := sha256.Sum256([]byte(Normalize(sql)))
	return fmt.Sprintf("%x", h[:8])
}

func GroupBlocks(goBlocks, tsBlocks []SQLBlock) []MatchGroup {
	index := make(map[string]*MatchGroup)

	addToIndex := func(b SQLBlock) {
		h := Hash(b.RawSQL)
		if g, ok := index[h]; ok {
			g.Blocks = append(g.Blocks, b)
		} else {
			index[h] = &MatchGroup{
				Hash:   h,
				Blocks: []SQLBlock{b},
			}
		}
	}

	for _, b := range goBlocks {
		addToIndex(b)
	}
	for _, b := range tsBlocks {
		addToIndex(b)
	}

	var groups []MatchGroup
	for _, g := range index {
		hasGo := false
		hasTS := false
		for _, b := range g.Blocks {
			if strings.HasSuffix(b.FilePath, ".go") {
				hasGo = true
				g.GoBlock = &b
			} else {
				hasTS = true
				g.TSBlock = &b
			}
		}
		if hasGo && hasTS {
			groups = append(groups, *g)
		}
	}

	return groups
}
