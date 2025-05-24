package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"example.com/charNorm" // Import the local charNorm library
)

const normalizationTableURL = "https://raw.githubusercontent.com/zy9ard3/zy9ard3.github.io/refs/heads/main/normalizations.html"

func main() {
	// 2b. Call charNorm.ParseNormalizationTable
	// fmt.Fprintln(os.Stderr, "Fetching and parsing normalization table...") // Debug message
	normMap, err := charNorm.ParseNormalizationTable(normalizationTableURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing normalization table: %v\n", err)
		os.Exit(1)
	}
	// fmt.Fprintln(os.Stderr, "Normalization table parsed successfully.") // Debug message

	// 2c. Determine the input payload
	var inputPayload string
	if len(os.Args) > 1 {
		// 2c.i. Command-line arguments provided
		inputPayload = strings.Join(os.Args[1:], " ")
		// fmt.Fprintf(os.Stderr, "Using input from command-line arguments: \"%s\"\n", inputPayload) // Debug message
	} else {
		// 2c.ii. No command-line arguments, read from stdin
		// fmt.Fprintln(os.Stderr, "No command-line arguments provided. Reading from stdin. Press Ctrl+D (or Ctrl+Z on Windows) to end input.") // User guidance
		reader := bufio.NewReader(os.Stdin)
		var sb strings.Builder
		// Read line by line as Stdin might not be closed by the user immediately
		for {
			line, err := reader.ReadString('\n')
			sb.WriteString(line)
			if err == io.EOF {
				break // End of input
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
		}
		inputPayload = strings.TrimSpace(sb.String())
		// fmt.Fprintf(os.Stderr, "Using input from stdin: \"%s\"\n", inputPayload) // Debug message
	}

	if inputPayload == "" && len(os.Args) <=1 { // only if no args AND stdin was empty
		// fmt.Fprintln(os.Stderr, "No input provided via arguments or stdin. Nothing to process.") // Info message
		// As GenerateVariations handles empty string to return {""}, let it proceed.
		// If an error is desired for no input, exit here:
		// os.Exit(0) // or 1 depending on desired behavior
	}


	// 2d. Call charNorm.GenerateVariations
	variations := charNorm.GenerateVariations(inputPayload, normMap)

	// 2e. Print each variation
	for _, variation := range variations {
		fmt.Println(variation)
	}
}
