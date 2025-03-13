///////go:build integration

package go_ftp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"testing"
	"time"
)

func TestIntegrationUploadFile(t *testing.T) {
	t.Run("UploadFile with CreateUploadDirectories=true should create directories", func(t *testing.T) {
		ftpHost, ftpTerm := setupFTPServer(t)
		defer ftpTerm()

		// Create a new FTP client
		client, err := NewClient(ClientConfig{
			Hostname: ftpHost,
			Username: "ftpuser",
			Password: "ftppass",
			Timeout:  10 * time.Second,
			//
			CreateUploadDirectories: true,
			//
			DisableEPSV: false,
			CAFile:      "",
			TLSConfig:   nil,
		})
		require.NoError(t, err)

		// Upload a file
		err = client.UploadFile("2025/04/03/test.txt", io.NopCloser(bytes.NewReader([]byte("hello world"))))
		require.NoError(t, err)
	})
}

func setupFTPServer(t *testing.T) (string, func()) {
	t.Helper()

	ctx := context.TODO()

	req := testcontainers.ContainerRequest{
		Image:        "fauria/vsftpd:latest",
		ExposedPorts: []string{"20/tcp", "21/tcp", "21100-21110/tcp"},
		WaitingFor:   wait.ForListeningPort("21/tcp"),
		Env: map[string]string{
			"FTP_USER": "ftpuser",
			"FTP_PASS": "ftppass",
		},
	}

	ftpContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get the host and port of the running container
	host, err := ftpContainer.Host(ctx)
	require.NoError(t, err)

	port, err := ftpContainer.MappedPort(ctx, "21")
	require.NoError(t, err)

	ftpHost := fmt.Sprintf("%s:%s", host, port.Port())
	ftpTerm := func() { ftpContainer.Terminate(ctx) }

	return ftpHost, ftpTerm
}
