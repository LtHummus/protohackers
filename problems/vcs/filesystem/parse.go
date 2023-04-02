package filesystem

import (
	"errors"
	"strings"
)

var (
	ErrIllegalFilename = errors.New("illegal file name")
	ErrNotFound        = errors.New("not found")
)

func parseFilepath(path string) ([]string, error) {
	if path == "" {
		return nil, ErrIllegalFilename
	}

	parts := strings.Split(path, "/")
	if parts[0] != "" {
		return nil, ErrIllegalFilename
	}

	if len(parts) < 2 {
		return nil, ErrIllegalFilename
	}

	parts = parts[1:]
	if parts[0] == "" {
		return nil, nil
	}

	return parts, nil
}
