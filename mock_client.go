// Copyright 2022 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package go_ftp

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type MockClient struct {
	root string

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
	return c.PingErr
}

func (c *MockClient) Dir() string {
	return c.root
}

func (c *MockClient) Close() error {
	return c.CloseErr
}

func (c *MockClient) Reader(path string) (*File, error) {
	if c.ReaderErr != nil {
		return nil, c.ReaderErr
	}
	return c.Open(path)
}

func (c *MockClient) Open(path string) (*File, error) {
	if c.OpenErr != nil {
		return nil, c.OpenErr
	}
	file, err := os.Open(filepath.Join(c.root, path))
	if err != nil {
		return nil, err
	}
	_, name := filepath.Split(path)
	return &File{
		Filename: name,
		Contents: file,
	}, nil
}

func (c *MockClient) Delete(path string) error {
	if c.DeleteErr != nil {
		return c.DeleteErr
	}
	return os.Remove(filepath.Join(c.root, path))
}

func (c *MockClient) UploadFile(path string, contents io.ReadCloser) error {
	if c.UploadFileErr != nil {
		return c.UploadFileErr
	}

	dir, _ := filepath.Split(path)
	if err := os.MkdirAll(filepath.Join(c.root, dir), 0777); err != nil {
		return err
	}

	bs, _ := io.ReadAll(contents)

	return os.WriteFile(filepath.Join(c.root, path), bs, 0600)
}

func (c *MockClient) ListFiles(dir string) ([]string, error) {
	if c.ListFilesErr != nil {
		return nil, c.ListFilesErr
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
	if c.WalkErr != nil {
		return c.WalkErr
	}

	d, err := filepath.Abs(filepath.Join(c.root, dir))
	if err != nil {
		return err
	}
	os.MkdirAll(d, 0777)

	return fs.WalkDir(os.DirFS(d), ".", fn)
}
