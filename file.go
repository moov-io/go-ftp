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

// File represents an fs.File object for a file on an FTP server.
type File struct {
	Filename string
	Contents io.ReadCloser

	// ModTime is the last modification time of the file. Open and Reader set
	// this from the server MDTM response when the server supports it. If the
	// server does not provide a time, ModTime is the current UTC time.
	ModTime time.Time

	fileinfo fs.FileInfo

	cleanup func() error
}

var _ fs.File = (&File{})

func (f *File) Close() error {
	if f == nil {
		return nil
	}
	if f.Contents != nil {
		if err := f.Contents.Close(); err != nil {
			return err
		}
	}
	if f.cleanup != nil {
		if err := f.cleanup(); err != nil {
			return err
		}
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

// fileInfo implements fs.FileInfo for a file on an FTP server.
type fileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
}

func (fi fileInfo) Name() string       { return fi.name }
func (fi fileInfo) Size() int64        { return fi.size }
func (fi fileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi fileInfo) ModTime() time.Time { return fi.modTime }
func (fi fileInfo) IsDir() bool        { return fi.mode.IsDir() }
func (fi fileInfo) Sys() any           { return nil }
