package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
	"shawnma.com/pic_rename/pkg/index"
)

func main() {
	var dbPath, dir string
	var deleteFlag bool
	flag.StringVarP(&dbPath, "db", "q", ".", "path to the database file")
	flag.StringVarP(&dir, "dir", "s", ".", "directory to index")
	flag.BoolVarP(&deleteFlag, "delete", "d", false, "interactively delete duplicate folders")
	flag.Parse()

	if dbPath == "" || dir == "" {
		fmt.Printf("%s: required --db and --dir\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	duplicates, err := index.UpdateIndex(dbPath, dir)
	if err != nil {
		panic(err)
	}

	for _, dup := range duplicates {
		fmt.Println("\n=== Duplicate Found ===")
		fmt.Println("Original Location:")
		fmt.Printf("  Folder: %s\n", dup.OriginalFolder)
		fmt.Println("  Files:")
		for _, file := range dup.OriginalFiles {
			fmt.Printf("    - %s\n", file)
		}
		fmt.Println("\nNew Location:")
		fmt.Printf("  Folder: %s\n", dup.NewFolder)
		fmt.Println("  Files:")
		for _, file := range dup.NewFiles {
			fmt.Printf("    - %s\n", file)
		}
		fmt.Println("=====================")

		if deleteFlag {
			fmt.Println("\nOptions:")
			fmt.Println("o - Delete original files")
			fmt.Println("n - Delete new files")
			fmt.Println("s - Skip (default)")
			fmt.Print("Enter your choice (o/n/s): ")

			var choice string
			fmt.Scanf("%s", &choice)
			choice = strings.ToLower(strings.TrimSpace(choice))

			switch choice {
			case "o":
				if err := deleteFiles(dup.OriginalFiles); err != nil {
					fmt.Printf("Error deleting original files: %v\n", err)
				} else {
					fmt.Printf("Successfully deleted original files\n")
				}
			case "n":
				if err := deleteFiles(dup.NewFiles); err != nil {
					fmt.Printf("Error deleting new files: %v\n", err)
				} else {
					fmt.Printf("Successfully deleted new files\n")
				}
			default:
				fmt.Println("Skipping...")
			}
		}
	}
}

func deleteFiles(files []string) error {
	for _, file := range files {
		// Convert to absolute path
		absPath, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", file, err)
		}

		// Check if the file exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			fmt.Printf("Warning: file does not exist: %s\n", absPath)
			continue
		}

		// Remove the file
		if err := os.Remove(absPath); err != nil {
			return fmt.Errorf("failed to delete file %s: %w", absPath, err)
		}
	}
	return nil
}
