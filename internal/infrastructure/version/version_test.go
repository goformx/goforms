package version_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/infrastructure/version"
)

func TestVersionInfo(t *testing.T) {
	info := version.GetInfo()

	// Test basic version info
	require.NotEmpty(t, info.Version, "Version should not be empty")
	require.NotEmpty(t, info.BuildTime, "BuildTime should not be empty")
	require.NotEmpty(t, info.GitCommit, "GitCommit should not be empty")
	require.NotEmpty(t, info.GoVersion, "GoVersion should not be empty")
}

func TestVersionValidation(t *testing.T) {
	tests := []struct {
		name    string
		info    version.Info
		wantErr bool
	}{
		{
			name: "valid version",
			info: version.Info{
				Version:   "1.0.0",
				BuildTime: time.Now().Format(time.RFC3339),
				GitCommit: "1234567",
				GoVersion: "go1.21.0",
			},
			wantErr: false,
		},
		{
			name: "empty version",
			info: version.Info{
				Version:   "",
				BuildTime: time.Now().Format(time.RFC3339),
				GitCommit: "1234567",
				GoVersion: "go1.21.0",
			},
			wantErr: true,
		},
		{
			name: "invalid build time",
			info: version.Info{
				Version:   "1.0.0",
				BuildTime: "invalid-time",
				GitCommit: "1234567",
				GoVersion: "go1.21.0",
			},
			wantErr: true,
		},
		{
			name: "short git commit",
			info: version.Info{
				Version:   "1.0.0",
				BuildTime: time.Now().Format(time.RFC3339),
				GitCommit: "123",
				GoVersion: "go1.21.0",
			},
			wantErr: true,
		},
		{
			name: "invalid go version",
			info: version.Info{
				Version:   "1.0.0",
				BuildTime: time.Now().Format(time.RFC3339),
				GitCommit: "1234567",
				GoVersion: "1.21.0",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.info.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVersionComparison(t *testing.T) {
	tests := []struct {
		name     string
		version1 version.Info
		version2 version.Info
		want     int
	}{
		{
			name:     "equal versions",
			version1: version.Info{Version: "1.0.0"},
			version2: version.Info{Version: "1.0.0"},
			want:     0,
		},
		{
			name:     "version1 greater",
			version1: version.Info{Version: "2.0.0"},
			version2: version.Info{Version: "1.0.0"},
			want:     1,
		},
		{
			name:     "version2 greater",
			version1: version.Info{Version: "1.0.0"},
			version2: version.Info{Version: "2.0.0"},
			want:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version1.Compare(tt.version2)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestVersionStatus(t *testing.T) {
	tests := []struct {
		name  string
		info  version.Info
		isDev bool
		isRel bool
		isPre bool
	}{
		{
			name:  "development version",
			info:  version.Info{Version: "dev"},
			isDev: true,
			isRel: false,
			isPre: false,
		},
		{
			name:  "release version",
			info:  version.Info{Version: "1.0.0"},
			isDev: false,
			isRel: true,
			isPre: false,
		},
		{
			name:  "pre-release version",
			info:  version.Info{Version: "v1.0.0-rc1"},
			isDev: false,
			isRel: false,
			isPre: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.isDev, tt.info.IsDev())
			require.Equal(t, tt.isRel, tt.info.IsRelease())
			require.Equal(t, tt.isPre, tt.info.IsPreRelease())
		})
	}
}

func TestGetBuildTime(t *testing.T) {
	validTime := time.Now().Format(time.RFC3339)
	info := version.Info{BuildTime: validTime}

	got, err := info.GetBuildTime()
	require.NoError(t, err)
	require.NotZero(t, got)

	// Test invalid time
	info.BuildTime = "invalid-time"
	_, err = info.GetBuildTime()
	require.Error(t, err)
}
