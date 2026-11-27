## Unreleased

IMPROVEMENTS

- docs: document the test FTP server, add upload and download examples, and correct the Go version and links
- docs: move Open and Reader notes onto the Client interface
- test: skip network tests when -short is set or when the FTP server is not reachable
- fix: set File.ModTime and Stat from the server MDTM and SIZE replies
- fix: align MockClient with the real client for missing directories, missing files, and Walk
- docs: add CONTRIBUTING.md and record the v0.4.1 release

## v0.4.1 (Released 2026-11-23)

This release updates dependencies, the Go toolchain, and CI workflows. There are no API changes.

BUILD

- chore(deps): update the Go toolchain to 1.27.4
- chore(deps): update module golang.org/x/sync to v0.25.0
- chore(deps): update GitHub Actions workflows

## v0.4.0 (Released 2025-01-23)

This release of moov-io/go-ftp changes `Walk` to pass directories to `fs.WalkDirFunc`. Previous versions incorrectly mishandled `fs.SkipDir` and did not provide directories to callers.

IMPROVEMENTS

- feat(readme): add file upload example (#17)
- fix: respect SkipDir to bypass directories
- test: verify deleting missing files isn't an error

BUILD

- fix(deps): update module github.com/stretchr/testify to v1.10.0 (#18)
- chore(deps): update github/codeql-action action to v3
- chore(deps): update actions/setup-go action to v5
- chore(deps): update actions/checkout action to v4

## v0.3.2 (Released 2023-11-08)

IMPROVEMENTS

- fix: allow Reader to stream larger files
- test: verify .Open can be used concurrently
- test: verify clients reconnect if used after Close()

## v0.3.1 (Released 2023-08-28)

IMPROVEMENTS

- fix: check walker error before stat calls

## v0.3.0 (Released 2023-08-23)

IMPROVEMENTS

- fix: clarify ListFiles returns case intensive matches but returns server case
- test: add "list and read" checks

## v0.2.0 (Released 2023-08-15)

IMPROVEMENTS

- fix: return full paths which matches moov-io/go-sftp behavior

## v0.1.1 (Released 2023-08-14)

IMPROVEMENTS

- docs: mention that we're having issues with multiple open readers
- fix: implement ListFiles with Walk to avoid some error cases
- fix: treat forced-root paths the same as relative in ListFiles
- test: copy over tlsDialOption tests from achgateway

BUILD

- chore(deps): update moov/fsftp docker tag to v0.2.2
- fix(deps): update module github.com/stretchr/testify to v1.8.4

## v0.1.0 (Released 2023-08-08)

This is the initial release of code extracted from achgateway.
