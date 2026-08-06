// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"errors"
	"os"
	"path/filepath"
)

// atomicReplaceFile installs tmpPath at targetPath after both files have been
// created in the same directory. os.Rename is atomic on supported production
// filesystems, but Windows can reject replacement when the target exists, so
// this helper retries after removing only the already-locked target file.
func atomicReplaceFile(tmpPath, targetPath string) error {
	if filepath.Dir(tmpPath) != filepath.Dir(targetPath) {
		return errors.New("atomic replace requires source and target in same directory")
	}
	if err := os.Rename(tmpPath, targetPath); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrExist) && !errors.Is(err, os.ErrPermission) {
		return err
	}
	if err := os.Remove(targetPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(tmpPath, targetPath)
}
