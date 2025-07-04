package analyzer

import (
	"go/ast"
	"strings"
)

// Critical packages that should never be deleted
var criticalPackages = map[string]bool{
	"sanitization":       true,
	"logging":            true,
	"config":             true,
	"database":           true,
	"interfaces":         true,
	"errors":             true,
	"events":             true,
	"middleware/access":  true,
	"middleware/session": true,
	"middleware/auth":    true,
	"validation":         true,
	"response":           true,
	"module":             true,
}

// SafetyAnalyzer handles safety classification and scoring
type SafetyAnalyzer struct{}

// NewSafetyAnalyzer creates a new safety analyzer
func NewSafetyAnalyzer() *SafetyAnalyzer {
	return &SafetyAnalyzer{}
}

// CalculateSafetyLevel determines the safety level and score for a file
func (sa *SafetyAnalyzer) CalculateSafetyLevel(analysis *FileAnalysis) {
	// Start with ultra-safe
	analysis.SafetyLevel = UltraSafe
	analysis.SafetyScore = 0

	// Apply all penalties and bonuses
	sa.applyCriticalPackagePenalty(analysis)
	sa.applyInterfacePenalty(analysis)
	sa.applyFxUsagePenalty(analysis)
	sa.applyImportPenalty(analysis)
	sa.applyTestPenalty(analysis)
	sa.applyTemplatePenalty(analysis)
	sa.applySizePenalty(analysis)
	sa.applyUnreachableBonus(analysis)
	sa.applyExportPenalty(analysis)
	sa.applyFunctionTypePenalty(analysis)
	sa.applyStateModificationPenalty(analysis)
	sa.applyPureUtilityBonus(analysis)

	// Determine final safety level based on score
	sa.determineFinalSafetyLevel(analysis)
}

// applyCriticalPackagePenalty applies penalty for critical packages
func (sa *SafetyAnalyzer) applyCriticalPackagePenalty(analysis *FileAnalysis) {
	if sa.isCriticalPackage(analysis.Path) {
		analysis.SafetyLevel = NeverDelete
		analysis.SafetyScore += 100000
		analysis.Reasons = append(analysis.Reasons, "Critical package - NEVER DELETE")
	}
}

// applyInterfacePenalty applies penalty for files with interfaces
func (sa *SafetyAnalyzer) applyInterfacePenalty(analysis *FileAnalysis) {
	if analysis.HasInterfaces {
		analysis.SafetyLevel = Dangerous
		analysis.SafetyScore += 50000
		analysis.Reasons = append(analysis.Reasons, "Contains interfaces - DANGEROUS")
	}
}

// applyFxUsagePenalty applies penalty for Fx DI usage
func (sa *SafetyAnalyzer) applyFxUsagePenalty(analysis *FileAnalysis) {
	if analysis.HasFxUsage {
		analysis.SafetyLevel = Dangerous
		analysis.SafetyScore += 20000
		analysis.Reasons = append(analysis.Reasons, "Fx DI usage - DANGEROUS")
	}
}

// applyImportPenalty applies penalty for imported files
func (sa *SafetyAnalyzer) applyImportPenalty(analysis *FileAnalysis) {
	if analysis.IsImported {
		analysis.SafetyLevel = Dangerous
		analysis.SafetyScore += 10000
		analysis.Reasons = append(analysis.Reasons, "Imported elsewhere - DANGEROUS")
	}
}

// applyTestPenalty applies penalty for files with tests
func (sa *SafetyAnalyzer) applyTestPenalty(analysis *FileAnalysis) {
	if analysis.HasTests {
		analysis.SafetyLevel = Dangerous
		analysis.SafetyScore += 5000
		analysis.Reasons = append(analysis.Reasons, "Has associated tests - DANGEROUS")
	}
}

// applyTemplatePenalty applies penalty for files with templates
func (sa *SafetyAnalyzer) applyTemplatePenalty(analysis *FileAnalysis) {
	if analysis.HasTemplates {
		analysis.SafetyLevel = Dangerous
		analysis.SafetyScore += 3000
		analysis.Reasons = append(analysis.Reasons, "Contains template usage - DANGEROUS")
	}
}

