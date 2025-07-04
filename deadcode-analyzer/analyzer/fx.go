package analyzer

import (
	"go/ast"
	"go/token"
	"strings"
)

// FxDetector handles detection of Fx dependency injection patterns
type FxDetector struct {
	providedFunctions map[string]bool      // functions provided via fx.Provide
	invokedFunctions  map[string]bool      // functions invoked via fx.Invoke
	modules           map[string]*FxModule // fx.Module definitions
	constructors      map[string]bool      // constructor functions
}

// FxModule represents an fx.Module definition
type FxModule struct {
	Name     string
	Provides []string
	Invokes  []string
	Options  []string
	File     string
}

// NewFxDetector creates a new Fx detector
func NewFxDetector() *FxDetector {
	return &FxDetector{
		providedFunctions: make(map[string]bool),
		invokedFunctions:  make(map[string]bool),
		modules:           make(map[string]*FxModule),
		constructors:      make(map[string]bool),
	}
}

// Analyze detects Fx patterns in a file
func (fd *FxDetector) Analyze(file *ast.File, analysis *FileAnalysis) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			fd.analyzeFxCall(x, analysis)
		case *ast.ImportSpec:
			fd.analyzeFxImport(x, analysis)
		case *ast.FuncDecl:
			fd.analyzeConstructor(x, analysis)
		case *ast.AssignStmt:
			fd.analyzeFxAssignment(x, analysis)
		}

		return true
	})
}

// analyzeFxCall analyzes Fx function calls
func (fd *FxDetector) analyzeFxCall(call *ast.CallExpr, analysis *FileAnalysis) {
	if sel, ok1 := call.Fun.(*ast.SelectorExpr); ok1 {
		if ident, ok2 := sel.X.(*ast.Ident); ok2 {
			if ident.Name == "fx" {
				analysis.HasFxUsage = true
				analysis.Reasons = append(analysis.Reasons, "Contains Fx dependency injection usage")

				switch sel.Sel.Name {
				case "Provide":
					fd.analyzeFxProvide(call, analysis)
				case "Invoke":
					fd.analyzeFxInvoke(call, analysis)
				case "Module":
					fd.analyzeFxModule(call, analysis)
				case "Annotate":
					fd.analyzeFxAnnotate(call, analysis)
				case "Options":
					fd.analyzeFxOptions(call, analysis)
				}
			}
		}
	}
}

// analyzeFxProvide analyzes fx.Provide calls
func (fd *FxDetector) analyzeFxProvide(call *ast.CallExpr, analysis *FileAnalysis) {
	for _, arg := range call.Args {
		switch v := arg.(type) {
		case *ast.FuncLit:
			analysis.Reasons = append(analysis.Reasons, "Provides anonymous function")
		case *ast.Ident:
			fd.providedFunctions[v.Name] = true
			analysis.Reasons = append(analysis.Reasons, "Provides function: "+v.Name)
		case *ast.SelectorExpr:
			if pkgIdent, ok1 := v.X.(*ast.Ident); ok1 {
				funcName := pkgIdent.Name + "." + v.Sel.Name
				fd.providedFunctions[funcName] = true
				analysis.Reasons = append(analysis.Reasons, "Provides function: "+funcName)
			}
		}
	}
}

// analyzeFxInvoke analyzes fx.Invoke calls
func (fd *FxDetector) analyzeFxInvoke(call *ast.CallExpr, analysis *FileAnalysis) {
	for _, arg := range call.Args {
		switch v := arg.(type) {
		case *ast.FuncLit:
			analysis.Reasons = append(analysis.Reasons, "Invokes anonymous function")
		case *ast.Ident:
			fd.invokedFunctions[v.Name] = true
			analysis.Reasons = append(analysis.Reasons, "Invokes function: "+v.Name)
		case *ast.SelectorExpr:
			if pkgIdent, ok1 := v.X.(*ast.Ident); ok1 {
				funcName := pkgIdent.Name + "." + v.Sel.Name
				fd.invokedFunctions[funcName] = true
				analysis.Reasons = append(analysis.Reasons, "Invokes function: "+funcName)
			}
		}
	}
}

// analyzeFxModule analyzes fx.Module calls
func (fd *FxDetector) analyzeFxModule(call *ast.CallExpr, analysis *FileAnalysis) {
	if len(call.Args) < 2 {
		return
	}

	nameLit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || nameLit.Kind != token.STRING {
		return
	}

	moduleName := strings.Trim(nameLit.Value, "\"")
	module := &FxModule{
		Name: moduleName,
		File: analysis.Path,
	}
	// Parse module options
	fd.parseModuleOptions(call.Args[1], module)
	fd.modules[moduleName] = module
	analysis.Reasons = append(analysis.Reasons, "Defines Fx module: "+moduleName)
}

