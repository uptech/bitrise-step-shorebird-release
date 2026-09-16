package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseShorebirdOutput(t *testing.T) {
	tests := []struct {
		name           string
		stdout         string
		wantReleaseVer string
	}{
		{
			name:           "release version with ansi colors and emoji",
			stdout:         "Some build logs...\n\x1b[36m📦 Release Version: 1.2.3\x1b[0m\nDone",
			wantReleaseVer: "1.2.3",
		},
		{
			name:   "no match returns empty, never errors",
			stdout: "nothing useful here",
		},
		{
			// Real stdout/stderr captured from a live `shorebird release android` run: the CLI's
			// progress spinner prints "Starting <task>..." then "Done <task result>" on completion,
			// with a build-metadata version (contains a literal "+").
			name:           "real release progress spinner output",
			stdout:         "Starting Determining release version...\nDone Release version: 9.25.0+509\nStarting Fetching releases...\nDone Fetching releases\n",
			wantReleaseVer: "9.25.0+509",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rv := parseShorebirdOutput(tt.stdout)
			assert.Equal(t, tt.wantReleaseVer, rv)
		})
	}
}

func TestBuildReleaseArgs_Android(t *testing.T) {
	cfg := config{
		DartDefines:       []string{"API_URL=https://example.com", "ENABLE_FOO=true"},
		AndroidOutputType: "appbundle",
		IOSCodesign:       true,
	}
	spec := platformSpec{displayName: "Android", platformArg: "android"}

	args, err := buildReleaseArgs(cfg, spec, "3.24.3")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"release", "android",
		"--dart-define=API_URL=https://example.com",
		"--dart-define=ENABLE_FOO=true",
		"--flutter-version=3.24.3",
		"--artifact=aab",
	}, args)
}

func TestBuildReleaseArgs_AndroidTargetPlatform(t *testing.T) {
	cfg := config{
		AndroidOutputType:     "apk",
		AndroidTargetPlatform: []string{"android-arm64"},
	}
	spec := platformSpec{displayName: "Android", platformArg: "android"}

	args, err := buildReleaseArgs(cfg, spec, "")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"release", "android",
		"--artifact=apk",
		"--target-platform=android-arm64",
	}, args)
}

func TestBuildReleaseArgs_IOS(t *testing.T) {
	cfg := config{
		IOSCodesign:           false,
		IOSExportMethod:       "app-store",
		IOSExportOptionsPlist: "ios/ExportOptions.plist",
	}
	spec := platformSpec{displayName: "iOS", platformArg: "ios"}

	args, err := buildReleaseArgs(cfg, spec, "")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"release", "ios",
		"--no-codesign",
		"--export-method=app-store",
		"--export-options-plist=ios/ExportOptions.plist",
	}, args)
}

func TestBuildReleaseArgs_AdditionalParams(t *testing.T) {
	cfg := config{
		AdditionalParams: `--public-key-cmd="echo hi"`,
	}
	spec := platformSpec{displayName: "Android", platformArg: "android", additionalParams: "--split-per-abi"}

	args, err := buildReleaseArgs(cfg, spec, "")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"release", "android",
		"--artifact=apk",
		"--public-key-cmd=echo hi",
		"--",
		"--split-per-abi",
	}, args)
}
