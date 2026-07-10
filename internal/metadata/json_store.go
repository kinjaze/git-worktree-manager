package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type JSONStore struct {
	path string
}

func NewJSONStore(path string) JSONStore {
	return JSONStore{path: path}
}

func (s JSONStore) Load() (File, error) {
	content, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return EmptyFile(), nil
	}
	if err != nil {
		return EmptyFile(), err
	}
	file := EmptyFile()
	if err := json.Unmarshal(content, &file); err != nil {
		return EmptyFile(), fmt.Errorf("metadata corrupt: %w", err)
	}
	if file.SchemaVersion == 0 {
		file.SchemaVersion = SchemaVersion
	}
	if file.SchemaVersion != SchemaVersion {
		return EmptyFile(), fmt.Errorf("unsupported metadata schema version: %d", file.SchemaVersion)
	}
	if file.Worktrees == nil {
		file.Worktrees = []Record{}
	}
	return file, nil
}

func (s JSONStore) Save(file File) error {
	file.SchemaVersion = SchemaVersion
	if file.Worktrees == nil {
		file.Worktrees = []Record{}
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	lock, err := AcquireLock(s.path)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Release() }()
	content, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, append(content, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s JSONStore) Upsert(record Record) error {
	file, err := s.Load()
	if err != nil {
		return err
	}
	for i := range file.Worktrees {
		if file.Worktrees[i].ID == record.ID {
			file.Worktrees[i] = record
			return s.Save(file)
		}
	}
	file.Worktrees = append(file.Worktrees, record)
	return s.Save(file)
}

func (s JSONStore) Remove(id string) error {
	file, err := s.Load()
	if err != nil {
		return err
	}
	filtered := file.Worktrees[:0]
	for _, record := range file.Worktrees {
		if record.ID != id {
			filtered = append(filtered, record)
		}
	}
	file.Worktrees = filtered
	return s.Save(file)
}
