// Copyright 2023 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package go_ftp

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
)

type ClientConfig struct {
	Hostname string
	Username string
	Password string

	Timeout     time.Duration
	DisableEPSV bool
	CAFile      string

	TLSConfig *tls.Config
}

type Client interface {
	Ping() error
	Close() error

	Open(path string) (*File, error)
	Reader(path string) (*File, error)

	Delete(path string) error
	UploadFile(path string, contents io.ReadCloser) error

	ListFiles(dir string) ([]string, error)
	Walk(dir string, fn fs.WalkDirFunc) error
}

func NewClient(cfg ClientConfig) (Client, error) {
	cc := &client{
		cfg: cfg,
	}

	// if err := rejectOutboundIPRange(cfg.SplitAllowedIPs(), cfg.FTP.Hostname); err != nil {
	// 	return nil, fmt.Errorf("ftp: %s is not whitelisted: %v", cfg.FTP.Hostname, err)
	// }

	_, err := cc.connection() // initial connection
	if err != nil {
		return cc, fmt.Errorf("ftp connect: %v", err)
	}
	return cc, nil
}

type client struct {
	conn *ftp.ServerConn
	cfg  ClientConfig
	mu   sync.Mutex // protects all read/write methods
}

// connection returns an ftp.ServerConn which is connected to the remote server.
// This function will attempt to establish a new connection if none exists already.
//
// connection must be called within a mutex lock as the underlying FTP client is not
// goroutine-safe.
func (cc *client) connection() (*ftp.ServerConn, error) {
	if cc == nil {
		return nil, errors.New("nil client or config")
	}

	if cc.conn != nil {
		// Verify the connection works and f not drop through and reconnect
		if err := cc.conn.NoOp(); err == nil {
			return cc.conn, nil
		} else {
			// Our connection is having issues, so retry connecting
			cc.conn.Quit()
		}
	}

	// Setup our FTP connection
	opts := []ftp.DialOption{
		ftp.DialWithTimeout(cc.cfg.Timeout),
		ftp.DialWithDisabledEPSV(cc.cfg.DisableEPSV),
	}
	tlsOpt, err := tlsDialOption(cc.cfg.TLSConfig, cc.cfg.CAFile)
	if err != nil {
		return nil, err
	}
	if tlsOpt != nil {
		opts = append(opts, *tlsOpt)
	}

	// Make the first connection
	conn, err := ftp.Dial(cc.cfg.Hostname, opts...)
	if err != nil {
		return nil, err
	}
	if err := conn.Login(cc.cfg.Username, cc.cfg.Password); err != nil {
		return nil, err
	}
	cc.conn = conn

	return cc.conn, nil
}

func tlsDialOption(conf *tls.Config, caFilePath string) (*ftp.DialOption, error) {
	if caFilePath == "" {
		return nil, nil
	}
	bs, err := os.ReadFile(caFilePath)
	if err != nil {
		return nil, fmt.Errorf("tlsDialOption: failed to read %s: %v", caFilePath, err)
	}
	pool, err := x509.SystemCertPool()
	if pool == nil || err != nil {
		pool = x509.NewCertPool()
	}
	ok := pool.AppendCertsFromPEM(bs)
	if !ok {
		return nil, fmt.Errorf("tlsDialOption: problem with AppendCertsFromPEM from %s", caFilePath)
	}
	if conf == nil {
		conf = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	conf.RootCAs = pool

	opt := ftp.DialWithTLS(conf)
	return &opt, nil
}

func (cc *client) Ping() error {
	if cc == nil {
		return errors.New("nil FTP client")
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()

	conn, err := cc.connection()
	if err != nil {
		return err
	}

	return conn.NoOp()
}

func (cc *client) Close() error {
	if cc == nil || cc.conn == nil {
		return nil
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()

	conn, err := cc.connection()
	if err != nil {
		return err
	}
	return conn.Quit()
}

// Open will return the contents at path and consume the entire file contents.
// WARNING: This method can use a lot of memory by consuming the entire file into memory.
func (cc *client) Open(path string) (*File, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	resp, err := cc.openFilepath(path)
	if err != nil {
		return nil, err
	}
	data, err := readResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("reading %s failed: %w", path, err)
	}
	return &File{
		Filename: filepath.Base(path),
		Contents: data,
	}, nil
}

// Reader will open the file at path and provide a reader to access its contents.
// Callers need to close the returned Contents
//
// Callers should be aware that network errors while reading can occur since contents
// are streamed from the FTP server. Having multiple open readers may not be supported.
func (cc *client) Reader(path string) (*File, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	file, err := cc.openFilepath(path)
	if err != nil {
		return nil, err
	}
	return &File{
		Filename: filepath.Base(path),
		Contents: file,
	}, nil
}

func (cc *client) Delete(path string) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if path == "" || strings.HasSuffix(path, "/") {
		return fmt.Errorf("FTP client: invalid path %v", path)
	}

	conn, err := cc.connection()
	if err != nil {
		return err
	}

	err = conn.Delete(path)
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		return err
	}
	return nil
}

