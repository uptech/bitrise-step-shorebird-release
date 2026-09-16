package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"

	"github.com/bitrise-io/go-xcode/certificateutil"
	shellquote "github.com/kballard/go-shellquote"
)

const codesignField = "ios-signing-cert"
const noCodesignFlag = "--no-codesign"

var flutterConfigPath = filepath.Join(os.Getenv("HOME"), ".flutter_settings")

// prepareIOSCodesign mirrors the reference Flutter Build step's codesign-identity handling: it lets the
// user override which installed codesigning identity Flutter (and therefore Shorebird, which shells out
// to Flutter under the hood) should use, by writing it into ~/.flutter_settings.
func prepareIOSCodesign(cfg config) error {
	fmt.Println()
	log.Infof("iOS Codesign settings")

	iosParams, err := shellquote.Split(cfg.IOSAdditionalParams)
	if err != nil {
		return fmt.Errorf("failed to parse iOS additional parameters: %w", err)
	}
	if slices.Contains(iosParams, noCodesignFlag) {
		log.Printf(" - Skipping codesign preparation, %s parameter set", noCodesignFlag)
		return nil
	}
	if !cfg.IOSCodesign {
		log.Printf(" - Skipping codesign preparation, ios_codesign is false")
		return nil
	}

	log.Printf(" Installed codesign identities:")
	installedCertificates, err := certificateutil.InstalledCodesigningCertificateNames()
	if err != nil {
		return fmt.Errorf("failed to fetch installed codesign identities: %w", err)
	}
	for _, identity := range installedCertificates {
		log.Printf(" - %s", identity)
	}

	if len(installedCertificates) == 0 {
		return fmt.Errorf("no codesign identities installed")
	}

	var flutterSettings map[string]string
	flutterSettingsExists, err := pathutil.IsPathExists(flutterConfigPath)
	if err != nil {
		return fmt.Errorf("failed to check if %s exists: %w", flutterConfigPath, err)
	}
	if flutterSettingsExists {
		flutterSettingsContent, err := fileutil.ReadBytesFromFile(flutterConfigPath)
		if err != nil {
			return fmt.Errorf("error while reading %s: %w", flutterConfigPath, err)
		}
		if err := json.Unmarshal(flutterSettingsContent, &flutterSettings); err != nil {
			return fmt.Errorf("failed to parse .flutter_settings file: %w", err)
		}
	} else {
		flutterSettings = map[string]string{}
	}

	if cfg.IOSCodesignIdentity != "" {
		log.Warnf(" Override codesign identity:")
		log.Printf(" - Store: %s", cfg.IOSCodesignIdentity)
		if !slices.Contains(installedCertificates, cfg.IOSCodesignIdentity) {
			return fmt.Errorf("the selected identity %q is not installed on the system", cfg.IOSCodesignIdentity)
		}
		flutterSettings[codesignField] = cfg.IOSCodesignIdentity
		newSettingsContent, err := json.MarshalIndent(flutterSettings, "", " ")
		if err != nil {
			return fmt.Errorf("failed to marshal .flutter_settings file: %w", err)
		}
		if err := fileutil.WriteBytesToFile(flutterConfigPath, newSettingsContent); err != nil {
			return fmt.Errorf("error while writing .flutter_settings file: %w", err)
		}
		log.Donef(" - Done")
		return nil
	}

	log.Printf(" Stored Flutter codesign settings:")
	storedIdentity, ok := flutterSettings[codesignField]
	if !ok {
		log.Printf(" - No codesign identity set")
	} else {
		log.Printf(" - %s", storedIdentity)
		if !slices.Contains(installedCertificates, storedIdentity) {
			return fmt.Errorf("identity %q is not installed on the system", storedIdentity)
		}
	}

	return nil
}
