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
