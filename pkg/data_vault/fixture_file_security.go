package data_vault

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/virtengine/virtengine/pkg/data_vault/internal/pathnorm"
)

var errFixtureStoreInUse = errors.New("fixture store is already open by another process or instance")

func rejectFixtureSymlink(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	for current := abs; ; current = filepath.Dir(current) {
		info, statErr := os.Lstat(current)
		if statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("fixture path ancestor must not be a symlink: %s", current)
			}
			if runtime.GOOS == "windows" {
				longPath, longErr := pathnorm.LongPath(current)
				if longErr != nil {
					return longErr
				}
				resolved, resolveErr := filepath.EvalSymlinks(current)
				if resolveErr != nil {
					return resolveErr
				}
				resolvedAbs, resolveErr := filepath.Abs(resolved)
				if resolveErr != nil {
					return resolveErr
				}
				longAbs, longErr := filepath.Abs(longPath)
				if longErr != nil {
					return longErr
				}
				// EvalSymlinks resolves reparse points but also expands 8.3
				// short names, so compare against the long-name spelling.
				// Otherwise a legitimate short-name ancestor such as
				// C:\Users\RUNNER~1\AppData\Local\Temp (the %TEMP% GitHub's
				// Windows runners hand out) is misreported as a reparse point.
				if !strings.EqualFold(filepath.Clean(resolvedAbs), filepath.Clean(longAbs)) {
					return fmt.Errorf("fixture path ancestor is a reparse point: %s", current)
				}
			}
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return statErr
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	return nil
}

func enforceFixturePathSecurity(path string, directory bool, options FixtureSecurityOptions) error {
	if err := rejectFixtureSymlink(path); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		if !options.UnsafeWindowsDevelopment {
			return errors.New("fixture filesystem cannot enforce safe Windows ACLs; UnsafeWindowsDevelopment is required")
		}
		return nil
	}
	mode := os.FileMode(0o600)
	if directory {
		mode = 0o700
	}
	if err := os.Chmod(path, mode); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
