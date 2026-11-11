# Create Shorebird Release

[![Step changelog](https://shields.io/github/v/release/uptech/bitrise-step-shorebird-release?include_prereleases&label=changelog&color=blueviolet)](https://github.com/uptech/bitrise-step-shorebird-release/releases)

Builds and publishes a new Shorebird release for iOS and/or Android.

<details>
<summary>Description</summary>

Runs [`shorebird release`](https://docs.shorebird.dev/code-push/release/) - a drop-in replacement for
`flutter build` that compiles your app and uploads its Dart artifacts to Shorebird's servers as the
baseline future patches diff against. You still need to upload the generated artifact
(`.ipa`/`.apk`/`.aab`) to the App Store / Play Store yourself (or via a follow-up deploy Step) - Shorebird
doesn't do that part.

### Flutter version

Shorebird maintains its own private Flutter cache, independent of any system-installed Flutter. Leave
**Flutter version** empty to auto-detect the version of the system Flutter already on `PATH` (via
`flutter --version --machine`) and pass it through, keeping the Shorebird-built binary in lockstep with
whatever Flutter version the rest of your workflow uses - or set it explicitly to pin a version. If
Flutter isn't installed/found on `PATH` at all and no version is set explicitly, the Step fails with a
clear error - a Flutter Install (or FVM) Step must run before this one in that case.

### Releasing for multiple platforms

Set **Platform** to `both` to release iOS and Android in one Step run - each platform is built and
published as its own `shorebird release` invocation, registered together as one version release.
Releasing for iOS (or both iOS and Android together) requires a macOS runner.

### Caching

This Step doesn't cache anything itself. Use Bitrise's
[key-based caching](https://docs.bitrise.io/en/bitrise-ci/dependencies-and-caching/dependencies-and-caching-overview#key-based-caching)
Steps around it instead:
- **Restore Dart Cache** / **Save Dart Cache** for pub packages
- **Restore Gradle Cache** / **Save Gradle Cache** for Android builds
- **Restore Cocoapods Cache** / **Save Cocoapods Cache** for iOS builds
- **Restore Cache** / **Save Cache** for the Shorebird CLI itself (`~/.shorebird`)

When `~/.shorebird` is restored, the Step runs `shorebird upgrade` instead of re-running the Shorebird
install script, and exports the installed CLI's git revision as `SHOREBIRD_CLI_REVISION`. Include it in
the Save Cache key so the cache is refreshed whenever Shorebird updates:

```yaml
- restore-cache@3:
    inputs:
    - key: shorebird-{{ .OS }}-{{ .Arch }}-
- shorebird-release@2: {}
- save-cache@1:
    inputs:
    - key: shorebird-{{ .OS }}-{{ .Arch }}-{{ getenv "SHOREBIRD_CLI_REVISION" }}
    - paths: ~/.shorebird
    - is_key_unique: "true"
```

Restore Cache's key acts as a prefix, so it restores the most recently saved Shorebird cache.

### Building a smaller Android APK for one architecture

By default, three Android ABIs are compiled (`arm`, `arm64`, `x86_64`). Use **Android target platform(s)**
to restrict this to one or more specific architectures. Note this only applies to `release` - patches
never need it.

### Signing issues

If you hit iOS provisioning/certificate errors: for local builds, enable "Automatically manage signing"
in Xcode's Signing & Capabilities tab. For CI, set up an `ExportOptions.plist` describing your export
method, provisioning profiles, signing certificate, signing style, and team ID, and point **Export options
plist path** at it.

### Side-loading and MDM

Shorebird doesn't require App Store or Play Store publishing - it works with side-loading and MDM
distribution on Android (and enterprise iOS programs work similarly). Set **Android artifact type** to
`apk` to produce a directly installable `.apk` instead of an `.aab`.

### Managing releases after publishing

This Step only creates new releases. To list, inspect, or delete existing releases, use the
`shorebird releases` CLI commands directly, or the [Shorebird console](https://console.shorebird.dev).
Deleting a release from the console is irreversible and removes all of its patches.

### Configuring the Step
1. In **Project Location** the root directory of your Flutter project is automatically filled out. It must
   contain a `shorebird.yaml` (created by running `shorebird init` locally and committing the result) -
   this Step fails fast with a clear error if it's missing.
2. Set **Shorebird API token** to a Shorebird Console API key, stored as a Bitrise Secret.
3. Select **Platform**, leave **Flutter version** empty (or pin one explicitly).
4. For Android, pick **Android artifact type** and adjust **Android output pattern** if your build output
   goes somewhere non-default.
5. For iOS, configure **Codesign iOS build?**, **Export method**, and **Export options plist path** as
   needed, and adjust **iOS output pattern** if needed.

### Useful links
- [Shorebird](https://shorebird.dev)
- [Creating a release](https://docs.shorebird.dev/code-push/release/)
- [Managing releases](https://docs.shorebird.dev/code-push/release/#manage-releases)
- [Code signing](https://devcenter.bitrise.io/code-signing/code-signing-index/)

</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/steps/adding-steps-to-a-workflow.html).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `project_location` | The root dir of your Flutter project. Must contain a `shorebird.yaml` (created by running `shorebird init` locally and committing the result) - the Step fails with a clear error if it's missing. | required | `$BITRISE_SOURCE_DIR` |
| `shorebird_token` | Exported as the `SHOREBIRD_TOKEN` env var before invoking the Shorebird CLI. The Step fails fast if this is empty.  Create an API key in the Shorebird Console under Account -> API Keys, store it as a Bitrise Secret (mark it sensitive), and reference it here, e.g. `$SHOREBIRD_TOKEN`. | required, sensitive |  |
| `platform` | The selected platform will be released, or both iOS and Android if you select both. Releasing for iOS requires a macOS runner. | required | `both` |
| `flutter_version` | Maps to `--flutter-version`. Shorebird manages its own private Flutter cache under `~/.shorebird`, independent of any system-installed Flutter. Leaving this empty makes the Step auto-detect the version of the system Flutter already on `PATH` (via `flutter --version --machine`) and pass that value through. If Flutter isn't installed/found on `PATH` at all, the Step fails with a clear error - a Flutter Install (or FVM) Step **must** run before this Step in that case. |  |  |
| `flavor` | Maps to `--flavor`. |  |  |
| `target` | Maps to `-t`/`--target`, e.g. `lib/main_prod.dart`. |  |  |
| `build_name` | Maps to `--build-name`. |  |  |
| `build_number` | Maps to `--build-number`, e.g. `$BITRISE_BUILD_NUMBER`. |  |  |
| `dry_run` | Maps to `-n`/`--dry-run`. | required | `false` |
| `obfuscate` | Maps to `--obfuscate`. | required | `false` |
| `split_debug_info` | Maps to `--split-debug-info=<dir>`. |  |  |
| `dart_define_list` | Separate multiple defines with a newline, e.g.: ``` API_URL=https://example.com ENABLE_FOO=true ``` |  |  |
| `dart_define_file_list` | Separate multiple file paths with a newline. |  |  |
| `additional_params` | Appended directly to the `shorebird release` invocation. Use this for advanced/rarely-needed flags this Step doesn't have a dedicated input for, e.g. `--public-key-cmd=...`. |  |  |
| `ios_codesign` | Maps to --codesign (true) / --no-codesign (false) | required | `true` |
| `ios_codesign_identity` | Override codesign identity in .flutter_settings. |  |  |
| `ios_export_method` | Maps to `--export-method`. | required | `app-store` |
| `ios_export_options_plist` | Maps to `--export-options-plist`. See "Signing issues" above. |  |  |
| `ios_additional_params` | Raw `flutter build` args, appended after a literal `--` separator to the iOS invocation. |  | `--release` |
| `ios_output_pattern` | Separate multiple patterns with a newline. | required | `*build/ios/ipa/*.ipa` |
| `android_output_type` | apk maps to --artifact=apk (default), appbundle maps to --artifact=aab. | required | `apk` |
| `android_target_platform_list` | Maps to `--target-platform`. Separate multiple values with a newline. See "Building a smaller Android APK for one architecture" above. |  |  |
| `android_additional_params` | Raw `flutter build` args, appended after a literal `--` separator to the Android invocation. |  | `--release` |
| `android_output_pattern` | Separate multiple patterns with a newline. Only the selected **Android artifact type** is exported, even if the pattern matches both apk and aab files. | required | `*build/app/outputs/apk/*/*.apk *build/app/outputs/bundle/*/*.aab` |
| `is_debug_mode` | If debug mode is enabled, the step will print verbose logs | required | `false` |
</details>

<details>
<summary>Outputs</summary>

| Environment Variable | Description |
| --- | --- |
| `BITRISE_APK_PATH` | The created .apk file's path. |
| `BITRISE_APK_PATH_LIST` | All created .apk file paths, separated by \|. |
| `BITRISE_AAB_PATH` | Path of the generated (and copied) AAB file - after filtering. |
| `BITRISE_AAB_PATH_LIST` | List of the generated (and copied) AAB file paths - after filtering, separated by \|. |
| `BITRISE_IPA_PATH` | The created .ipa file's path. |
| `SHOREBIRD_CLI_REVISION` | Git revision of the Shorebird CLI used for the release. Changes with every Shorebird update, so use it in a Save Cache key for `~/.shorebird` - see "Caching" above. |
| `SHOREBIRD_RELEASE_VERSION` | The release version created. Best-effort parsed from the Shorebird CLI's stdout/stderr - may be empty if parsing fails, since Shorebird does not currently offer a machine-readable output format for `release`. |
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/uptech/bitrise-step-shorebird-release/pulls) and [issues](https://github.com/uptech/bitrise-step-shorebird-release/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://docs.bitrise.io/en/bitrise-ci/bitrise-cli/running-your-first-local-build-with-the-cli.html).

Learn more about developing steps:

- [Create your own step](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/developing-your-own-bitrise-step/developing-a-new-step.html)

Made with 💜 by Uptech Studio