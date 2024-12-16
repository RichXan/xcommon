package xutil

import (
	"os"
	"path/filepath"
)

func CreateFile(filename string, mode os.FileMode) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func SetFileModeWithCreating(filename string, mode os.FileMode) error {
	info, err := os.Stat(filename)

	if os.IsNotExist(err) {
		f, crtErr := CreateFile(filename, mode)
		if crtErr != nil {
			return crtErr
		}
		return f.Close()
	} else if err != nil {
		return err
	} else if info.Mode() != mode {
		err = os.Chmod(filename, mode)
		return err
	}
	return nil
}
