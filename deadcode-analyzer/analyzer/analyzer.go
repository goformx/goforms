package analyzer

import (
	"fmt"
	"log"
	"sort"
)

// Analyzer is the main analysis engine
type Analyzer struct {
	config     *Config
	parser     *Parser
	safety     *SafetyAnalyzer
	callGraph  *CallGraphAnalyzer
	fxDetector *FxDetector
	imports    *ImportAnalyzer
	templates  *TemplateDetector
	tests      *TestAnalyzer
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		config: &Config{
			ProjectRoot: ".",
		},
		parser:     NewParser(),
		safety:     NewSafetyAnalyzer(),
		callGraph:  NewCallGraphAnalyzer(),
		fxDetector: NewFxDetector(),
		imports:    NewImportAnalyzer(),
		templates:  NewTemplateDetector(),
		tests:      NewTestAnalyzer(),
	}
}

// SetVerbose enables verbose output
func (a *Analyzer) SetVerbose(verbose bool) {
	a.config.Verbose = verbose
}

// SetOutputFile sets the output file for detailed results
func (a *Analyzer) SetOutputFile(outputFile string) {
	a.config.OutputFile = outputFile
}

// SetJSONOutput enables JSON output format
func (a *Analyzer) SetJSONOutput(json bool) {
	a.config.JSONOutput = json
}

// SetProjectRoot sets the project root directory
func (a *Analyzer) SetProjectRoot(root string) {
	a.config.ProjectRoot = root
}

// Run executes the complete analysis
func (a *Analyzer) Run() error {
	if a.config.Verbose {
		log.Println("🔍 Loading GoForms project...")
	}

	// Load and parse packages
	if err := a.parser.LoadProject(a.config.ProjectRoot); err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	if a.config.Verbose {
		log.Println("🔍 Building call graph...")
	}

	// Build call graph
	if err := a.callGraph.Build(a.parser.GetPackages()); err != nil {
		log.Printf("Warning: Failed to build call graph: %v", err)
	}

	if a.config.Verbose {
		log.Println("🔍 Building import graph...")
	}

	// Build import graph
	if err := a.imports.BuildGraph(a.parser.GetPackages()); err != nil {
		log.Printf("Warning: Failed to build import graph: %v", err)
	}

	if a.config.Verbose {
		log.Println("🔍 Analyzing files...")
	}

	// Analyze all files
	results, err := a.analyzeAllFiles()
	if err != nil {
		return fmt.Errorf("failed to analyze files: %w", err)
	}

	// Print results
	a.printResults(results)

	return nil
}

// analyzeAllFiles analyzes all Go files in the project
func (a *Analyzer) analyzeAllFiles() (*Results, error) {
	// Get all Go files
	files, err := a.parser.GetGoFiles(InternalDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get Go files: %w", err)
	}

	// Pre-allocate slice with known size
	fileAnalyses := make([]*FileAnalysis, 0, len(files))

	// Analyze each file
	for _, filePath := range files {
		analysis, analysisErr := a.analyzeFile(filePath)
		if analysisErr != nil {
			if a.config.Verbose {
				log.Printf("Error analyzing %s: %v", filePath, analysisErr)
			}

			continue // Continue with other files
		}

		fileAnalyses = append(fileAnalyses, analysis)
	}

	// Sort by safety score (lower is safer)
	sort.Slice(fileAnalyses, func(i, j int) bool {
		return fileAnalyses[i].SafetyScore < fileAnalyses[j].SafetyScore
	})

	// Calculate statistics
	results := &Results{
		Files:      fileAnalyses,
		TotalFiles: len(fileAnalyses),
		Summary:    make(map[string]interface{}),
	}

	for _, analysis := range fileAnalyses {
		switch analysis.SafetyLevel {
		case UltraSafe:
			results.UltraSafe++
		case PotentiallySafe:
			results.PotentiallySafe++
		case Dangerous:
			results.Dangerous++
		case NeverDelete:
			results.NeverDelete++
		}
	}

	return results, nil
}

// analyzeFile performs comprehensive analysis of a single file
func (a *Analyzer) analyzeFile(filePath string) (*FileAnalysis, error) {
	analysis := &FileAnalysis{
		Path:        filePath,
		Reasons:     []string{},
		SafetyLevel: UltraSafe,
	}

	// Parse the file
	file, err := a.parser.ParseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	// Basic file analysis
	a.parser.AnalyzeFile(file, analysis)

	// Detect patterns
	a.fxDetector.Analyze(file, analysis)
	a.templates.Analyze(file, analysis)
	a.tests.Analyze(file, analysis)

	// Check reachability
	a.callGraph.AnalyzeReachability(filePath, analysis)

	// Check imports/exports
	a.imports.Analyze(filePath, analysis)

	// Calculate safety level and score
	a.safety.CalculateSafetyLevel(analysis)

	return analysis, nil
}

// printResults outputs the analysis results
func (a *Analyzer) printResults(results *Results) {
	fmt.Println("🔍 GoForms Dead Code Analysis Results")
	fmt.Println("=====================================")
	fmt.Println()

	// Print summary
	fmt.Printf("📊 Summary:\n")
	fmt.Printf("  Total files analyzed: %d\n", results.TotalFiles)
	fmt.Printf("  Ultra-safe candidates: %d\n", results.UltraSafe)
	fmt.Printf("  Potentially safe: %d\n", results.PotentiallySafe)
	fmt.Printf("  Dangerous: %d\n", results.Dangerous)
	fmt.Printf("  Never delete: %d\n", results.NeverDelete)
	fmt.Println()

	// Print ultra-safe candidates
	a.printUltraSafeCandidates(results)

	// Print dangerous files
	a.printDangerousFiles(results)
}

// printUltraSafeCandidates prints the ultra-safe deletion candidates
func (a *Analyzer) printUltraSafeCandidates(results *Results) {
	fmt.Println("🟢 ULTRA-SAFE Candidates (Safe to delete):")
	fmt.Println("-------------------------------------------")

	ultraSafeFound := false

	for _, result := range results.Files {
		if result.SafetyLevel != UltraSafe {
			continue
		}

		ultraSafeFound = true

		fmt.Printf("  %s\n", result.Path)
		fmt.Printf("    Functions: %d/%d unreachable\n", result.UnreachableFunctions, result.TotalFunctions)
		fmt.Printf("    Lines: %d\n", result.TotalLines)
		fmt.Printf("    Safety Score: %d\n", result.SafetyScore)

		if len(result.Reasons) > 0 {
			fmt.Printf("    Reasons: %s\n", result.Reasons)
		}

		fmt.Println()
	}

	if !ultraSafeFound {
		fmt.Println("  No ultra-safe candidates found (this is good!)")
	}

	fmt.Println()
}

// printDangerousFiles prints the dangerous files that should not be deleted
func (a *Analyzer) printDangerousFiles(results *Results) {
	fmt.Println("🔴 DANGEROUS Files (DO NOT DELETE):")
	fmt.Println("-----------------------------------")

	for _, result := range results.Files {
		if result.SafetyLevel == Dangerous || result.SafetyLevel == NeverDelete {
			fmt.Printf("  %s (%s)\n", result.Path, result.SafetyLevel)

			if len(result.Reasons) > 0 {
				fmt.Printf("    Reasons: %s\n", result.Reasons)
			}
		}
	}
}
