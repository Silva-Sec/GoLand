package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/sqweek/dialog"
)

// Result struct to hold search results
type Result struct {
	filename   string
	lineNumber int
	line       string
}

func searchWordsInFile(filename string, word string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

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
			results <- Result{filename, lineNumber, line}
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
	}
}

func searchWordsInDirectory(dirname string, word string, resultsFile *os.File, progress *widget.ProgressBar, resultCount *widget.Label, output *widget.Entry) {
	var wg sync.WaitGroup
	results := make(chan Result)

	// Collect all files
	var files []string
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".txt") || strings.HasSuffix(info.Name(), ".log")) {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through directory %s: %v\n", dirname, err)
		return
	}

	totalFiles := len(files)
	progress.SetValue(0)
	resultCount.SetText("Results found: 0")
	output.SetText("")

	// Start a goroutine to search each file
	for _, file := range files {
		wg.Add(1)
		go searchWordsInFile(file, word, results, &wg)
	}

	// Close the results channel once all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results
	count := 0
	processedFiles := 0
	for result := range results {
		resultString := fmt.Sprintf("Found '%s' in file %s at line %d: %s\n", word, result.filename, result.lineNumber, result.line)
		fmt.Print(resultString)
		resultsFile.WriteString(resultString)
		output.SetText(output.Text + resultString)
		count++
		resultCount.SetText(fmt.Sprintf("Results found: %d", count))

		// Update progress bar
		processedFiles++
		progress.SetValue(float64(processedFiles) / float64(totalFiles))
	}
}

func openLogFile(logFileName string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", logFileName)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", logFileName)
	case "darwin":
		cmd = exec.Command("open", logFileName)
	default:
		fmt.Printf("Unsupported platform")
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
	}
}

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.LightTheme())
	myWindow := myApp.NewWindow("Log Analyzer")

	// Create input fields
	dirEntry := widget.NewEntry()
	dirEntry.SetPlaceHolder("Enter directory to search")
	wordEntry := widget.NewEntry()
	wordEntry.SetPlaceHolder("Enter word to search for")

	// Create progress bar and result count label
	progress := widget.NewProgressBar()
	resultCount := widget.NewLabel("Results found: 0")

	// Create output box
	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("Search output will be displayed here...")
	output.Wrapping = fyne.TextWrapWord
	output.SetMinSize(fyne.NewSize(200, 200))
	outputBox := container.NewScroll(output)
	outputBox.SetMinSize(fyne.NewSize(400, 200))

	var resultsFileName string

	// Create directory selection button
	selectDirButton := widget.NewButton("Select Directory", func() {
		dir, err := dialog.Directory().Title("Select Directory").Browse()
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		if dir != "" {
			dirEntry.SetText(dir)
		}
	})

	// Create search button
	searchButton := widget.NewButton("Search", func() {
		directory := dirEntry.Text
		word := wordEntry.Text

		if directory == "" {
			dialog.ShowInformation("Error", "Please select a directory to search.", myWindow)
			return
		}

		if word == "" {
			dialog.ShowInformation("Error", "Please enter a word to search for.", myWindow)
			return
		}

		// Get the current working directory
		cwd, err := os.Getwd()
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error getting current working directory: %v", err), myWindow)
			return
		}

		// Get current time for the results file name
		currentTime := time.Now()
		resultsFileName = filepath.Join(cwd, fmt.Sprintf("results_%s.txt", currentTime.Format("20060102_150405")))

		// Create or open the results file
		resultsFile, err := os.Create(resultsFileName)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error creating results file: %v", err), myWindow)
			return
		}
		defer resultsFile.Close()

		// Search for the word in the directory
		go searchWordsInDirectory(directory, word, resultsFile, progress, resultCount, output)

		dialog.ShowInformation("Search Started", "The search has started. Please wait...", myWindow)
	})

	// Create open log button
	openLogButton := widget.NewButton("Open Log File", func() {
		if resultsFileName != "" {
			openLogFile(resultsFileName)
		} else {
			dialog.ShowInformation("No Log File", "No log file to open. Please perform a search first.", myWindow)
		}
	})

	// Create layout and add widgets
	content := container.NewVBox(
		container.NewHBox(dirEntry, selectDirButton),
		wordEntry,
		searchButton,
		progress,
		outputBox,
		resultCount,
		openLogButton,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(600, 600))
	myWindow.ShowAndRun()
}
