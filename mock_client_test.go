// Copyright 2022 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package go_ftp_test

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ftp "github.com/moov-io/go-ftp"

	"github.com/stretchr/testify/require"
)

func TestMockClient(t *testing.T) {
	client := ftp.NewMockClient(t)
	require.NoError(t, client.Ping())
	defer require.NoError(t, client.Close())

	_, err := client.Open("/missing.txt")
	require.Error(t, err)

	err = client.Delete("/missing.txt")
	require.NoError(t, err)

	body := io.NopCloser(strings.NewReader("contents"))
	err = client.UploadFile("/exists.txt", body)
	require.NoError(t, err)

	client.UploadFileErr = errors.New("bad error")
	err = client.UploadFile("/exists.txt", body)
	require.Error(t, err)

	paths, err := client.ListFiles("/")
	require.NoError(t, err)
	require.Len(t, paths, 1)
	require.Equal(t, "/exists.txt", paths[0])
}

func TestMockClient_ListAndOpenFiles(t *testing.T) {
	client := ftp.NewMockClient(t)
	require.NoError(t, client.Ping())
	defer require.NoError(t, client.Close())

	require.NoError(t, os.MkdirAll(filepath.Join(client.Dir(), "path"), 0o755))
	require.NoError(t, client.UploadFile("/path/f1.txt", io.NopCloser(strings.NewReader("foo"))))
	require.NoError(t, client.UploadFile("/path/f2.txt", io.NopCloser(strings.NewReader("foo"))))

	foundFiles, err := client.ListFiles("/path/")
	require.NoError(t, err)
	require.Len(t, foundFiles, 2)

	for _, file := range foundFiles {
		found, err := client.Open(file)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.False(t, found.ModTime.IsZero())
		stat, err := found.Stat()
		require.NoError(t, err)
		require.NotNil(t, stat)

		contents, err := io.ReadAll(found.Contents)
		require.NoError(t, err)
		require.Equal(t, "foo", string(contents))
		require.NoError(t, found.Close())

		// Consume the file with Reader
		found, err = client.Reader(file)
		require.NoError(t, err)
		require.NotNil(t, found)

		contents, err = io.ReadAll(found.Contents)
		require.NoError(t, err)
		require.Equal(t, "foo", string(contents))
		require.NoError(t, found.Close())
	}

	// Walk the directory and list files
	var walkedFiles []string
	err = client.Walk("/path/", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		walkedFiles = append(walkedFiles, path)
		return nil
	})
	require.NoError(t, err)
	require.Contains(t, walkedFiles, "f1.txt", "f2.txt")
}

func TestMockClient_UploadRequiresDirectory(t *testing.T) {
	client := ftp.NewMockClient(t)

	err := client.UploadFile("/dir/does/not/exist.txt", io.NopCloser(strings.NewReader("no data")))
	require.Error(t, err)
}

func TestMockClient_WalkMissingDir(t *testing.T) {
	client := ftp.NewMockClient(t)

	err := client.Walk("does/not/exist", func(path string, d fs.DirEntry, err error) error {
		return nil
	})
	require.Error(t, err)
}
