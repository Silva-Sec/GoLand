package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SearchWordsInFile searches for specific words in a single file and prints the results.
func SearchWordsInFile(filename string, words []string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 1
	for scanner.Scan() {
		line := scanner.Text()
		for _, word := range words {
			if strings.Contains(line, word) {
				fmt.Printf("Found '%s' in file %s at line %d\n", word, filename, lineNumber)
			}
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
	}
}

// SearchWordsInDirectory searches for specific words in all text files within a directory.
func SearchWordsInDirectory(dirname string, words []string) {
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".txt") {
			SearchWordsInFile(path, words)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through directory %s: %v\n", dirname, err)
	}
}

func main() {
	// Define the directory to search and the words to look for
	directory := "./texts" // Change this to your directory
	words := []string{"word1", "word2", "word3"} // Replace with your words

	// Search for the words in the directory
	SearchWordsInDirectory(directory, words)
}
