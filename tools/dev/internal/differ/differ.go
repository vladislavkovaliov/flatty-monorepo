package differ

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func UnifiedDiff(original, modified, filePath string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "dev-diff-*")
	if err != nil {
		return "", fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	origPath := filepath.Join(tmpDir, "orig")
	modPath := filepath.Join(tmpDir, "mod")

	if err := os.WriteFile(origPath, []byte(original), 0644); err != nil {
		return "", fmt.Errorf("write orig: %w", err)
	}
	if err := os.WriteFile(modPath, []byte(modified), 0644); err != nil {
		return "", fmt.Errorf("write mod: %w", err)
	}

	cmd := exec.Command("diff", "-u", origPath, modPath)
	out, err := cmd.Output()
	if err != nil {
		// diff exits 1 when files differ — that's expected
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 1 {
				return "", fmt.Errorf("diff failed: %w", err)
			}
		} else {
			return "", fmt.Errorf("diff failed: %w", err)
		}
	}

	if len(out) == 0 {
		return "", nil
	}

	diff := string(out)
	diff = replacePath(diff, "orig", filePath)
	diff = replacePath(diff, "mod", filePath)

	return diff, nil
}

func replacePath(diff, old, new string) string {
	oldA := fmt.Sprintf("--- %s", old)
	newA := fmt.Sprintf("+++ %s", old)
	oldB := fmt.Sprintf("--- a/%s", old)
	newB := fmt.Sprintf("+++ b/%s", old)

	result := diff
	result = replaceLine(result, oldA, fmt.Sprintf("--- a/%s", new))
	result = replaceLine(result, newA, fmt.Sprintf("+++ b/%s", new))
	result = replaceLine(result, oldB, fmt.Sprintf("--- a/%s", new))
	result = replaceLine(result, newB, fmt.Sprintf("+++ b/%s", new))

	return result
}

func replaceLine(s, old, new string) string {
	lines := []byte(s)
	oldB := []byte(old + "\n")
	newB := []byte(new + "\n")
	result := make([]byte, 0, len(s))

	i := 0
	for i <= len(lines) {
		if i+len(oldB) <= len(lines) && string(lines[i:i+len(oldB)]) == string(oldB) {
			result = append(result, newB...)
			i += len(oldB)
		} else {
			if i < len(lines) {
				result = append(result, lines[i])
			}
			i++
		}
	}

	return string(result)
}
