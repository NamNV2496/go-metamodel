package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/namnv2496/metamodel/generator"
)

var (
	source      = flag.String("source", "./...", "Source file, directory, or pattern (e.g., models.go, ., ./..., ./pkg/...)")
	destination = flag.String("destination", "metamodel", "Output file or directory (default: alongside each source as <name>_metamodel.go)")
	packageName = flag.String("packageName", "metamodel", "Package name for generated file (default: metamodel, optional for custome with pattern <packageName>, e.g., models)")
	tag         = flag.String("tag", "json", "Specific tag name to generate (optional, e.g., json, bson, gorm)")
	tableName   = flag.String("tableName", "", "Specific table name to generate (default: <structName>s, optional for custome e.g., users, mock_test)")
)

func main() {
	flag.Parse()

	if *source == "" {
		fmt.Fprintln(os.Stderr, "Error: -source flag is required")
		flag.Usage()
		os.Exit(1)
	}

	sources, err := resolveSourceFiles(*source)
	if err != nil {
		log.Fatalf("Error resolving source: %v", err)
	}
	if len(sources) == 0 {
		log.Fatalf("No .go source files found for: %s", *source)
	}

	dest := *destination
	// When scanning multiple files, treat destination as a directory
	if len(sources) > 1 && dest != "" && !strings.HasSuffix(dest, "/") && !strings.HasSuffix(dest, string(filepath.Separator)) {
		dest = dest + string(filepath.Separator)
	}

	errCount := 0
	for _, src := range sources {
		cfg := generator.Config{
			Source:      src,
			Destination: dest,
			PackageName: *packageName,
			Tag:         *tag,
			TableName:   *tableName,
		}
		if err := generator.Generate(cfg); err != nil {
			log.Printf("Error generating metamodel for %s: %v", src, err)
			errCount++
			continue
		}
		fmt.Printf("Successfully generated metamodel for %s\n", src)
	}

	if errCount > 0 {
		os.Exit(1)
	}
}

func resolveSourceFiles(src string) ([]string, error) {
	// Handle ./... or path/to/pkg/... recursive patterns
	if src == "./..." || strings.HasSuffix(src, "/...") {
		baseDir := strings.TrimSuffix(src, "/...")
		if baseDir == "." || baseDir == "" {
			baseDir = "."
		}
		return walkGoFiles(baseDir, true)
	}

	// Handle plain directory
	info, err := os.Stat(src)
	if err == nil && info.IsDir() {
		return walkGoFiles(src, false)
	}

	// Single file
	return []string{src}, nil
}

func walkGoFiles(root string, recursive bool) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path == root {
				return nil
			}
			if !recursive {
				return filepath.SkipDir
			}
			// Skip vendor and hidden directories
			base := info.Name()
			if base == "vendor" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Skip test files and already-generated metamodel files
		if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_metamodel.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}
