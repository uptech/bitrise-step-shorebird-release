package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-steputils/tools"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

type config struct {
	ProjectLocation string          `env:"project_location,dir"`
	ShorebirdToken  stepconf.Secret `env:"shorebird_token"`
	CacheLevel      string          `env:"cache_level,opt[all,none]"`
	IsDebugMode     bool            `env:"is_debug_mode,opt[true,false]"`

	Platform              string   `env:"platform,opt[both,ios,android]"`
	FlutterVersion        string   `env:"flutter_version"`
	Flavor                string   `env:"flavor"`
	Target                string   `env:"target"`
	BuildName             string   `env:"build_name"`
	BuildNumber           string   `env:"build_number"`
	DartDefines           []string `env:"dart_defines,multiline"`
	DartDefineFiles       []string `env:"dart_define_files,multiline"`
	Obfuscate             bool     `env:"obfuscate,opt[true,false]"`
	SplitDebugInfo        string   `env:"split_debug_info"`
	DryRun                bool     `env:"dry_run,opt[true,false]"`
	AdditionalParams      string   `env:"additional_params"`
	AndroidTargetPlatform []string `env:"android_target_platform,multiline"`

	IOSCodesign           bool     `env:"ios_codesign,opt[true,false]"`
	IOSCodesignIdentity   string   `env:"ios_codesign_identity"`
	IOSExportMethod       string   `env:"ios_export_method,opt[app-store,ad-hoc,development,enterprise]"`
	IOSExportOptionsPlist string   `env:"ios_export_options_plist"`
	IOSAdditionalParams   string   `env:"ios_additional_params"`
	IOSOutputPattern      []string `env:"ios_output_pattern,multiline"`

	AndroidOutputType       string   `env:"android_output_type,opt[apk,appbundle]"`
	AndroidAdditionalParams string   `env:"android_additional_params"`
	AndroidOutputPattern    []string `env:"android_output_pattern,multiline"`
}

func failf(msg string, args ...interface{}) {
	log.Errorf(msg, args...)
	os.Exit(1)
}

func main() {
	var cfg config
	if err := stepconf.Parse(&cfg); err != nil {
		failf("Process config: failed to parse input: %s", err)
	}
	stepconf.Print(cfg)
	log.SetEnableDebugLog(cfg.IsDebugMode)

	projectLocationAbs, err := filepath.Abs(cfg.ProjectLocation)
	if err != nil {
		failf("Process config: failed to get absolute project path of %s: %s", cfg.ProjectLocation, err)
	}

	exist, err := pathutil.IsDirExists(projectLocationAbs)
	if err != nil {
		failf("Process config: failed to check if project path exists: %s", err)
	} else if !exist {
		failf("Process config: project path does not exist")
	}

	if string(cfg.ShorebirdToken) == "" {
		failf("Process config: shorebird_token is empty. Set the 'Shorebird API token' input to a Shorebird Console API key (Account > API Keys), stored as a Bitrise Secret.")
	}
	if err := os.Setenv("SHOREBIRD_TOKEN", string(cfg.ShorebirdToken)); err != nil {
		failf("Process config: failed to set SHOREBIRD_TOKEN env var: %s", err)
	}
	if err := os.Setenv("CI", "true"); err != nil {
		failf("Process config: failed to set CI env var: %s", err)
	}

	fmt.Println()
	log.Infof("Ensuring Shorebird CLI is installed")
	shorebirdBin, err := ensureShorebirdCLI()
	if err != nil {
		failf("Run: failed to install/locate Shorebird CLI: %s", err)
	}

	if err := verifyShorebirdProject(projectLocationAbs); err != nil {
		failf("Process config: %s", err)
	}

	var flutterVersion string
	if cfg.FlutterVersion != "" {
		flutterVersion = cfg.FlutterVersion
	} else {
		fmt.Println()
		log.Infof("Detecting installed Flutter version")
		flutterVersion, err = detectFlutterVersion()
		if err != nil {
			failf("Process config: %s", err)
		}
		log.Donef("- Detected Flutter version: %s", flutterVersion)
	}

	if cfg.Platform == "ios" || cfg.Platform == "both" {
		if err := prepareIOSCodesign(cfg); err != nil {
			failf("Run: %s", err)
		}
	}

	var releaseVersion string
	for _, spec := range platformSpecs(cfg) {
		if !spec.buildable(cfg.Platform) {
			continue
		}

		args, err := buildReleaseArgs(cfg, spec, flutterVersion)
		if err != nil {
			failf("Process config: failed to build shorebird arguments for %s: %s", spec.displayName, err)
		}

		fmt.Println()
		log.Infof("Shorebird release: %s", spec.displayName)
		stdout, err := runShorebirdCommand(shorebirdBin, projectLocationAbs, args)
		if err != nil {
			failf("Run: shorebird release failed for %s: %s", spec.displayName, err)
		}

		if rv := parseShorebirdOutput(stdout); rv != "" {
			releaseVersion = rv
		}

		fmt.Println()
		log.Infof("Export %s artifact", spec.displayName)

		artifacts, err := findArtifacts(projectLocationAbs, spec.outputPathPatterns)
		if err != nil {
			failf("Export outputs: failed to find artifacts: %s", err)
		}
		if len(artifacts) < 1 {
			failf(`Export outputs: artifact path pattern (%s) did not match any artifacts on the path (%s).
Check that the output pattern and 'Project Location' are correct.`, spec.outputPathPatterns, projectLocationAbs)
		}

		if err := exportArtifacts(cfg, spec, artifacts); err != nil {
			failf("Export outputs: failed to export %s artifacts: %s", spec.displayName, err)
		}
	}

	if releaseVersion != "" {
		if err := tools.ExportEnvironmentWithEnvman("SHOREBIRD_RELEASE_VERSION", releaseVersion); err != nil {
			log.Warnf("Failed to export SHOREBIRD_RELEASE_VERSION: %s", err)
		} else {
			log.Donef("- SHOREBIRD_RELEASE_VERSION: %s", releaseVersion)
		}
	} else {
		log.Debugf("Could not parse a release version from the Shorebird CLI's output")
	}

	if cfg.CacheLevel == "all" {
		fmt.Println()
		log.Infof("Collecting cache")

		if err := cacheCocoapodsDeps(projectLocationAbs); err != nil {
			log.Warnf("Failed to collect cocoapods cache, error: %s", err)
		}
		if err := cacheCarthageDeps(projectLocationAbs); err != nil {
			log.Warnf("Failed to collect carthage cache, error: %s", err)
		}
		if err := cacheAndroidDeps(projectLocationAbs); err != nil {
			log.Warnf("Failed to collect android cache, error: %s", err)
		}
		if err := cacheFlutterDeps(projectLocationAbs); err != nil {
			log.Warnf("Failed to collect flutter cache, error: %s", err)
		}
		if err := cacheShorebirdCLI(); err != nil {
			log.Warnf("Failed to collect shorebird CLI cache, error: %s", err)
		}
	}
}
