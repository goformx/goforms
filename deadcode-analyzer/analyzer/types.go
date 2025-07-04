package analyzer

// Project structure constants
const (
	InternalDir       = "internal"
	ApplicationDir    = "internal/application"
	DomainDir         = "internal/domain"
	InfrastructureDir = "internal/infrastructure"
	PresentationDir   = "internal/presentation"
)

// Safety levels for deletion
type SafetyLevel int

const (
	UltraSafe SafetyLevel = iota
	PotentiallySafe
	Dangerous
	NeverDelete
)

func (s SafetyLevel) String() string {
	switch s {
	case UltraSafe:
		return "ULTRA-SAFE"
	case PotentiallySafe:
		return "POTENTIALLY-SAFE"
	case Dangerous:
		return "DANGEROUS"
	case NeverDelete:
		return "NEVER-DELETE"
	default:
		return "UNKNOWN"
	}
}

// File analysis result
type FileAnalysis struct {
	Path                 string
	PackageName          string
	TotalFunctions       int
	UnreachableFunctions int
	TotalLines           int
	SafetyLevel          SafetyLevel
	SafetyScore          int
	Reasons              []string
	HasInterfaces        bool
	HasDIUsage           bool
	IsImported           bool
	IsCritical           bool
	HasFxUsage           bool
	HasTests             bool
	HasTemplates         bool
	IsTestFile           bool
	ExportedFunctions    []string
	ImportedPackages     []string
}

// Package information
type PackageInfo struct {
	Path    string
	Name    string
	Files   []string
	Imports []string
	Exports []string
	IsTest  bool
}

// Analysis configuration
type Config struct {
	Verbose     bool
	OutputFile  string
	JSONOutput  bool
	ProjectRoot string
}

// Analysis results
type Results struct {
	Files           []*FileAnalysis
	TotalFiles      int
	UltraSafe       int
	PotentiallySafe int
	Dangerous       int
	NeverDelete     int
	Summary         map[string]interface{}
}
