# Contributing

Thanks for your interest in go-ftp. Feedback on the documentation, bug reports, and pull requests are all welcome.

Please review the [Code of Conduct](https://github.com/moov-io/moov-slack-code-of-conduct) to ensure you agree with the values of this project.

We use GitHub to manage reviews of pull requests.

* If you have a small fix or improvement, open a pull request and describe the change.
* If you plan a larger change, first propose your idea in a GitHub issue.
* Follow the [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) and the [Go Proverbs](https://go-proverbs.github.io/).

## Getting the code

This project uses [Go Modules](https://go.dev/blog/using-go-modules) and Go 1.26 or newer. See [Golang's install instructions](https://go.dev/doc/install) for help setting up Go. We recommend a [tagged release](https://github.com/moov-io/go-ftp/releases/latest) for production.

```
git clone https://github.com/moov-io/go-ftp.git
cd go-ftp
```

To send a patch, [fork the project](https://github.com/moov-io/go-ftp), add your fork as a remote, and push your branch to the fork.

## Testing

Unit tests and the mock client can be run without Docker:

```
go test -short ./...
```

The integration tests talk to an FTP server on `127.0.0.1:2121` with username `admin` and password `123456`. Start that server with Docker Compose:

```
make setup
go test ./...
make teardown
```

`make check` downloads the Moov lint script and needs network access. GitHub Actions runs that check on each pull request.

## Pull Requests

A good pull request is a complete change that adds value, with a title and summary that explain the work, and tests for the new behaviour. Keep the change small when you can. A change is complete when it can be merged.

## Secure Coding

Please read the [Web Application Secure Coding Practices](https://github.com/OWASP/Go-SCP) for Go.
