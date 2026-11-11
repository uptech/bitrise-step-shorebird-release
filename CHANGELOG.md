# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - 2.0.0

Addresses the Bitrise StepLib review feedback on the 1.0.0 submission
([bitrise-steplib#5223](https://github.com/bitrise-io/bitrise-steplib/pull/5223)). Major version bump
because inputs were renamed/removed.

### Changed
- **Breaking:** renamed list inputs to use the `_list` suffix, per Bitrise's step input guidelines:
  `dart_defines` -> `dart_define_list`, `dart_define_files` -> `dart_define_file_list`,
  `android_target_platform` -> `android_target_platform_list`.
- **Breaking:** `shorebird_token` no longer defaults to `$SHOREBIRD_TOKEN` - it's now empty and must be set
  explicitly (e.g. to `$SHOREBIRD_TOKEN`).
- Step title is now "Create Shorebird Release".
- `flutter_version` is no longer marked required, since leaving it empty auto-detects the version.

### Added
- `deps` block declaring `curl` and `git` (brew and apt_get).
- `SHOREBIRD_CLI_REVISION` output: the git revision of the Shorebird CLI used, for use in Save Cache keys
  so the `~/.shorebird` cache is refreshed whenever Shorebird updates.
- A Shorebird CLI restored from cache is now upgraded (`shorebird upgrade`) before use, so cached installs
  pick up new Shorebird releases like a fresh install does.

### Removed
- **Breaking:** `cache_level` input and all built-in (legacy) caching. Use Bitrise's key-based caching
  Steps instead (Restore/Save Dart, Gradle, Cocoapods Cache, and Restore/Save Cache for `~/.shorebird`) -
  see the "Caching" section of the Step description.

## [1.0.0] - 2026-09-16

Initial release. Builds and publishes a new Shorebird release (`shorebird release`) for iOS and/or
Android. Split out of the combined Shorebird Build step into a dedicated, single-purpose step closely
mirroring [docs.shorebird.dev/code-push/release](https://docs.shorebird.dev/code-push/release/).

### Added
- `platform` (`both`/`ios`/`android`), `flutter_version` (auto-detects from the installed system Flutter
  if left empty), `flavor`, `target`, `build_name`, `build_number`, `dry_run`, `obfuscate`,
  `split_debug_info` inputs, mirroring the docs' Options table.
- `ios_codesign`, `ios_codesign_identity`, `ios_export_method`, `ios_export_options_plist` inputs,
  mirroring the docs' iOS-specific options.
- `android_output_type` (maps to `--artifact`) and `android_target_platform` (maps to `--target-platform`,
  for building a smaller APK for one architecture).
- `dart_defines` / `dart_define_files` inputs for `--dart-define` / `--dart-define-from-file`.
- `ios_additional_params` / `android_additional_params` (forwarded to the underlying `flutter build` after
  a literal `--`) and a generic `additional_params` escape hatch for rarely-needed flags.
- Release Flutter version auto-sync: fails fast with a clear error if Flutter isn't installed on `PATH`
  and no explicit `flutter_version` was set.
- Shorebird CLI install-or-restore handling (`~/.shorebird`), with a `shorebird.yaml` pre-flight check that
  fails with a clear, actionable error message if it's missing from the project root.
- iOS codesign-identity handling.
- Artifact export: `BITRISE_APK_PATH(_LIST)`, `BITRISE_AAB_PATH(_LIST)`, `BITRISE_IPA_PATH`.
- Dependency caching (`cache_level: all`): pub packages, Android (Gradle), Cocoapods/Carthage, and the
  installed Shorebird CLI itself.
- Best-effort `SHOREBIRD_RELEASE_VERSION` output, parsed from the Shorebird CLI's stdout/stderr (it has no
  machine-readable output for `release`).

[1.0.0]: https://github.com/uptech/bitrise-step-shorebird-release/releases/tag/1.0.0
