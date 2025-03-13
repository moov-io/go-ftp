package go_ftp_test

import (
	"bytes"
	"io"
	"testing"

	go_ftp "github.com/moov-io/go-ftp"

	"github.com/stretchr/testify/require"
)

func TestIntegration_fauria_vsftpd(t *testing.T) {
	config := go_ftp.ClientConfig{
		Hostname: "localhost:4021",
		Username: "ftpuser",
		Password: "ftppass",

		CreateUploadDirectories: true,
	}
	client, err := go_ftp.NewClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	t.Run("UploadFile with CreateUploadDirectories", func(t *testing.T) {
		err := client.UploadFile("2025/test.txt", io.NopCloser(bytes.NewReader([]byte("hello world"))))
		require.NoError(t, err)

		file, err := client.Open("2025/04/03/test.txt")
		require.NoError(t, err)

		bs, err := io.ReadAll(file)
		require.NoError(t, err)
		require.Equal(t, "hello world", string(bs))
	})
}
