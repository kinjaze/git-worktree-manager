package metadata

import (
	"errors"
	"os"
)

type Lock struct {
	path string
}

func AcquireLock(metadataPath string) (Lock, error) {
	lockPath := metadataPath + ".lock"
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return Lock{}, err
	}
	_ = file.Close()
	return Lock{path: lockPath}, nil
}

func (l Lock) Release() error {
	if l.path == "" {
		return errors.New("empty lock path")
	}
	return os.Remove(l.path)
}
