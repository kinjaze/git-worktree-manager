package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Store struct {
	path string
}

func NewStore(path string) Store {
	return Store{path: path}
}

func (s Store) Load() (Config, error) {
	if s.path == "" {
		path, err := DefaultConfigPath()
		if err != nil {
			return Default(), err
		}
		s.path = path
	}
	content, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Default(), err
	}
	cfg := Default()
	if err := json.Unmarshal(content, &cfg); err != nil {
		return Default(), err
	}
	cfg.Language = NormalizeLanguage(cfg.Language)
	return cfg, nil
}

func (s Store) Save(cfg Config) error {
	if s.path == "" {
		path, err := DefaultConfigPath()
		if err != nil {
			return err
		}
		s.path = path
	}
	cfg.Language = NormalizeLanguage(cfg.Language)
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, append(content, '\n'), 0o644)
}
