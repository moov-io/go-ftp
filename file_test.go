// Copyright 2022 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package go_ftp

import (
	"io"
	"io/fs"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFile(t *testing.T) {
	var f *File
	err := f.Close()
	require.NoError(t, err)
	stat, err := f.Stat()
	require.ErrorIs(t, err, io.EOF)
	require.Nil(t, stat)
	n, err := f.Read(nil)
	require.ErrorIs(t, err, io.EOF)
	require.Equal(t, 0, n)

	f = &File{}
	err = f.Close()
	require.NoError(t, err)
	stat, err = f.Stat()
	require.NoError(t, err)
	require.Nil(t, stat)
	n, err = f.Read(nil)
	require.ErrorIs(t, err, io.EOF)
	require.Equal(t, 0, n)
}

func TestFileInfo(t *testing.T) {
	fi := fileInfo{
		name:    "returns.ach",
		size:    12,
		modTime: time.Date(2026, 11, 27, 16, 0, 0, 0, time.UTC),
	}
	require.Equal(t, "returns.ach", fi.Name())
	require.Equal(t, int64(12), fi.Size())
	require.Equal(t, fs.FileMode(0), fi.Mode())
	require.False(t, fi.IsDir())
	require.Nil(t, fi.Sys())

	f := &File{
		Filename: fi.Name(),
		ModTime:  fi.ModTime(),
		fileinfo: fi,
	}
	stat, err := f.Stat()
	require.NoError(t, err)
	require.Equal(t, fi.Name(), stat.Name())
	require.Equal(t, fi.Size(), stat.Size())
	require.Equal(t, fi.ModTime(), stat.ModTime())
}
