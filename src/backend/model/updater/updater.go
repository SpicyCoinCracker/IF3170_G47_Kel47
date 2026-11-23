
package updater

import (
	"errors"
	"log"
	"sync"

	"eigentsar/image/loader"
	"eigentsar/model/builder"
	"eigentsar/storage/dataset"
	storage "eigentsar/storage/model"
)

var ErrNotReady = errors.New("model perlu di-build dahulu (train dataset tidak ada, terlalu kecill, atau k invalid)")

type Status struct {
	Ready            bool    `json:"ready"`
	Building         bool    `json:"building"`
	K                int     `json:"k"`
	Rank             int     `json:"rank"`
	Width            int     `json:"width"`
	Height           int     `json:"height"`
	TrainImages      int     `json:"trainImages"`
	Subjects         int     `json:"subjects"`
	VarianceRetained float64 `json:"varianceRetained"`
	BuiltAt          string  `json:"builtAt,omitempty"`
	BuildMillis      int64   `json:"buildMillis"`
	ConfiguredK      int     `json:"configuredK"`
	LastError        string  `json:"lastError,omitempty"`
	SaveError        string  `json:"saveError,omitempty"`
}

type Manager struct {
	Store     dataset.Store
	ModelPath string

	buildMu  sync.Mutex 
	mu       sync.RWMutex
	cur      *builder.Model
	k        int
	building bool
	lastErr  string
	saveErr  string
}

func New(store dataset.Store, modelPath string, k int) *Manager {
	return &Manager{Store: store, ModelPath: modelPath, k: k}
}

func (m *Manager) Init() error {
	m.buildMu.Lock()
	defer m.buildMu.Unlock()
	if entries, err := m.Store.List(dataset.Train); err == nil {
		if fp, err := m.Store.Fingerprint(entries); err == nil {
			saved, err := storage.Load(m.ModelPath)
			if err == nil {
				err = storage.Compatible(saved, m.k, fp)
			}
			if err == nil {
				m.set(saved)
				return nil
			}
			log.Printf("tidak menggunakan saved model: %v", err)
		}
	}
	return m.rebuildLocked(m.k)
}

func (m *Manager) Rebuild(k int) error {
	m.buildMu.Lock()
	defer m.buildMu.Unlock()
	return m.rebuildLocked(k)
}

func (m *Manager) AddImage(subject string, data []byte) (dataset.Entry, error) {
	_, format, err := loader.Decode(data)
	if err != nil {
		return dataset.Entry{}, err
	}
	ext := ".jpg"
	if format == "png" {
		ext = ".png"
	}
	m.buildMu.Lock()
	defer m.buildMu.Unlock()
	e, err := m.Store.Add(subject, ext, data)
	if err != nil {
		return e, err
	}
	if err := m.rebuildLocked(0); err != nil {
		m.Store.Remove(e)
		return e, err
	}
	return e, nil
}

func (m *Manager) rebuildLocked(k int) error {
	if k <= 0 {
		k = m.k
	}
	m.mu.Lock()
	m.building = true
	m.mu.Unlock()
	nm, err := builder.Build(m.Store, k)
	var saveErr error
	if err == nil {
		saveErr = storage.Save(m.ModelPath, nm)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.building = false
	if err != nil {
		m.lastErr = err.Error()
		return err
	}
	m.cur, m.k, m.lastErr, m.saveErr = nm, k, "", ""
	if saveErr != nil {
		m.saveErr = saveErr.Error()
		log.Printf("model artifact tidak bisa di-save: %v", saveErr)
	}
	return nil
}

func (m *Manager) set(nm *builder.Model) {
	m.mu.Lock()
	m.cur = nm
	m.mu.Unlock()
}

func (m *Manager) Model() (*builder.Model, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.cur == nil {
		return nil, ErrNotReady
	}
	return m.cur, nil
}

func (m *Manager) Status() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := Status{Building: m.building, ConfiguredK: m.k, LastError: m.lastErr, SaveError: m.saveErr}
	if c := m.cur; c != nil {
		subj := map[string]bool{}
		for _, l := range c.Labels {
			subj[l] = true
		}
		s.Ready, s.K, s.Rank, s.Width, s.Height = true, c.K, c.Rank, c.Width, c.Height
		s.TrainImages, s.Subjects, s.VarianceRetained = len(c.Labels), len(subj), c.VarianceRetained()
		s.BuiltAt, s.BuildMillis = c.BuiltAt.Format("2006-01-02T15:04:05Z07:00"), c.BuildMillis
	}
	return s
}
