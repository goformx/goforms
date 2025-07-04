package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Parser handles package loading and AST parsing
type Parser struct {
	fset *token.FileSet
	pkgs []*packages.Package
}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{
		fset: token.NewFileSet(),
	}
}

// LoadProject loads all packages in the project
func (p *Parser) LoadProject(projectRoot string) error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedDeps,
		Fset: p.fset,
		Dir:  projectRoot,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	// Check for errors
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			// Log errors but don't fail
			fmt.Printf("Package %s has errors: %v\n", pkg.PkgPath, pkg.Errors)
		}
	}

	p.pkgs = pkgs

	return nil
}

// GetPackages returns the loaded packages
func (p *Parser) GetPackages() []*packages.Package {
	return p.pkgs
}

// ParseFile parses a single Go file
func (p *Parser) ParseFile(filePath string) (*ast.File, error) {
	file, err := parser.ParseFile(p.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	return file, nil
}

// GetGoFiles returns all Go files in the specified directory
func (p *Parser) GetGoFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

// AnalyzeFile performs basic analysis on a file
func (p *Parser) AnalyzeFile(file *ast.File, analysis *FileAnalysis) {
	// Get package name
	analysis.PackageName = file.Name.Name

	// Count lines
	analysis.TotalLines = p.fset.File(file.Pos()).LineCount()

	// Analyze functions and detect patterns
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			analysis.TotalFunctions++
			// Check if function is exported
			if x.Name.IsExported() {
				analysis.ExportedFunctions = append(analysis.ExportedFunctions, x.Name.Name)
			}
		case *ast.InterfaceType:
			analysis.HasInterfaces = true
			analysis.Reasons = append(analysis.Reasons, "Contains interface definitions")
		case *ast.ImportSpec:
			if x.Path != nil {
				analysis.ImportedPackages = append(analysis.ImportedPackages, strings.Trim(x.Path.Value, "\""))
			}
		}

		return true
	})
}
