package version

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
	GitCommit string `json:"gitCommit"`
	GoVersion string `json:"goVersion"`
}

//nolint:gochecknoglobals // Populated via -ldflags
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
	GoVersion = "unknown"
)