// applySizePenalty applies penalty based on file size
func (sa *SafetyAnalyzer) applySizePenalty(analysis *FileAnalysis) {
	analysis.SafetyScore += sa.calculateSizePenalty(analysis.TotalLines)
}

// applyUnreachableBonus applies bonus for unreachable functions
func (sa *SafetyAnalyzer) applyUnreachableBonus(analysis *FileAnalysis) {
	if analysis.TotalFunctions > 0 {
		unreachablePct := (analysis.UnreachableFunctions * 100) / analysis.TotalFunctions
		analysis.SafetyScore -= unreachablePct * 300
	}
}

// applyExportPenalty applies penalty for exported functions
func (sa *SafetyAnalyzer) applyExportPenalty(analysis *FileAnalysis) {
	if len(analysis.ExportedFunctions) > 0 {
		analysis.SafetyLevel = Dangerous
		analysis.SafetyScore += len(analysis.ExportedFunctions) * 1500
		analysis.Reasons = append(analysis.Reasons, "Has exported functions - DANGEROUS")
	}
}

// applyFunctionTypePenalty applies penalty based on function types
func (sa *SafetyAnalyzer) applyFunctionTypePenalty(analysis *FileAnalysis) {
	analysis.SafetyScore += sa.calculateFunctionTypePenalty(analysis)
}

// applyStateModificationPenalty applies penalty for state modification
func (sa *SafetyAnalyzer) applyStateModificationPenalty(analysis *FileAnalysis) {
	if sa.hasStateModification(analysis) {
		analysis.SafetyLevel = Dangerous
		analysis.SafetyScore += 8000
		analysis.Reasons = append(analysis.Reasons, "Modifies application state - DANGEROUS")
	}
}

// applyPureUtilityBonus applies bonus for pure utility functions
func (sa *SafetyAnalyzer) applyPureUtilityBonus(analysis *FileAnalysis) {
	if sa.isPureUtility(analysis) {
		analysis.SafetyScore -= 2000
		analysis.Reasons = append(analysis.Reasons, "Pure utility functions - SAFER")
	}
}

// determineFinalSafetyLevel determines the final safety level based on score
func (sa *SafetyAnalyzer) determineFinalSafetyLevel(analysis *FileAnalysis) {
	if analysis.SafetyScore < 1000 &&
		analysis.UnreachableFunctions == analysis.TotalFunctions &&
		analysis.TotalFunctions > 0 {
		analysis.SafetyLevel = UltraSafe
	} else if analysis.SafetyScore < 5000 {
		analysis.SafetyLevel = PotentiallySafe
	} else {
		analysis.SafetyLevel = Dangerous
	}
}

// calculateSizePenalty calculates a more nuanced size penalty
func (sa *SafetyAnalyzer) calculateSizePenalty(lines int) int {
	// Smaller files get lower penalties
	if lines <= 10 {
		return lines * 1
	} else if lines <= 50 {
		return lines * 2
	} else if lines <= 100 {
		return lines * 3
	} else {
		return lines * 4 // Higher penalty for very large files
	}
}

// calculateFunctionTypePenalty calculates penalties based on function types
func (sa *SafetyAnalyzer) calculateFunctionTypePenalty(analysis *FileAnalysis) int {
	penalty := 0

	// Check for init functions
	if sa.hasInitFunctions(analysis) {
		penalty += 5000
		analysis.Reasons = append(analysis.Reasons, "Contains init() functions - DANGEROUS")
	}

	// Check for main function
	if sa.hasMainFunction(analysis) {
		penalty += 10000
		analysis.Reasons = append(analysis.Reasons, "Contains main() function - DANGEROUS")
	}

	// Check for global variables
	if sa.hasGlobalVariables(analysis) {
		penalty += 3000
		analysis.Reasons = append(analysis.Reasons, "Contains global variables - DANGEROUS")
	}

	return penalty
}

