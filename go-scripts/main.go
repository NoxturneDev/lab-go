package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	rootDir := "/mnt/c/Users/galih/Documents/tmp"
	destDir := "/mnt/c/Users/galih/Documents/archive-files"

	// Ensure destination folder exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Printf("Failed to create archive folder: %v\n", err)
		return
	}

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the archive folder itself to avoid recursion
		if d.IsDir() && path == destDir {
			return filepath.SkipDir
		}

		if !d.IsDir() && strings.Contains(d.Name(), "uas") {
			newPath := filepath.Join(destDir, d.Name())
			fmt.Printf("Moving: %s → %s\n", path, newPath)
			if err := os.Rename(path, newPath); err != nil {
				fmt.Printf("  Failed to move: %v\n", err)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
	}
}
