package analyzer

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// CallGraphAnalyzer handles call graph analysis for reachability detection
type CallGraphAnalyzer struct {
	callGraph     map[string][]string // function -> list of functions it calls
	reverseGraph  map[string][]string // function -> list of functions that call it
	packageGraph  map[string]*PackageInfo
	entryPoints   map[string]bool     // functions that are entry points (main, exported, etc.)
	interfaceImpl map[string][]string // interface -> list of implementing types
}

// NewCallGraphAnalyzer creates a new call graph analyzer
func NewCallGraphAnalyzer() *CallGraphAnalyzer {
	return &CallGraphAnalyzer{
		callGraph:     make(map[string][]string),
		reverseGraph:  make(map[string][]string),
		packageGraph:  make(map[string]*PackageInfo),
		entryPoints:   make(map[string]bool),
		interfaceImpl: make(map[string][]string),
	}
}

// Build constructs the call graph from loaded packages
func (cga *CallGraphAnalyzer) Build(pkgs []*packages.Package) error {
	// First pass: collect all functions and identify entry points
	if err := cga.collectFunctions(pkgs); err != nil {
		return fmt.Errorf("failed to collect functions: %w", err)
	}

	// Second pass: build the call graph
	if err := cga.buildCallGraph(pkgs); err != nil {
		return fmt.Errorf("failed to build call graph: %w", err)
	}

	// Third pass: identify interface implementations
	if err := cga.analyzeInterfaces(pkgs); err != nil {
		return fmt.Errorf("failed to analyze interfaces: %w", err)
	}

	return nil
}

// collectFunctions collects all functions and identifies entry points
func (cga *CallGraphAnalyzer) collectFunctions(pkgs []*packages.Package) error {
	for _, pkg := range pkgs {
		if pkg.Types == nil {
			continue
		}

		packageInfo := &PackageInfo{
			Path:  pkg.PkgPath,
			Name:  pkg.Name,
			Files: make([]string, 0),
		}

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					funcName := fmt.Sprintf("%s.%s", pkg.PkgPath, x.Name.Name)

					// Identify entry points
					if x.Name.Name == "main" ||
						x.Name.Name == "init" ||
						x.Name.IsExported() ||
						cga.isFxEntryPoint(x) {
						cga.entryPoints[funcName] = true
					}

					// Initialize call graph entries
					cga.callGraph[funcName] = make([]string, 0)
					cga.reverseGraph[funcName] = make([]string, 0)
				}

				return true
			})
		}

		cga.packageGraph[pkg.PkgPath] = packageInfo
	}

	return nil
}

// buildCallGraph builds the actual call graph by analyzing function calls
func (cga *CallGraphAnalyzer) buildCallGraph(pkgs []*packages.Package) error {
	for _, pkg := range pkgs {
		if pkg.Types == nil {
			continue
		}

		for _, file := range pkg.Syntax {
			cga.analyzeFileForCalls(file, pkg)
		}
	}

	return nil
}

// analyzeFileForCalls analyzes a single file for function calls
func (cga *CallGraphAnalyzer) analyzeFileForCalls(file *ast.File, pkg *packages.Package) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Body == nil {
				return true
			}

			callerName := fmt.Sprintf("%s.%s", pkg.PkgPath, x.Name.Name)
			cga.analyzeFunctionCalls(x.Body, callerName, pkg)
		}

		return true
	})
}

// analyzeFunctionCalls analyzes function calls within a function body
func (cga *CallGraphAnalyzer) analyzeFunctionCalls(body *ast.BlockStmt, callerName string, pkg *packages.Package) {
	ast.Inspect(body, func(n ast.Node) bool {
		switch call := n.(type) {
		case *ast.CallExpr:
			cga.processFunctionCall(call, callerName, pkg)
		}

		return true
	})
}

// processFunctionCall processes a single function call
func (cga *CallGraphAnalyzer) processFunctionCall(call *ast.CallExpr, callerName string, pkg *packages.Package) {
	calledFuncs := cga.resolveFunctionCall(call, pkg)
	for _, called := range calledFuncs {
		if called != "" {
			cga.callGraph[callerName] = append(cga.callGraph[callerName], called)
			cga.reverseGraph[called] = append(cga.reverseGraph[called], callerName)
		}
	}
}

// resolveFunctionCall resolves a function call to its actual function name
func (cga *CallGraphAnalyzer) resolveFunctionCall(call *ast.CallExpr, pkg *packages.Package) []string {
	var results []string

	switch fun := call.Fun.(type) {
	case *ast.Ident:
		// Direct function call: funcName()
		results = append(results, fmt.Sprintf("%s.%s", pkg.PkgPath, fun.Name))

	case *ast.SelectorExpr:
		// Method call: obj.method() or pkg.func()
		if ident, ok := fun.X.(*ast.Ident); ok {
			results = append(results, cga.resolveSelectorCall(ident, fun.Sel.Name, pkg)...)
		}
	}

	return results
}

