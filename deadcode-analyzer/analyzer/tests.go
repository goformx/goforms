package analyzer

import (
	"go/ast"
	"strings"
)

// TestAnalyzer handles test file analysis
type TestAnalyzer struct{}

// NewTestAnalyzer creates a new test analyzer
func NewTestAnalyzer() *TestAnalyzer {
	return &TestAnalyzer{}
}

// Analyze detects test-related patterns in a file
func (ta *TestAnalyzer) Analyze(file *ast.File, analysis *FileAnalysis) {
	ta.detectTestFile(analysis)
	ast.Inspect(file, func(n ast.Node) bool {
		ta.detectTestFunction(n, analysis)
		ta.detectTestImport(n, analysis)
		ta.detectTestCall(n, analysis)
		return true
	})
}

func (ta *TestAnalyzer) detectTestFile(analysis *FileAnalysis) {
	if strings.HasSuffix(analysis.Path, "_test.go") {
		analysis.IsTestFile = true
		analysis.Reasons = append(analysis.Reasons, "Test file")
	}
}

func (ta *TestAnalyzer) detectTestFunction(n ast.Node, analysis *FileAnalysis) {
	fn, ok := n.(*ast.FuncDecl)
	if !ok {
		return
	}
	if strings.HasPrefix(fn.Name.Name, "Test") ||
		strings.HasPrefix(fn.Name.Name, "Benchmark") ||
		strings.HasPrefix(fn.Name.Name, "Example") {
		analysis.HasTests = true
		analysis.Reasons = append(analysis.Reasons, "Contains test functions")
	}
}

func (ta *TestAnalyzer) detectTestImport(n ast.Node, analysis *FileAnalysis) {
	imp, ok := n.(*ast.ImportSpec)
	if !ok || imp.Path == nil {
		return
	}
	importPath := strings.Trim(imp.Path.Value, "\"")
	if strings.Contains(importPath, "testing") ||
		strings.Contains(importPath, "testify") {
		analysis.HasTests = true
		analysis.Reasons = append(analysis.Reasons, "Imports testing packages")
	}
}

func (ta *TestAnalyzer) detectTestCall(n ast.Node, analysis *FileAnalysis) {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return
	}
	if strings.Contains(strings.ToLower(ident.Name), "test") ||
		strings.Contains(strings.ToLower(sel.Sel.Name), "assert") ||
		strings.Contains(strings.ToLower(sel.Sel.Name), "require") {
		analysis.HasTests = true
		analysis.Reasons = append(analysis.Reasons, "Contains testing calls")
	}
}

// DetectTestPatterns detects specific test patterns
func (ta *TestAnalyzer) DetectTestPatterns(file *ast.File) []string {
	var patterns []string
	ast.Inspect(file, func(n ast.Node) bool {
		ta.appendTestFunctionPattern(n, &patterns)
		ta.appendTestCallPattern(n, &patterns)
		return true
	})
	return patterns
}

func (ta *TestAnalyzer) appendTestFunctionPattern(n ast.Node, patterns *[]string) {
	fn, ok := n.(*ast.FuncDecl)
	if !ok {
		return
	}
	if strings.HasPrefix(fn.Name.Name, "Test") ||
		strings.HasPrefix(fn.Name.Name, "Benchmark") ||
		strings.HasPrefix(fn.Name.Name, "Example") {
		*patterns = append(*patterns, fn.Name.Name)
	}
}

func (ta *TestAnalyzer) appendTestCallPattern(n ast.Node, patterns *[]string) {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return
	}
	if strings.Contains(strings.ToLower(ident.Name), "test") ||
		strings.Contains(strings.ToLower(sel.Sel.Name), "assert") ||
		strings.Contains(strings.ToLower(sel.Sel.Name), "require") {
		*patterns = append(*patterns, ident.Name+"."+sel.Sel.Name)
	}
}

// IsTestFile checks if a file is a test file
func (ta *TestAnalyzer) IsTestFile(filePath string) bool {
	return strings.HasSuffix(filePath, "_test.go")
}
