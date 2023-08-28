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