// uploadFile saves the content of File at the given filename in the OutboundPath directory
//
// The File's contents will always be closed
func (cc *client) UploadFile(path string, contents io.ReadCloser) (err error) {
	defer contents.Close()

	cc.mu.Lock()
	defer cc.mu.Unlock()

	conn, err := cc.connection()
	if err != nil {
		return err
	}

	dir, filename := filepath.Split(path)
	if dir != "" {
		// Jump to previous directory after command is done
		wd, err := conn.CurrentDir()
		if err != nil {
			return err
		}
		defer func(previous string) {
			// Return to our previous directory when initially called
			if cleanupErr := conn.ChangeDir(previous); cleanupErr != nil {
				err = fmt.Errorf("FTP: problem uploading %s: %w", filename, cleanupErr)
			}
		}(wd)

		// Move into directory to run the command
		if err := conn.ChangeDir(dir); err != nil {
			return err
		}
	}

	// Write file contents into path
	// Take the base of f.Filename and our (out of band) OutboundPath to avoid accepting a write like '../../../../etc/passwd'.
	return conn.Stor(filename, contents)
}

// ListFiles will return the paths of files within dir. Paths are returned as locations from dir,
// so if dir is an absolute path the returned paths will be.
//
// Paths are matched in case-insensitive comparisons, but results are returned exactly as they
// appear on the server.
func (c *client) ListFiles(dir string) ([]string, error) {
	pattern := filepath.Clean(strings.TrimPrefix(dir, string(os.PathSeparator)))
	switch {
	case dir == "/":
		pattern = "*"
	case pattern == ".":
		if dir == "" {
			pattern = "*"
		} else {
			pattern = filepath.Join(dir, "*")
		}
	case pattern != "":
		pattern += "/*"
	}

	var filenames []string
	err := c.Walk(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Check if the server's path matches what we're searching for in a case-insensitive comparison.
		matches, err := filepath.Match(strings.ToLower(pattern), strings.ToLower(path))
		if matches && err == nil {
			// Return the path with exactly the case on the server.
			idx := strings.Index(strings.ToLower(path), strings.ToLower(strings.TrimSuffix(pattern, "*")))
			if idx > -1 {
				path = path[idx:]
				if strings.HasPrefix(dir, "/") && !strings.HasPrefix(path, "/") {
					path = "/" + path
				}
				filenames = append(filenames, path)
			} else {
				// Fallback to Go logic of presenting the path
				filenames = append(filenames, filepath.Join(dir, filepath.Base(path)))
			}
		}
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("listing %s failed: %w", dir, err)
	}
	return filenames, nil
}

// Walk will traverse dir and call fs.WalkDirFunc on each entry.
//
// Follow the docs for fs.WalkDirFunc for details on traversal. Walk accepts fs.SkipDir to not process directories.
func (cc *client) Walk(dir string, fn fs.WalkDirFunc) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	conn, err := cc.connection()
	if err != nil {
		return err
	}

	if dir != "" && dir != "." {
		// Jump to previous directory after command is done
		wd, err := conn.CurrentDir()
		if err != nil {
			return err
		}
		defer func(previous string) {
			// Return to our previous directory when initially called
			if cleanupErr := conn.ChangeDir(previous); cleanupErr != nil {
				err = fmt.Errorf("FTP: problem walking %s: %w", dir, cleanupErr)
			}
		}(wd)

		// Move into directory to run the command
		if err := conn.ChangeDir(dir); err != nil {
			return err
		}
	}

	// Setup a Walker for each file
	walker := conn.Walk(dir)
	for walker.Next() {
		entry := Entry{
			fd: walker.Stat(),
		}
		if entry.IsDir() {
			continue
		}
		err = fn(walker.Path(), entry, walker.Err())
		if err != nil {
			if err == fs.SkipDir {
				walker.SkipDir()
			} else {
				return fmt.Errorf("walking %s failed: %w", walker.Path(), err)
			}
		}
	}
	return nil
}

func (cc *client) openFilepath(path string) (resp *ftp.Response, err error) {
	conn, err := cc.connection()
	if err != nil {
		return nil, err
	}

	dir, filename := filepath.Split(path)
	if dir != "" {
		// Jump to previous directory after command is done
		wd, err := conn.CurrentDir()
		if err != nil {
			return nil, err
		}
		defer func(previous string) {
			// Return to our previous directory when initially called
			if cleanupErr := conn.ChangeDir(previous); cleanupErr != nil {
				err = fmt.Errorf("FTP: problem with readFiles: %w", cleanupErr)
			}
		}(wd)

		// Move into directory to run the command
		if err := conn.ChangeDir(dir); err != nil {
			return nil, err
		}
	}

	resp, err = conn.Retr(filename)
	if err != nil {
		return nil, fmt.Errorf("retrieving %s failed: %w", path, err)
	}

	return resp, nil
}

func readResponse(resp *ftp.Response) (io.ReadCloser, error) {
	defer resp.Close()

	var buf bytes.Buffer
	n, err := io.Copy(&buf, resp)
	// If there was nothing downloaded and no error then assume it's a directory.
	//
	// The FTP client doesn't have a STAT command, so we can't quite ensure this
	// was a directory.
	//
	// See https://github.com/moovfinancial/paygate/issues/494
	if n == 0 && err == nil {
		return io.NopCloser(&buf), nil
	}
	if err != nil {
		return nil, fmt.Errorf("n=%d error=%v", n, err)
	}
	return io.NopCloser(&buf), nil
}
