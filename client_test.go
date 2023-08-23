// Copyright 2023 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package go_ftp_test

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	go_ftp "github.com/moov-io/go-ftp"
	mhttptest "github.com/moov-io/go-ftp/internal/httptest"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	client, err := go_ftp.NewClient(go_ftp.ClientConfig{
		Hostname: "127.0.0.1:2121",
		Username: "admin",
		Password: "123456",
	})
	require.NotNil(t, client)
	require.NoError(t, err)

	require.NoError(t, client.Ping())
	defer client.Close()

	t.Run("open", func(t *testing.T) {
		file, err := client.Open("first.txt")
		require.NoError(t, err)
		t.Cleanup(func() { file.Close() })

		var buf bytes.Buffer
		io.Copy(&buf, file)
		require.Equal(t, "hello world", strings.TrimSpace(buf.String()))
	})

	t.Run("reader", func(t *testing.T) {
		file, err := client.Reader("archive/old.txt")
		require.NoError(t, err)
		t.Cleanup(func() { file.Close() })

		var buf bytes.Buffer
		io.Copy(&buf, file)
		require.Equal(t, "previous data", strings.TrimSpace(buf.String()))
	})

	t.Run("upload and delete", func(t *testing.T) {
		body := io.NopCloser(strings.NewReader("example data"))
		err := client.UploadFile("new.txt", body)
		require.NoError(t, err)

		file, err := client.Open("new.txt")
		require.NoError(t, err)

		var buf bytes.Buffer
		io.Copy(&buf, file)
		require.Equal(t, "example data", strings.TrimSpace(buf.String()))
		require.NoError(t, file.Close())

		err = client.Delete("new.txt")
		require.NoError(t, err)

		file, err = client.Open("new.txt")
		require.Nil(t, file)
		require.ErrorContains(t, err, "retrieving new.txt failed: 551 File not available")
	})

	t.Run("list", func(t *testing.T) {
		filenames, err := client.ListFiles(".")
		require.NoError(t, err)
		require.ElementsMatch(t, filenames, []string{"first.txt", "second.txt", "empty.txt"})

		filenames, err = client.ListFiles("/")
		require.NoError(t, err)
		require.ElementsMatch(t, filenames, []string{"/first.txt", "/second.txt", "/empty.txt"})
	})

	t.Run("list subdir", func(t *testing.T) {
		filenames, err := client.ListFiles("archive")
		require.NoError(t, err)
		require.ElementsMatch(t, filenames, []string{"archive/old.txt", "archive/empty2.txt"})

		filenames, err = client.ListFiles("/archive")
		require.NoError(t, err)
		require.ElementsMatch(t, filenames, []string{"/archive/old.txt", "/archive/empty2.txt"})

		filenames, err = client.ListFiles("/archive/")
		require.NoError(t, err)
		require.ElementsMatch(t, filenames, []string{"/archive/old.txt", "/archive/empty2.txt"})
	})

	t.Run("list and read", func(t *testing.T) {
		filenames, err := client.ListFiles("/with-empty")
		require.NoError(t, err)

		// randomize filename order
		rand.Shuffle(len(filenames), func(i, j int) {
			filenames[i], filenames[j] = filenames[j], filenames[i]
		})
		require.ElementsMatch(t, filenames, []string{
			"/with-empty/EMPTY1.txt", "/with-empty/empty_file2.txt",
			"/with-empty/data.txt", "/with-empty/data2.txt",
		})

		// read each file and get back expected contents
		var contents []string
		for i := range filenames {
			var file *go_ftp.File
			if i/2 == 0 {
				file, err = client.Open(filenames[i])
			} else {
				file, err = client.Reader(filenames[i])
			}
			require.NoError(t, err, fmt.Sprintf("filenames[%d]", i))
			require.NotNil(t, file, fmt.Sprintf("filenames[%d]", i))
			require.NotNil(t, file.Contents, fmt.Sprintf("filenames[%d]", i))

			bs, err := io.ReadAll(file.Contents)
			require.NoError(t, err)

			contents = append(contents, string(bs))
		}

		require.ElementsMatch(t, contents, []string{"", "", "also data\n", "has data\n"})
	})

	t.Run("list case testing", func(t *testing.T) {
		files, err := client.ListFiles("/upper")
		require.NoError(t, err)
		require.ElementsMatch(t, files, []string{"/Upper/names.txt"})

		files, err = client.ListFiles("ARCHIVE")
		require.NoError(t, err)
		require.ElementsMatch(t, files, []string{"archive/old.txt", "archive/empty2.txt"})
	})

	t.Run("walk", func(t *testing.T) {
		var found []string
		err := client.Walk(".", func(path string, d fs.DirEntry, err error) error {
			found = append(found, path)
			return nil
		})
		require.NoError(t, err)
		require.ElementsMatch(t, found, []string{
			"Upper/names.txt",
			"first.txt", "second.txt", "empty.txt",
			"archive/old.txt", "archive/empty2.txt",
			"with-empty/EMPTY1.txt", "with-empty/empty_file2.txt",
			"with-empty/data.txt", "with-empty/data2.txt",
		})
	})

	t.Run("walk subdir", func(t *testing.T) {
		var found []string
		err := client.Walk("/archive", func(path string, d fs.DirEntry, err error) error {
			found = append(found, path)
			return nil
		})
		require.NoError(t, err)
		require.ElementsMatch(t, found, []string{
			"/archive/old.txt", "/archive/empty2.txt",
		})
	})

	require.NoError(t, client.Close())
}

