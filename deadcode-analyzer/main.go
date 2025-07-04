package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/goformx/goforms/deadcode-analyzer/analyzer"
)

func main() {
	var (
		outputFile = flag.String("output", "", "Output file for detailed results (optional)")
		verbose    = flag.Bool("verbose", false, "Enable verbose output")
		json       = flag.Bool("json", false, "Output results in JSON format")
	)
	flag.Parse()

	// Create analyzer instance
	analyzerInstance := analyzer.NewAnalyzer()

	// Configure analyzer
	analyzerInstance.SetVerbose(*verbose)
	analyzerInstance.SetOutputFile(*outputFile)
	analyzerInstance.SetJSONOutput(*json)

	// Run analysis
	fmt.Println("🔍 GoForms Dead Code Analyzer")
	fmt.Println("=============================")
	fmt.Println()

	if err := analyzerInstance.Run(); err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}
}
