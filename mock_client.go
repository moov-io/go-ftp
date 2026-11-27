// Copyright 2022 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package go_ftp

import (
	"cmp"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MockClient is an in-memory Client that stores files in a temporary directory.
// NewMockClient requires a *testing.T and is intended for tests.
//
// The mock matches the real client in these cases:
//   - UploadFile does not create missing parent directories
//   - Delete of a missing file returns nil
//   - Open and Reader set Filename, ModTime, and Stat
//   - Walk does not call fn for the root directory itself
type MockClient struct {
	root string

	// Err is an optional error which is returned for all methods unless a
	// method specific error (e.g. UploadFileErr) is defined
	Err error

	PingErr  error
	CloseErr error

	OpenErr   error
	ReaderErr error

	DeleteErr     error
	UploadFileErr error

	ListFilesErr error
	WalkErr      error
}

var _ Client = (&MockClient{})

func NewMockClient(t *testing.T) *MockClient {
	return &MockClient{
		root: t.TempDir(),
	}
}

func (c *MockClient) Ping() error {
	return cmp.Or(c.PingErr, c.Err)
}

func (c *MockClient) Dir() string {
	return c.root
}

func (c *MockClient) Close() error {
	return cmp.Or(c.CloseErr, c.Err)
}

func (c *MockClient) Reader(path string) (*File, error) {
	if c.Err != nil || c.ReaderErr != nil {
		return nil, cmp.Or(c.ReaderErr, c.Err)
	}
	return c.Open(path)
}

func (c *MockClient) Open(path string) (*File, error) {
	if c.Err != nil || c.OpenErr != nil {
		return nil, cmp.Or(c.OpenErr, c.Err)
	}
	file, err := os.Open(filepath.Join(c.root, path))
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	_, name := filepath.Split(path)
	return &File{
		Filename: name,
		Contents: file,
		ModTime:  info.ModTime(),
		fileinfo: info,
	}, nil
}

func (c *MockClient) Delete(path string) error {
	if c.Err != nil || c.DeleteErr != nil {
		return cmp.Or(c.DeleteErr, c.Err)
	}
	err := os.Remove(filepath.Join(c.root, path))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (c *MockClient) UploadFile(path string, contents io.ReadCloser) error {
	if contents != nil {
		defer contents.Close()
	}
	if c.Err != nil || c.UploadFileErr != nil {
		return cmp.Or(c.UploadFileErr, c.Err)
	}

	dir, _ := filepath.Split(path)
	if dir != "" && dir != string(os.PathSeparator) {
		if _, err := os.Stat(filepath.Join(c.root, dir)); err != nil {
			return err
		}
	}

	bs, err := io.ReadAll(contents)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(c.root, path), bs, 0600)
}

func (c *MockClient) ListFiles(dir string) ([]string, error) {
	if c.Err != nil || c.ListFilesErr != nil {
		return nil, cmp.Or(c.ListFilesErr, c.Err)
	}

	os.MkdirAll(filepath.Join(c.root, dir), 0777)

	fds, err := os.ReadDir(filepath.Join(c.root, dir))
	if err != nil {
		return nil, err
	}
	var out []string
	for i := range fds {
		fd := filepath.Join(dir, strings.TrimPrefix(fds[i].Name(), c.root))
		out = append(out, fd)
	}
	return out, nil
}

func (c *MockClient) Walk(dir string, fn fs.WalkDirFunc) error {
	if c.Err != nil || c.WalkErr != nil {
		return cmp.Or(c.WalkErr, c.Err)
	}

	d, err := filepath.Abs(filepath.Join(c.root, dir))
	if err != nil {
		return err
	}
	if _, err := os.Stat(d); err != nil {
		return err
	}

	// Skip the walk root itself so the mock matches the real client.
	return fs.WalkDir(os.DirFS(d), ".", func(path string, entry fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		return fn(path, entry, err)
	})
}
