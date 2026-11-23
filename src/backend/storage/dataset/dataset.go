
package dataset

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	Train      = "train"
	Validation = "validation"
)

var (
	nameRe     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
	ErrInvalid = errors.New("invalid name")
	ErrExists  = errors.New("image sudah ada dalam dataset")
	imageExtOK = map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
)

type Entry struct{ Split, Subject, File string }

func (e Entry) ID() string { return e.Split + "/" + e.Subject + "/" + e.File }

type Store struct{ Root string }

func validSplit(s string) bool { return s == Train || s == Validation }

func (s Store) path(e Entry) string { return filepath.Join(s.Root, e.Split, e.Subject, e.File) }

func (s Store) Path(e Entry) string { return s.path(e) }

func (s Store) Resolve(split, subject, file string) (Entry, error) {
	e := Entry{split, subject, file}
	if !validSplit(split) || !nameRe.MatchString(subject) || !nameRe.MatchString(file) || !imageExtOK[strings.ToLower(filepath.Ext(file))] {
		return e, ErrInvalid
	}
	if st, err := os.Stat(s.path(e)); err != nil || !st.Mode().IsRegular() {
		return e, os.ErrNotExist
	}
	return e, nil
}

func (s Store) List(split string) ([]Entry, error) {
	if !validSplit(split) {
		return nil, ErrInvalid
	}
	subjects, err := os.ReadDir(filepath.Join(s.Root, split))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, sd := range subjects {
		if !sd.IsDir() || !nameRe.MatchString(sd.Name()) {
			continue
		}
		files, err := os.ReadDir(filepath.Join(s.Root, split, sd.Name()))
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if f.Type().IsRegular() && nameRe.MatchString(f.Name()) && imageExtOK[strings.ToLower(filepath.Ext(f.Name()))] {
				out = append(out, Entry{split, sd.Name(), f.Name()})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out, nil
}

func (s Store) Add(subject, ext string, data []byte) (Entry, error) {
	if !nameRe.MatchString(subject) || !imageExtOK[ext] {
		return Entry{}, fmt.Errorf("%w: subyek harus sesuai %s", ErrInvalid, nameRe)
	}
	sum := sha256.Sum256(data)
	e := Entry{Train, subject, hex.EncodeToString(sum[:8]) + ext}
	if err := os.MkdirAll(filepath.Dir(s.path(e)), 0o755); err != nil {
		return Entry{}, err
	}
	f, err := os.OpenFile(s.path(e), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return Entry{}, ErrExists
	}
	if err != nil {
		return Entry{}, err
	}
	_, werr := f.Write(data)
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		os.Remove(s.path(e))
		return Entry{}, werr
	}
	return e, nil
}


func (s Store) Remove(e Entry) {
	os.Remove(s.path(e))
	os.Remove(filepath.Dir(s.path(e))) 
}


func (s Store) Fingerprint(entries []Entry) (string, error) {
	h := sha256.New()
	for _, e := range entries {
		st, err := os.Stat(s.path(e))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s|%d|%d\n", e.ID(), st.Size(), st.ModTime().UnixNano())
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