func TestClientErrors(t *testing.T) {
	client, err := go_ftp.NewClient(go_ftp.ClientConfig{
		Hostname: "127.0.0.1:2121",
		Username: "admin",
		Password: "123456",
	})
	require.NotNil(t, client)
	require.NoError(t, err)

	require.NoError(t, client.Ping())
	defer client.Close()

	t.Run("open", func(t *testing.T) {
		file, err := client.Open("not-found.txt")
		require.ErrorContains(t, err, "551 File not available")
		require.Nil(t, file)
	})

	t.Run("reader", func(t *testing.T) {
		file, err := client.Reader("not-found.txt")
		require.ErrorContains(t, err, "551 File not available")
		require.Nil(t, file)
	})

	t.Run("upload", func(t *testing.T) {
		body := io.NopCloser(strings.NewReader("no data"))
		err := client.UploadFile("dir/does/not/exist.txt", body)
		require.ErrorContains(t, err, "550 Directory change to /dir/does/not failed: lstat /data/dir/does/not: no such file or directory")
	})

	t.Run("list", func(t *testing.T) {
		filenames, err := client.ListFiles("does/not/exist")
		require.NoError(t, err)
		require.Len(t, filenames, 0)
	})

	t.Run("walk", func(t *testing.T) {
		var found []string
		err := client.Walk("does/not/exist", func(path string, d fs.DirEntry, err error) error {
			found = append(found, path)
			return nil
		})
		require.ErrorContains(t, err, "550 Directory change to /does/not/exist failed: lstat /data/does/not/exist: no such file or directory")
		require.Len(t, found, 0)
	})

	require.NoError(t, client.Close())
}

func TestClientFailure(t *testing.T) {
	client, err := go_ftp.NewClient(go_ftp.ClientConfig{
		Hostname: "127.0.0.1:2121",
		Username: "incorrect",
		Password: "wrong",
	})
	require.NotNil(t, client)
	require.ErrorContains(t, err, "ftp connect: 530 Incorrect password, not logged in")

	require.ErrorContains(t, client.Ping(), "530 Incorrect password, not logged in")
	require.NoError(t, client.Close())
}

func TestClient__tlsDialOption(t *testing.T) {
	if testing.Short() {
		return // skip network calls
	}

	cafile, err := mhttptest.GrabConnectionCertificates(t, "google.com:443")
	require.NoError(t, err)
	defer os.Remove(cafile)

	client, err := go_ftp.NewClient(go_ftp.ClientConfig{
		Hostname: "127.0.0.1:2121",
		Username: "admin",
		Password: "123456",
		Timeout:  5 * time.Second,
		CAFile:   cafile,
	})
	require.ErrorContains(t, err, "tls: first record does not look like a TLS handshake")
	require.NotNil(t, client)
	require.NoError(t, client.Close())
}
