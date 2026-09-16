package main

import (
	"fmt"
	"slices"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
)

type platformSpec struct {
	displayName        string
	platformArg        string
	platformSelectors  []string
	outputPathPatterns []string
	additionalParams   string
}

func (s platformSpec) buildable(platform string) bool {
	return slices.Contains(s.platformSelectors, platform)
}

func platformSpecs(cfg config) []platformSpec {
	return []platformSpec{
		{
			displayName:        "iOS",
			platformArg:        "ios",
			platformSelectors:  []string{"both", "ios"},
			outputPathPatterns: cfg.IOSOutputPattern,
			additionalParams:   cfg.IOSAdditionalParams,
		},
		{
			displayName:        "Android",
			platformArg:        "android",
			platformSelectors:  []string{"both", "android"},
			outputPathPatterns: cfg.AndroidOutputPattern,
			additionalParams:   cfg.AndroidAdditionalParams,
		},
	}
}

// buildReleaseArgs constructs the full argv (excluding the "shorebird" binary itself) for a single
// `shorebird release <platform>` invocation, mirroring https://docs.shorebird.dev/code-push/release/'s
// Options / iOS-specific options.
func buildReleaseArgs(cfg config, spec platformSpec, flutterVersion string) ([]string, error) {
	args := []string{"release", spec.platformArg}

	if cfg.DryRun {
		args = append(args, "--dry-run")
	}
	if cfg.Target != "" {
		args = append(args, "-t", cfg.Target)
	}
	if cfg.Flavor != "" {
		args = append(args, "--flavor="+cfg.Flavor)
	}
	if cfg.BuildName != "" {
		args = append(args, "--build-name="+cfg.BuildName)
	}
	if cfg.BuildNumber != "" {
		args = append(args, "--build-number="+cfg.BuildNumber)
	}
	for _, kv := range cfg.DartDefines {
		args = append(args, "--dart-define="+kv)
	}
	for _, f := range cfg.DartDefineFiles {
		args = append(args, "--dart-define-from-file="+f)
	}
	if cfg.Obfuscate {
		args = append(args, "--obfuscate")
	}
	if cfg.SplitDebugInfo != "" {
		args = append(args, "--split-debug-info="+cfg.SplitDebugInfo)
	}
	if flutterVersion != "" {
		args = append(args, "--flutter-version="+flutterVersion)
	}

	if spec.platformArg == "android" {
		artifact := "apk"
		if cfg.AndroidOutputType == "appbundle" {
			artifact = "aab"
		}
		args = append(args, "--artifact="+artifact)
		for _, arch := range cfg.AndroidTargetPlatform {
			args = append(args, "--target-platform="+arch)
		}
	}

	if spec.platformArg == "ios" {
		if !cfg.IOSCodesign {
			args = append(args, "--no-codesign")
		}
		if cfg.IOSExportMethod != "" {
			args = append(args, "--export-method="+cfg.IOSExportMethod)
		}
		if cfg.IOSExportOptionsPlist != "" {
			args = append(args, "--export-options-plist="+cfg.IOSExportOptionsPlist)
		}
	}

	if strings.TrimSpace(cfg.AdditionalParams) != "" {
		extra, err := shellquote.Split(cfg.AdditionalParams)
		if err != nil {
			return nil, fmt.Errorf("parse additional_params: %w", err)
		}
		args = append(args, extra...)
	}

	if strings.TrimSpace(spec.additionalParams) != "" {
		extra, err := shellquote.Split(spec.additionalParams)
		if err != nil {
			return nil, fmt.Errorf("parse %s additional parameters: %w", spec.displayName, err)
		}
		args = append(args, "--")
		args = append(args, extra...)
	}

	return args, nil
}
