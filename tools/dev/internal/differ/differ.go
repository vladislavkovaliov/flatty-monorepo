package differ

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	label := strings.TrimPrefix(filePath, "/")
	cmd := exec.Command("diff", "-u",
		"--label", fmt.Sprintf("a/%s", label),
		"--label", fmt.Sprintf("b/%s", label),
		origPath, modPath)
	out, err := cmd.Output()
	if err != nil {
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

	return string(out), nil
}
