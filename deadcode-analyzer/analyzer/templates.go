package analyzer

import (
	"go/ast"
	"strings"
)

// TemplateDetector handles detection of template usage
type TemplateDetector struct{}

// NewTemplateDetector creates a new template detector
func NewTemplateDetector() *TemplateDetector {
	return &TemplateDetector{}
}

// Analyze detects template usage in a file
func (td *TemplateDetector) Analyze(file *ast.File, analysis *FileAnalysis) {
	ast.Inspect(file, func(n ast.Node) bool {
		td.detectTemplateCall(n, analysis)
		td.detectTemplateImport(n, analysis)
		td.detectTemplateStructField(n, analysis)
		return true
	})
}

// detectTemplateCall checks for template usage in function calls
func (td *TemplateDetector) detectTemplateCall(n ast.Node, analysis *FileAnalysis) {
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
	if strings.Contains(strings.ToLower(ident.Name), "template") ||
		strings.Contains(strings.ToLower(sel.Sel.Name), "template") {
		analysis.HasTemplates = true
		analysis.Reasons = append(analysis.Reasons, "Contains template usage")
	}
}

// detectTemplateImport checks for template package imports
func (td *TemplateDetector) detectTemplateImport(n ast.Node, analysis *FileAnalysis) {
	imp, ok := n.(*ast.ImportSpec)
	if !ok || imp.Path == nil {
		return
	}
	importPath := strings.Trim(imp.Path.Value, "\"")
	if strings.Contains(importPath, "html/template") ||
		strings.Contains(importPath, "text/template") {
		analysis.HasTemplates = true
		analysis.Reasons = append(analysis.Reasons, "Imports template packages")
	}
}

// detectTemplateStructField checks for template-related struct fields
func (td *TemplateDetector) detectTemplateStructField(n ast.Node, analysis *FileAnalysis) {
	st, ok := n.(*ast.StructType)
	if !ok {
		return
	}
	for _, field := range st.Fields.List {
		if field.Names != nil {
			for _, name := range field.Names {
				if strings.Contains(strings.ToLower(name.Name), "template") {
					analysis.HasTemplates = true
					analysis.Reasons = append(analysis.Reasons, "Contains template-related struct fields")
				}
			}
		}
	}
}

// DetectTemplatePatterns detects specific template patterns
func (td *TemplateDetector) DetectTemplatePatterns(file *ast.File) []string {
	var patterns []string

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if strings.Contains(strings.ToLower(ident.Name), "template") ||
						strings.Contains(strings.ToLower(sel.Sel.Name), "template") {
						patterns = append(patterns, ident.Name+"."+sel.Sel.Name)
					}
				}
			}
		}
		return true
	})

	return patterns
}
