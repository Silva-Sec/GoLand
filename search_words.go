package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// SearchWordsInFile searches for a specific word in a single file and writes the results to a file.
func SearchWordsInFile(filename string, word string, results *os.File) {
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
		if strings.Contains(line, word) {
			result := fmt.Sprintf("Found '%s' in file %s at line %d: %s\n", word, filename, lineNumber, line)
			fmt.Print(result)
			results.WriteString(result)
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
	}
}

// SearchWordsInDirectory searches for a specific word in all .txt and .log files within a directory and its subfolders.
func SearchWordsInDirectory(dirname string, word string, results *os.File) {
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".txt") || strings.HasSuffix(info.Name(), ".log")) {
			SearchWordsInFile(path, word, results)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through directory %s: %v\n", dirname, err)
	}
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Word Search")

	// Create input fields
	dirEntry := widget.NewEntry()
	dirEntry.SetPlaceHolder("Enter directory to search")
	wordEntry := widget.NewEntry()
	wordEntry.SetPlaceHolder("Enter word to search for")

	// Create button
	searchButton := widget.NewButton("Search", func() {
		directory := dirEntry.Text
		word := wordEntry.Text

		// Get current time for the results file name
		currentTime := time.Now()
		resultsFileName := fmt.Sprintf("results_%s.txt", currentTime.Format("20060102_150405"))

		// Create or open the results file
		resultsFile, err := os.Create(resultsFileName)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error creating results file: %v", err), myWindow)
			return
		}
		defer resultsFile.Close()

		// Search for the word in the directory
		SearchWordsInDirectory(directory, word, resultsFile)

		dialog.ShowInformation("Search Completed", fmt.Sprintf("Results saved in %s", resultsFileName), myWindow)
	})

	// Create layout and add widgets
	content := container.NewVBox(
		dirEntry,
		wordEntry,
		searchButton,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(400, 200))
	myWindow.ShowAndRun()
}
