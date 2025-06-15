package version

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVersionInfo(t *testing.T) {
	info := GetInfo()

	// Test basic version info
	assert.NotEmpty(t, info.Version, "Version should not be empty")
	assert.NotEmpty(t, info.BuildTime, "BuildTime should not be empty")
	assert.NotEmpty(t, info.GitCommit, "GitCommit should not be empty")
	assert.NotEmpty(t, info.GoVersion, "GoVersion should not be empty")
}

func TestVersionValidation(t *testing.T) {
	tests := []struct {
		name    string
		info    Info
		wantErr bool
	}{
		{
			name: "valid version",
			info: Info{
				Version:   "1.0.0",
				BuildTime: time.Now().Format(time.RFC3339),
				GitCommit: "1234567",
				GoVersion: "go1.21.0",
			},
			wantErr: false,
		},
		{
			name: "empty version",
			info: Info{
				Version:   "",
				BuildTime: time.Now().Format(time.RFC3339),
				GitCommit: "1234567",
				GoVersion: "go1.21.0",
			},
			wantErr: true,
		},
		{
			name: "invalid build time",
			info: Info{
				Version:   "1.0.0",
				BuildTime: "invalid-time",
				GitCommit: "1234567",
				GoVersion: "go1.21.0",
			},
			wantErr: true,
		},
		{
			name: "short git commit",
			info: Info{
				Version:   "1.0.0",
				BuildTime: time.Now().Format(time.RFC3339),
				GitCommit: "123",
				GoVersion: "go1.21.0",
			},
			wantErr: true,
		},
		{
			name: "invalid go version",
			info: Info{
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVersionComparison(t *testing.T) {
	tests := []struct {
		name     string
		version1 Info
		version2 Info
		want     int
	}{
		{
			name:     "equal versions",
			version1: Info{Version: "1.0.0"},
			version2: Info{Version: "1.0.0"},
			want:     0,
		},
		{
			name:     "version1 greater",
			version1: Info{Version: "2.0.0"},
			version2: Info{Version: "1.0.0"},
			want:     1,
		},
		{
			name:     "version2 greater",
			version1: Info{Version: "1.0.0"},
			version2: Info{Version: "2.0.0"},
			want:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version1.Compare(tt.version2)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVersionStatus(t *testing.T) {
	tests := []struct {
		name  string
		info  Info
		isDev bool
		isRel bool
		isPre bool
	}{
		{
			name:  "development version",
			info:  Info{Version: "dev"},
			isDev: true,
			isRel: false,
			isPre: false,
		},
		{
			name:  "release version",
			info:  Info{Version: "1.0.0"},
			isDev: false,
			isRel: true,
			isPre: false,
		},
		{
			name:  "pre-release version",
			info:  Info{Version: "v1.0.0-rc1"},
			isDev: false,
			isRel: false,
			isPre: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isDev, tt.info.IsDev())
			assert.Equal(t, tt.isRel, tt.info.IsRelease())
			assert.Equal(t, tt.isPre, tt.info.IsPreRelease())
		})
	}
}

func TestGetBuildTime(t *testing.T) {
	validTime := time.Now().Format(time.RFC3339)
	info := Info{BuildTime: validTime}

	got, err := info.GetBuildTime()
	assert.NoError(t, err)
	assert.NotZero(t, got)

	// Test invalid time
	info.BuildTime = "invalid-time"
	_, err = info.GetBuildTime()
	assert.Error(t, err)
}