// hasStateModification checks if the file modifies application state
func (sa *SafetyAnalyzer) hasStateModification(analysis *FileAnalysis) bool {
	// This would require parsing the file to check for:
	// - Database operations
	// - File system operations
	// - Network calls
	// - Global variable modifications
	// For now, we'll use heuristics based on file path and content

	stateModifyingPatterns := []string{
		"database", "db", "sql", "gorm",
		"file", "fs", "os", "io",
		"http", "net", "url",
		"cache", "redis", "memcached",
		"queue", "kafka", "rabbitmq",
		"session", "cookie", "auth",
	}

	for _, pattern := range stateModifyingPatterns {
		if strings.Contains(strings.ToLower(analysis.Path), pattern) {
			return true
		}
	}

	return false
}

// isPureUtility checks if the file contains only pure utility functions
func (sa *SafetyAnalyzer) isPureUtility(analysis *FileAnalysis) bool {
	// Pure utility patterns:
	// - Only contains pure functions (no side effects)
	// - No database/file/network operations
	// - No global state modifications
	// - Small, focused functions

	pureUtilityPatterns := []string{
		"utils", "util", "helper", "helpers",
		"math", "string", "time", "format",
		"convert", "transform", "parse",
	}

	for _, pattern := range pureUtilityPatterns {
		if strings.Contains(strings.ToLower(analysis.Path), pattern) {
			return true
		}
	}

	// Check if it's a small file with simple functions
	if analysis.TotalLines <= 50 && analysis.TotalFunctions <= 5 {
		return true
	}

	return false
}

// hasInitFunctions checks if the file contains init() functions
func (sa *SafetyAnalyzer) hasInitFunctions(analysis *FileAnalysis) bool {
	// This would require parsing the file to check for init() functions
	// For now, we'll use a simple heuristic
	return strings.Contains(strings.ToLower(analysis.Path), "init") ||
		strings.Contains(strings.ToLower(analysis.Path), "setup")
}

// hasMainFunction checks if the file contains main() function
func (sa *SafetyAnalyzer) hasMainFunction(analysis *FileAnalysis) bool {
	return strings.HasSuffix(analysis.Path, "main.go") ||
		strings.Contains(analysis.Path, "/cmd/")
}

// hasGlobalVariables checks if the file contains global variables
func (sa *SafetyAnalyzer) hasGlobalVariables(analysis *FileAnalysis) bool {
	// This would require parsing the file to check for global variable declarations
	// For now, we'll use heuristics
	return strings.Contains(strings.ToLower(analysis.Path), "global") ||
		strings.Contains(strings.ToLower(analysis.Path), "var")
}

// isCriticalPackage checks if a file is in a critical package
func (sa *SafetyAnalyzer) isCriticalPackage(filePath string) bool {
	for criticalPkg := range criticalPackages {
		if strings.Contains(filePath, criticalPkg) {
			return true
		}
	}
	return false
}

// AnalyzeFunctionComplexity analyzes the complexity of functions in a file
func (sa *SafetyAnalyzer) AnalyzeFunctionComplexity(file *ast.File) (int, []string) {
	complexityTotal := 0
	var reasons []string

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			complexity := sa.calculateFunctionComplexity(x)
			complexityTotal += complexity
			if complexity > 10 {
				reasons = append(reasons, "High complexity function: "+x.Name.Name)
			}
		}
		return true
	})

	return complexityTotal, reasons
}

// calculateFunctionComplexity calculates cyclomatic complexity of a function
func (sa *SafetyAnalyzer) calculateFunctionComplexity(fn *ast.FuncDecl) int {
	if fn.Body == nil {
		return 1
	}

	complexity := 1 // Base complexity

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.SelectStmt:
			complexity++
		case *ast.BinaryExpr:
			// Check for logical operators that increase complexity
			if x.Op.String() == "||" || x.Op.String() == "&&" {
				complexity++
			}
		}
		return true
	})

	return complexity
}