// parseModuleOptions parses the options for an Fx module
func (fd *FxDetector) parseModuleOptions(arg ast.Expr, module *FxModule) {
	options, ok := arg.(*ast.CompositeLit)
	if !ok {
		return
	}

	for _, option := range options.Elts {
		if call, ok1 := option.(*ast.CallExpr); ok1 {
			if sel, ok2 := call.Fun.(*ast.SelectorExpr); ok2 {
				module.Options = append(module.Options, sel.Sel.Name)
			}
		}
	}
}

// analyzeFxAnnotate analyzes fx.Annotate calls
func (fd *FxDetector) analyzeFxAnnotate(_ *ast.CallExpr, analysis *FileAnalysis) {
	analysis.Reasons = append(analysis.Reasons, "Uses Fx annotations")
}

// analyzeFxOptions analyzes fx.Options calls
func (fd *FxDetector) analyzeFxOptions(_ *ast.CallExpr, analysis *FileAnalysis) {
	analysis.Reasons = append(analysis.Reasons, "Uses Fx options grouping")
}

// analyzeFxImport analyzes Fx imports
func (fd *FxDetector) analyzeFxImport(importSpec *ast.ImportSpec, analysis *FileAnalysis) {
	if importSpec.Path != nil {
		importPath := strings.Trim(importSpec.Path.Value, "\"")
		if strings.Contains(importPath, "go.uber.org/fx") {
			analysis.HasFxUsage = true
			analysis.Reasons = append(analysis.Reasons, "Imports Fx framework")
		}
	}
}

// analyzeConstructor analyzes constructor functions
func (fd *FxDetector) analyzeConstructor(fn *ast.FuncDecl, analysis *FileAnalysis) {
	// Check for constructor patterns
	if fd.isConstructor(fn) {
		fd.constructors[fn.Name.Name] = true
		analysis.Reasons = append(analysis.Reasons, "Contains constructor function: "+fn.Name.Name)
	}

	// Check for reflection usage in constructors
	if fd.hasReflectionUsage(fn) {
		analysis.Reasons = append(analysis.Reasons, "Uses reflection in constructor")
	}
}

// analyzeFxAssignment analyzes Fx-related assignments
func (fd *FxDetector) analyzeFxAssignment(assign *ast.AssignStmt, analysis *FileAnalysis) {
	// Look for assignments that might be Fx-related
	for _, rhs := range assign.Rhs {
		if call, ok1 := rhs.(*ast.CallExpr); ok1 {
			fd.analyzeFxCall(call, analysis)
		}
	}
}

// isConstructor checks if a function is a constructor
func (fd *FxDetector) isConstructor(fn *ast.FuncDecl) bool {
	// Constructor patterns:
	// 1. Function name starts with "New"
	// 2. Function name starts with "Create"
	// 3. Function name starts with "Provide"
	// 4. Function returns a struct or interface
	name := fn.Name.Name

	return strings.HasPrefix(name, "New") ||
		strings.HasPrefix(name, "Create") ||
		strings.HasPrefix(name, "Provide") ||
		strings.HasPrefix(name, "Make")
}

// hasReflectionUsage checks if a function uses reflection
func (fd *FxDetector) hasReflectionUsage(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}

	hasReflection := false

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if sel, ok1 := x.Fun.(*ast.SelectorExpr); ok1 {
				if ident, ok2 := sel.X.(*ast.Ident); ok2 {
					if ident.Name == "reflect" {
						hasReflection = true

						return false // Stop traversal
					}
				}
			}
		}

		return true
	})

	return hasReflection
}

// DetectFxPatterns detects specific Fx patterns
func (fd *FxDetector) DetectFxPatterns(file *ast.File) []string {
	var patterns []string

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if sel, ok1 := x.Fun.(*ast.SelectorExpr); ok1 {
				if ident, ok2 := sel.X.(*ast.Ident); ok2 {
					if ident.Name == "fx" {
						patterns = append(patterns, "fx."+sel.Sel.Name)
					}
				}
			}
		}

		return true
	})

	return patterns
}

// IsFunctionProvided checks if a function is provided via Fx
func (fd *FxDetector) IsFunctionProvided(funcName string) bool {
	return fd.providedFunctions[funcName]
}

// IsFunctionInvoked checks if a function is invoked via Fx
func (fd *FxDetector) IsFunctionInvoked(funcName string) bool {
	return fd.invokedFunctions[funcName]
}

// IsConstructor checks if a function is a constructor
func (fd *FxDetector) IsConstructor(funcName string) bool {
	return fd.constructors[funcName]
}

// GetModules returns all Fx modules
func (fd *FxDetector) GetModules() map[string]*FxModule {
	return fd.modules
}
