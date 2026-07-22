package cache

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	stat, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) || stat.Size() == 0 {
		return false
	}

	return true
}

func CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer CloseWithErrorLog(d)

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func CloseWithErrorLog(c io.Closer) {
	if err := c.Close(); err != nil {
		slog.Error("failed to close resource", "Error", err)
	}
}
