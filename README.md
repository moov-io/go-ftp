[![Moov Banner Logo](https://user-images.githubusercontent.com/20115216/104214617-885b3c80-53ec-11eb-8ce0-9fc745fb5bfc.png)](https://github.com/moov-io)

<p align="center">
  <a href="https://slack.moov.io/">Community</a>
  ·
  <a href="https://moov.io/blog/">Blog</a>
  <br>
</p>

[![GoDoc](https://pkg.go.dev/badge/github.com/moov-io/go-ftp)](https://pkg.go.dev/github.com/moov-io/go-ftp)
[![Build Status](https://github.com/moov-io/go-ftp/workflows/Go/badge.svg)](https://github.com/moov-io/go-ftp/actions)
[![Coverage Status](https://codecov.io/gh/moov-io/go-ftp/branch/master/graph/badge.svg)](https://codecov.io/gh/moov-io/go-ftp)
[![Go Report Card](https://goreportcard.com/badge/github.com/moov-io/go-ftp)](https://goreportcard.com/report/github.com/moov-io/go-ftp)
[![Repo Size](https://img.shields.io/github/languages/code-size/moov-io/go-ftp?label=project%20size)](https://github.com/moov-io/go-ftp)
[![Apache 2 License](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/moov-io/go-ftp/master/LICENSE)
[![Slack Channel](https://slack.moov.io/badge.svg?bg=e01563&fgColor=fffff)](https://slack.moov.io/)
[![GitHub Stars](https://img.shields.io/github/stars/moov-io/go-ftp)](https://github.com/moov-io/go-ftp)
[![Twitter](https://img.shields.io/twitter/follow/moov?style=social)](https://twitter.com/moov?lang=en)

# moov-io/go-ftp

Moov's mission is to give developers an easy way to create and integrate bank processing into their own software products. Our open source projects are each focused on solving a single responsibility in financial services and designed around performance, scalability, and ease of use.

Go FTP provides a simple FTP [client interface](https://pkg.go.dev/github.com/moov-io/go-ftp#Client) for uploading, listing, and opening files on an FTP server.

This project requires **Go 1.26 or higher**. The current `go.mod` also sets `toolchain go1.27.4`.

```go
type Client interface {
	Ping() error
	Close() error

	Open(path string) (*File, error)
	Reader(path string) (*File, error)

	Delete(path string) error
	UploadFile(path string, contents io.ReadCloser) error

	ListFiles(dir string) ([]string, error)
	Walk(dir string, fn fs.WalkDirFunc) error
}
```

`Open` reads the entire file into memory. Prefer `Reader` when you want to stream a large file. The client does not support more than one open `Reader` at a time.

`ListFiles` returns the files in `dir` (not subdirectories; use `Walk` for a recursive listing). The returned paths use the same style as `dir`: `ListFiles("returns")` returns `"returns/RET1.ach"`, and `ListFiles("/returns")` returns `"/returns/RET1.ach"`. Use the returned path with `Open`, `Reader`, `UploadFile`, and `Delete`.

`UploadFile` writes to an existing directory. It does not create missing parent directories, and it always closes the provided `io.ReadCloser`.

The library also includes a [mock client implementation](https://pkg.go.dev/github.com/moov-io/go-ftp#MockClient) which uses a local filesystem temporary directory for testing. `NewMockClient` requires a `*testing.T` and is intended for tests. The mock matches the real client for `UploadFile` (the parent directory must exist), `ModTime` / `Stat` on `Open` and `Reader`, and `Walk` (the root directory itself is not passed to the callback). `Delete` of a missing file returns nil on the mock.

## Upload example

Set `Timeout` so `NewClient` does not wait without a limit when the host is unreachable. The destination directory must already exist on the server.

```go
package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	ftp "github.com/moov-io/go-ftp"
)

func main() {
	client, err := ftp.NewClient(ftp.ClientConfig{
		Hostname: "127.0.0.1:2121",
		Username: "admin",
		Password: "123456",
		Timeout:  10 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if err := client.Ping(); err != nil {
		log.Fatal(err)
	}

	fileData, err := os.Open("file.txt")
	if err != nil {
		log.Fatal(err)
	}

	// "archive" must already exist on the server.
	dest := filepath.Join("archive", "file.txt")
	if err := client.UploadFile(dest, fileData); err != nil {
		log.Fatal(err)
	}
}
```

## Download, archive, and delete example

This example lists files, reads each file, copies it to an archive directory, and then deletes the original. The archive directory must already exist.

```go
package main

import (
	"bytes"
	"io"
	"log"
	"path/filepath"
	"time"

	ftp "github.com/moov-io/go-ftp"
)

func main() {
	client, err := ftp.NewClient(ftp.ClientConfig{
		Hostname: "127.0.0.1:2121",
		Username: "admin",
		Password: "123456",
		Timeout:  10 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	files, err := client.ListFiles("returns")
	if err != nil {
		log.Fatal(err)
	}

	for _, path := range files {
		file, err := client.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			log.Fatal(err)
		}

		archived := filepath.Join("archive", filepath.Base(path))
		if err := client.UploadFile(archived, io.NopCloser(bytes.NewReader(data))); err != nil {
			log.Fatal(err)
		}
		if err := client.Delete(path); err != nil {
			log.Fatal(err)
		}
	}
}
```

To stream a large file instead of loading it into memory, call `Reader` in place of `Open` and close the returned `File` after you read it.

## Project status

Moov Go FTP is actively used in production environments. Please star the project if you are interested in its progress. Please let us know if you encounter any bugs/unclear documentation or have feature suggestions by opening up an issue or pull request. Thanks!

## Getting help

| channel                                                     | info                                                                                                                                    |
|-------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| [Project Documentation](https://pkg.go.dev/github.com/moov-io/go-ftp) | Package docs on pkg.go.dev.                                                                                             |
| Twitter [@moov](https://twitter.com/moov)	                  | You can follow Moov.io's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories. |
| [GitHub Issue](https://github.com/moov-io/go-ftp/issues)   | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.                     |
| [moov-io slack](https://slack.moov.io/)                     | Join our slack channel to have an interactive discussion about the development of the project.                                          |

## Supported and tested platforms

- 64-bit Linux (Ubuntu, Debian), macOS, and Windows

## Contributing

Yes please! Please review our [Contributing guide](CONTRIBUTING.md) and [Code of Conduct](https://github.com/moov-io/moov-slack-code-of-conduct) to get started!

This project uses [Go Modules](https://go.dev/blog/using-go-modules) and Go 1.26 or newer. See [Golang's install instructions](https://go.dev/doc/install) for help setting up Go. You can download the source code and we offer [tagged and released versions](https://github.com/moov-io/go-ftp/releases/latest) as well. We highly recommend you use a tagged release for production.

### Releasing

To make a release of go-ftp, open a pull request that updates `CHANGELOG.md` with the next version number and details. Then push the tag (for example `git push origin v1.0.0`) to origin so that CI can make the release.

### Testing

Unit tests and the mock client can be run without Docker:

```
go test -short ./...
```

The tests in `client_test.go` and `network_test.go` talk to an FTP server on `127.0.0.1:2121` with username `admin` and password `123456`. Docker and Docker Compose are required for that path. The tests skip when you pass `-short`, and they also skip when the server is not reachable.

```
make setup    # docker compose up -d --force-recreate --remove-orphans
go test ./...
make teardown
```

`make setup` starts the `moov/fsftp:v0.2.2` image from `docker-compose.yml`. The server listens on port 2121 and uses passive ports 30000-30009. It mounts `testdata/ftp-server` as the server root.

`make check` downloads the Moov lint script and needs network access. Current overall coverage can be found on [Codecov](https://app.codecov.io/gh/moov-io/go-ftp/).

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.
