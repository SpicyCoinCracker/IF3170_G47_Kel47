package model

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"eigentsar/model/builder"
)

func Save(path string, m *builder.Model) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "model-*.tmp")
	if err != nil {
		return err
	}
	err = gob.NewEncoder(tmp).Encode(m)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp.Name(), path)
	}
	if err != nil {
		os.Remove(tmp.Name())
	}
	return err
}

func Load(path string) (*builder.Model, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var m builder.Model
	if err := gob.NewDecoder(f).Decode(&m); err != nil {
		return nil, fmt.Errorf("model artifact %s corrupt (%v)", path, err)
	}
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("model artifact %s tidak dapat dipakai: %w", path, err)
	}
	return &m, nil
}

func Compatible(m *builder.Model, k int, fingerprint string) error {
	if m.K != k {
		return fmt.Errorf("saved model memiliki k=%d, konfigurasi k=%d", m.K, k)
	}
	if m.Fingerprint != fingerprint {
		return fmt.Errorf("dataset training telah diubah")
	}
	return nil
}
