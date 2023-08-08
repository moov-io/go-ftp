// Copyright 2022 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package go_ftp

import (
	"io"
	"io/fs"
	"time"

	"github.com/jlaffaye/ftp"
)

// File represents a fs.File object of a location on a SFTP server.
type File struct {
	Filename string
	Contents io.ReadCloser

	// ModTime is a timestamp of when the last modification occurred
	// to this file. The default will be the current UTC time.
	ModTime time.Time

	fileinfo fs.FileInfo
}

var _ fs.File = (&File{})

func (f *File) Close() error {
	if f == nil {
		return nil
	}
	if f.Contents != nil {
		return f.Contents.Close()
	}
	return nil
}

func (f *File) Stat() (fs.FileInfo, error) {
	if f == nil {
		return nil, io.EOF
	}
	return f.fileinfo, nil
}

func (f *File) Read(buf []byte) (int, error) {
	if f == nil || f.Contents == nil {
		return 0, io.EOF
	}
	return f.Contents.Read(buf)
}

// Entry implements fs.DirEntry
type Entry struct {
	fd *ftp.Entry
}

var _ fs.DirEntry = (&Entry{})

func (e Entry) Name() string {
	return e.fd.Name
}

func (e Entry) IsDir() bool {
	return e.fd.Type == ftp.EntryTypeFolder
}

// Type only returns fs.ModeDir or fs.ModeSymlink
func (e Entry) Type() fs.FileMode {
	switch e.fd.Type {
	case ftp.EntryTypeFile:
		// TODO(adam):
	case ftp.EntryTypeFolder:
		return fs.ModeDir
	case ftp.EntryTypeLink:
		return fs.ModeSymlink
	}
	return fs.ModeIrregular
}

func (e Entry) Info() (fs.FileInfo, error) {
	return nil, nil
}
