// Copyright 2022 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package go_ftp_test

import (
	"testing"

	go_ftp "github.com/moov-io/go-ftp"

	"github.com/stretchr/testify/require"
)

func TestNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag was provided")
	}

	client, err := go_ftp.NewClient(go_ftp.ClientConfig{
		Hostname: "127.0.0.1:2121",
		Username: "admin",
		Password: "123456",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	t.Run("Ping", func(t *testing.T) {
		require.NoError(t, client.Ping())
	})

	t.Run("Read after Closing", func(t *testing.T) {
		// Close the connection but have the caller try without knowing it's closed
		require.NoError(t, client.Close())

		files, err := client.ListFiles("/")
		require.NoError(t, err)
		require.Greater(t, len(files), 0)

		// close it again for fun
		require.NoError(t, client.Close())

		// try again
		files, err = client.ListFiles("/")
		require.NoError(t, err)
		require.Greater(t, len(files), 0)
	})
}
