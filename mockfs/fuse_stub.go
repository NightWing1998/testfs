//go:build !linux
// +build !linux

// Package mockfs provides a stubbed implementation on non-Linux platforms.
package mockfs

import (
	"context"
	"errors"
)

// Mount is not supported on non-Linux platforms. It returns an error to make
// this explicit to callers.
func (fs *MockFS) Mount(ctx context.Context, mountpoint string) error {
	return errors.New("mockfs fuse support is available only on linux")
}

// Close is a no-op on unsupported platforms.
func (fs *MockFS) Close() error {
	return nil
}
