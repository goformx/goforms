package analyzer

import (
	"strings"

	"golang.org/x/tools/go/packages"
)

// ImportAnalyzer handles import/export analysis
type ImportAnalyzer struct {
	importGraph map[string][]string // package -> list of packages that import it
}

// NewImportAnalyzer creates a new import analyzer
func NewImportAnalyzer() *ImportAnalyzer {
	return &ImportAnalyzer{
		importGraph: make(map[string][]string),
	}
}

// BuildGraph constructs the import graph from loaded packages
func (ia *ImportAnalyzer) BuildGraph(pkgs []*packages.Package) error {
	for _, pkg := range pkgs {
		if pkg.Types == nil {
			continue
		}

		for _, importedPkg := range pkg.Imports {
			ia.importGraph[importedPkg.PkgPath] = append(ia.importGraph[importedPkg.PkgPath], pkg.PkgPath)
		}
	}

	return nil
}

// Analyze performs import analysis on a file
func (ia *ImportAnalyzer) Analyze(filePath string, analysis *FileAnalysis) {
	// Check if this file's package is imported by other packages
	packagePath := ia.getPackagePathFromFilePath(filePath)
	if importers, exists := ia.importGraph[packagePath]; exists && len(importers) > 0 {
		analysis.IsImported = true
		analysis.Reasons = append(analysis.Reasons, "Imported by other packages")
	}
}

// getPackagePathFromFilePath extracts the package path from a file path
func (ia *ImportAnalyzer) getPackagePathFromFilePath(filePath string) string {
	// Convert file path to package path
	// e.g., "internal/domain/user/repository.go" -> "github.com/goformx/goforms/internal/domain/user"

	// Remove the filename
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

// IsPackageImported checks if a package is imported by other packages
func (ia *ImportAnalyzer) IsPackageImported(packagePath string) bool {
	importers, exists := ia.importGraph[packagePath]

	return exists && len(importers) > 0
}

// GetImporters returns the list of packages that import a given package
func (ia *ImportAnalyzer) GetImporters(packagePath string) []string {
	return ia.importGraph[packagePath]
}