// resolveSelectorCall resolves a selector call (obj.method or pkg.func) to possible function names
func (cga *CallGraphAnalyzer) resolveSelectorCall(ident *ast.Ident, selName string, pkg *packages.Package) []string {
	if pkg.TypesInfo == nil {
		// Fallback: assume it's a local method call
		return []string{fmt.Sprintf("%s.%s.%s", pkg.PkgPath, ident.Name, selName)}
	}

	obj := pkg.TypesInfo.ObjectOf(ident)
	if obj == nil {
		// Fallback: assume it's a local method call
		return []string{fmt.Sprintf("%s.%s.%s", pkg.PkgPath, ident.Name, selName)}
	}

	if pkgName, ok := obj.(*types.PkgName); ok {
		// It's a package import: pkg.func()
		return []string{fmt.Sprintf("%s.%s", pkgName.Imported().Path(), selName)}
	}

	// It's a local variable: obj.method()
	// This is more complex and would need type information
	return []string{fmt.Sprintf("%s.%s.%s", pkg.PkgPath, ident.Name, selName)}
}

// analyzeInterfaces analyzes interface implementations
func (cga *CallGraphAnalyzer) analyzeInterfaces(pkgs []*packages.Package) error {
	for _, pkg := range pkgs {
		if pkg.Types == nil {
			continue
		}

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.InterfaceType:
					// Interface types are handled in TypeSpec
				case *ast.TypeSpec:
					// Check if this type implements any interfaces
					if x.Type != nil {
						if structType, ok := x.Type.(*ast.StructType); ok {
							typeName := fmt.Sprintf("%s.%s", pkg.PkgPath, x.Name.Name)
							cga.findInterfaceImplementations(typeName, structType, pkg)
						}
					}
				}

				return true
			})
		}
	}

	return nil
}

// findInterfaceImplementations finds which interfaces a type implements
func (cga *CallGraphAnalyzer) findInterfaceImplementations(typeName string, _ *ast.StructType, _ *packages.Package) {
	// This is a simplified implementation
	// In a full implementation, we'd need to check method signatures
	// For now, we'll just note that this type might implement interfaces
	for interfaceName := range cga.interfaceImpl {
		cga.interfaceImpl[interfaceName] = append(cga.interfaceImpl[interfaceName], typeName)
	}
}

// isFxEntryPoint checks if a function is an Fx entry point
func (cga *CallGraphAnalyzer) isFxEntryPoint(fn *ast.FuncDecl) bool {
	// Check for Fx-specific patterns
	if fn.Body == nil {
		return false
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch call := n.(type) {
		case *ast.CallExpr:
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "fx" {
						switch sel.Sel.Name {
						case "Provide", "Invoke", "Annotate":
							return false // Stop traversal, found Fx usage
						}
					}
				}
			}
		}

		return true
	})

	return false
}

// AnalyzeReachability determines which functions in a file are unreachable
func (cga *CallGraphAnalyzer) AnalyzeReachability(filePath string, analysis *FileAnalysis) {
	// Get the package path from the file path
	packagePath := cga.getPackagePathFromFilePath(filePath)

	// Find all functions in this file
	var fileFunctions []string
	for funcName := range cga.callGraph {
		if strings.HasPrefix(funcName, packagePath) {
			fileFunctions = append(fileFunctions, funcName)
		}
	}

	// Count unreachable functions
	unreachableCount := 0
	for _, funcName := range fileFunctions {
		if !cga.isFunctionReachable(funcName) {
			unreachableCount++
		}
	}

	analysis.UnreachableFunctions = unreachableCount
}

// isFunctionReachable determines if a function is reachable from any entry point
func (cga *CallGraphAnalyzer) isFunctionReachable(funcName string) bool {
	// If it's an entry point, it's reachable
	if cga.entryPoints[funcName] {
		return true
	}

	// Check if any entry point can reach this function
	visited := make(map[string]bool)

	for entryPoint := range cga.entryPoints {
		if cga.canReach(entryPoint, funcName, visited) {
			return true
		}
	}

	return false
}

// canReach checks if start can reach target using DFS
func (cga *CallGraphAnalyzer) canReach(start, target string, visited map[string]bool) bool {
	if start == target {
		return true
	}

	if visited[start] {
		return false
	}

	visited[start] = true

	// Check all functions called by start
	for _, called := range cga.callGraph[start] {
		if cga.canReach(called, target, visited) {
			return true
		}
	}

	return false
}

// getPackagePathFromFilePath extracts the package path from a file path
func (cga *CallGraphAnalyzer) getPackagePathFromFilePath(filePath string) string {
	// Convert file path to package path
	// e.g., "internal/domain/user/repository.go" -> "github.com/goformx/goforms/internal/domain/user"

	parts := strings.Split(filePath, "/")
	if len(parts) < 2 {
		return ""
	}

	// Find the "internal" directory
	internalIndex := -1
	for i, part := range parts {
		if part == "internal" {
			internalIndex = i

			break
		}
	}

	if internalIndex == -1 {
		return ""
	}

	// Build the package path
	packageParts := parts[internalIndex:]
	if len(packageParts) > 1 {
		// Remove the .go extension from the last part
		lastPart := packageParts[len(packageParts)-1]
		if strings.HasSuffix(lastPart, ".go") {
			packageParts[len(packageParts)-1] = strings.TrimSuffix(lastPart, ".go")
		}
	}

	return "github.com/goformx/goforms/" + strings.Join(packageParts, "/")
}

// IsFunctionReachable checks if a specific function is reachable
func (cga *CallGraphAnalyzer) IsFunctionReachable(funcName string) bool {
	return cga.isFunctionReachable(funcName)
}

// GetCallers returns all functions that call the given function
func (cga *CallGraphAnalyzer) GetCallers(funcName string) []string {
	return cga.reverseGraph[funcName]
}

// GetCallees returns all functions called by the given function
func (cga *CallGraphAnalyzer) GetCallees(funcName string) []string {
	return cga.callGraph[funcName]
}
