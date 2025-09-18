package scan

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ResolveDirectory normalises a filesystem path and ensures it refers to a directory.
func ResolveDirectory(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("target path is empty")
	}

	candidate := filepath.Clean(path)
	if !filepath.IsAbs(candidate) {
		if abs, err := filepath.Abs(candidate); err == nil {
			candidate = abs
		}
	}

	info, err := os.Stat(candidate)
	if err != nil {
		return "", fmt.Errorf("target path %q not accessible: %w", candidate, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("target path %q is not a directory", candidate)
	}

	return candidate, nil
}

// RunSemgrep executes Semgrep with the provided configuration.
func RunSemgrep(parent context.Context, targetPath, config string, timeout time.Duration) ([]byte, error) {
	if config == "" {
		config = "auto"
	}

	ctx, cancel := withOptionalTimeout(parent, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "semgrep", "scan", "--config", config, "--json", "--quiet", targetPath)
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return output, fmt.Errorf("semgrep timed out after %s", timeout)
	}
	if err != nil {
		return output, fmt.Errorf("semgrep failed: %w\n%s", err, output)
	}
	return output, nil
}

// RunGrype executes Grype against the supplied directory.
func RunGrype(parent context.Context, targetPath string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := withOptionalTimeout(parent, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "grype", "--output", "json", "dir:"+targetPath)
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return output, fmt.Errorf("grype timed out after %s", timeout)
	}
	if err != nil {
		return output, fmt.Errorf("grype failed: %w\n%s", err, output)
	}
	return output, nil
}

func withOptionalTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return parent, func() {}
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	return ctx, cancel
}
