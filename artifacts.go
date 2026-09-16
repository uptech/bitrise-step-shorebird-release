package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/output"
	"github.com/bitrise-io/go-steputils/tools"
	"github.com/bitrise-io/go-utils/log"

	glob "github.com/ryanuber/go-glob"
)

// findArtifacts walks projectLocation looking for files matching any of outputPathPatterns (relative
// glob patterns, e.g. "*build/ios/ipa/*.ipa"). Unlike the reference Flutter Build step, every Shorebird
// release artifact (.ipa/.apk/.aab) is a file, never a directory, so this only ever looks for files.
func findArtifacts(projectLocation string, outputPathPatterns []string) ([]string, error) {
	var paths []string
	for _, pattern := range outputPathPatterns {
		found, err := findPaths(projectLocation, pattern)
		if err != nil {
			return nil, err
		}
		paths = append(paths, found...)
	}
	return paths, nil
}

func findPaths(location string, outputPathPattern string) (out []string, err error) {
	err = filepath.Walk(location, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || !glob.Glob(outputPathPattern, path) {
			return nil
		}
		out = append(out, path)
		return nil
	})
	if len(out) == 0 && err == nil {
		log.Debugf("couldn't find output artifact on path: " + filepath.Join(location, outputPathPattern))
	}
	return
}

// exportArtifacts exports the artifacts found for a release-mode build. Android artifacts (apk/aab) are
// exported to the standard multi-file + single-file Bitrise outputs, matching the reference Flutter Build
// step exactly. iOS always produces a single .ipa (already fully signed/exported by Shorebird), exported
// to the new BITRISE_IPA_PATH output.
func exportArtifacts(cfg config, spec platformSpec, artifacts []string) error {
	deployDir := os.Getenv("BITRISE_DEPLOY_DIR")

	if spec.platformArg == "ios" {
		return exportIOSIPA(artifacts, deployDir)
	}
	return exportAndroidArtifacts(cfg.AndroidOutputType, artifacts, deployDir)
}

func exportIOSIPA(artifacts []string, deployDir string) error {
	artifact := artifacts[len(artifacts)-1]
	if len(artifacts) > 1 {
		log.Warnf("- Multiple artifacts found: %v, exporting %s", artifacts, artifact)
	}

	deployedFilePath := filepath.Join(deployDir, filepath.Base(artifact))
	if err := output.ExportOutputFile(artifact, deployedFilePath, "BITRISE_IPA_PATH"); err != nil {
		return err
	}
	log.Donef("- BITRISE_IPA_PATH: %s", deployedFilePath)
	return nil
}

func exportAndroidArtifacts(androidOutputType string, artifacts []string, deployDir string) error {
	artifacts = filterAndroidArtifactsBy(androidOutputType, artifacts)

	singleFileOutputEnvName := "BITRISE_APK_PATH"
	multipleFileOutputEnvName := "BITRISE_APK_PATH_LIST"
	if androidOutputType == "appbundle" {
		singleFileOutputEnvName = "BITRISE_AAB_PATH"
		multipleFileOutputEnvName = "BITRISE_AAB_PATH_LIST"
	}

	var deployedFiles []string
	for _, artifactPath := range artifacts {
		deployedFilePath := filepath.Join(deployDir, filepath.Base(artifactPath))
		if err := output.ExportOutputFile(artifactPath, deployedFilePath, singleFileOutputEnvName); err != nil {
			return err
		}
		deployedFiles = append(deployedFiles, deployedFilePath)
	}
	if err := tools.ExportEnvironmentWithEnvman(multipleFileOutputEnvName, strings.Join(deployedFiles, "\n")); err != nil {
		return err
	}

	deployedSingleFile := ""
	if len(deployedFiles) > 0 {
		deployedSingleFile = deployedFiles[len(deployedFiles)-1]
	}
	log.Donef("- %s: %s", singleFileOutputEnvName, deployedSingleFile)
	log.Donef("- %s: %s", multipleFileOutputEnvName, strings.Join(deployedFiles, "|"))
	return nil
}

// filterAndroidArtifactsBy keeps only artifacts matching the requested extension, mirroring the
// reference step's behavior of exporting only the selected Android output type even when the output
// pattern would otherwise match both apk and aab files.
func filterAndroidArtifactsBy(androidOutputType string, artifacts []string) []string {
	wantExt := ".apk"
	if androidOutputType == "appbundle" {
		wantExt = ".aab"
	}

	var index int
	for _, artifact := range artifacts {
		if path.Ext(artifact) != wantExt {
			log.Debugf("Artifact (%s) found by output patterns, but it's not the selected output type (%s) - Skip", artifact, androidOutputType)
			continue
		}
		artifacts[index] = artifact
		index++
	}
	return artifacts[:index]
}
