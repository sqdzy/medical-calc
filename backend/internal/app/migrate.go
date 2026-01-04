package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations runs database migrations.
// migrationsPath must include scheme, e.g. file://migrations
func RunMigrations(databaseURL, migrationsPath string) error {
	resolved, err := resolveMigrationsPath(migrationsPath)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	m, err := migrate.New(resolved, databaseURL)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	srcErr, dbErr := m.Close()
	if srcErr != nil {
		return fmt.Errorf("close migrate source: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("close migrate db: %w", dbErr)
	}

	return nil
}

func resolveMigrationsPath(migrationsPath string) (string, error) {
	// Default used by the repo. Keep it explicit and scheme-based.
	if strings.TrimSpace(migrationsPath) == "" {
		migrationsPath = "file://./migrations"
	}

	// If a plain filesystem path is provided, treat it as relative/absolute dir.
	if !strings.Contains(migrationsPath, "://") {
		return fileURLForDir(migrationsPath)
	}

	// Only normalize local file source. Other sources (s3, github, etc.) are passed through.
	if !strings.HasPrefix(migrationsPath, "file://") {
		return migrationsPath, nil
	}

	pathPart := strings.TrimPrefix(migrationsPath, "file://")
	pathPart = strings.TrimSpace(pathPart)
	if pathPart == "" {
		pathPart = "./migrations"
	}

	// If absolute (Unix-like or Windows drive), just normalize to file URL.
	if filepath.IsAbs(pathPart) || hasWindowsDrive(pathPart) {
		return fileURLForDir(pathPart)
	}

	// Try resolving relative paths against:
	// 1) current working directory
	// 2) executable directory
	// 3) parent of executable directory (common when running build/api.exe)
	candidates := []string{pathPart}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, pathPart))
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates, filepath.Join(exeDir, pathPart))
		candidates = append(candidates, filepath.Join(filepath.Dir(exeDir), pathPart))
	}

	for _, candidate := range candidates {
		if dirExists(candidate) {
			return fileURLForDir(candidate)
		}
	}

	return migrationsPath, nil
}

func fileURLForDir(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if !dirExists(abs) {
		// Still return a normalized URL so migrate reports a meaningful error.
		return "file://" + filepath.ToSlash(abs), nil
	}
	return "file://" + filepath.ToSlash(abs), nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func hasWindowsDrive(p string) bool {
	// Detect patterns like C:/... or C:\...
	if len(p) < 2 {
		return false
	}
	return p[1] == ':'
}
