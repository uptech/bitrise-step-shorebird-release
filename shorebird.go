package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
)

const shorebirdInstallScript = `curl --proto '=https' --tlsv1.2 -sSf https://raw.githubusercontent.com/shorebirdtech/install/main/install.sh | bash`

func shorebirdBinPath() string {
	return filepath.Join(os.Getenv("HOME"), ".shorebird", "bin", "shorebird")
}

// ensureShorebirdCLI returns a path to a working shorebird binary, either already on PATH, restored
// from a Bitrise cache (~/.shorebird, when cache_level: all restored it from a previous run), or freshly
// installed via Shorebird's official install script.
func ensureShorebirdCLI() (string, error) {
	if p, err := exec.LookPath("shorebird"); err == nil {
		log.Donef("- Using shorebird CLI already on PATH: %s", p)
		return p, nil
	}

	binPath := shorebirdBinPath()
	if fi, err := os.Stat(binPath); err == nil && !fi.IsDir() {
		log.Donef("- Using cached shorebird CLI at %s", binPath)
		return binPath, nil
	}

	log.Printf("Shorebird CLI not found, installing...")
	cmd := command.New("bash", "-c", shorebirdInstallScript).SetStdout(os.Stdout).SetStderr(os.Stderr)
	log.Donef("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("install shorebird CLI: %w", err)
	}

	if _, err := os.Stat(binPath); err != nil {
		return "", fmt.Errorf("shorebird binary not found at %s after install: %w", binPath, err)
	}
	log.Donef("- Installed shorebird CLI at %s", binPath)
	return binPath, nil
}

// verifyShorebirdProject checks that a shorebird.yaml exists at the project root, failing fast with a
// clear, actionable error message otherwise (shorebird release would fail anyway, but with a less helpful
// message).
func verifyShorebirdProject(projectLocation string) error {
	shorebirdYamlPath := filepath.Join(projectLocation, "shorebird.yaml")
	if _, err := os.Stat(shorebirdYamlPath); err != nil {
		return fmt.Errorf("shorebird.yaml not found in %s. Run `shorebird init` locally and commit the generated shorebird.yaml to your repo before using this Step", projectLocation)
	}
	return nil
}

type flutterVersionMachine struct {
	FrameworkVersion string `json:"frameworkVersion"`
}

// detectFlutterVersion returns the version of the Flutter SDK already installed on PATH, so a release
// build can be pinned to it via --flutter-version. Fails with a clear, actionable error message if
// Flutter isn't installed at all.
func detectFlutterVersion() (string, error) {
	if _, err := exec.LookPath("flutter"); err != nil {
		return "", fmt.Errorf("Flutter is not installed / not on PATH. This step requires Flutter to already be installed (e.g. via the Flutter Installer step or FVM) before it runs - it MUST be set up in a prior step")
	}

	out, err := command.New("flutter", "--version", "--machine").RunAndReturnTrimmedOutput()
	if err != nil {
		return "", fmt.Errorf("run `flutter --version --machine`: %w", err)
	}

	var v flutterVersionMachine
	if err := json.Unmarshal([]byte(out), &v); err != nil || v.FrameworkVersion == "" {
		return "", fmt.Errorf("parse `flutter --version --machine` output: %w", err)
	}
	return v.FrameworkVersion, nil
}

// runShorebirdCommand runs the shorebird CLI with the given args, streaming stdout/stderr live to the
// console while also capturing both streams into the returned string, so it can be scraped for outputs
// later. Both streams are captured (not just stdout) because Shorebird's mason_logger-based progress
// output - including the "Release Version: X" line - is written to stderr, not stdout.
func runShorebirdCommand(shorebirdBin, dir string, args []string) (string, error) {
	var buf bytes.Buffer
	cmd := command.New(shorebirdBin, args...).
		SetDir(dir).
		SetStdout(io.MultiWriter(os.Stdout, &buf)).
		SetStderr(io.MultiWriter(os.Stderr, &buf))

	log.Donef("$ %s", cmd.PrintableCommandArgs())
	err := cmd.Run()
	return buf.String(), err
}

var (
	ansiRe           = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	releaseVersionRe = regexp.MustCompile(`(?i)Release Version:\s*(\S+)`)
)

// parseShorebirdOutput best-effort scrapes a release version out of the Shorebird CLI's human-readable
// stdout/stderr. Shorebird has no machine-readable --json output for `release`, so this is inherently
// fragile - callers must treat an empty result as "not found", never as an error.
func parseShorebirdOutput(stdout string) (releaseVersion string) {
	clean := ansiRe.ReplaceAllString(stdout, "")
	if m := releaseVersionRe.FindStringSubmatch(clean); len(m) == 2 {
		releaseVersion = strings.TrimSpace(m[1])
	}
	return
}
