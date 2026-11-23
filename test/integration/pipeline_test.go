package integration

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"eigentsar/image/loader"
	"eigentsar/model/builder"
	"eigentsar/model/projection"
	"eigentsar/model/recognition"
	"eigentsar/model/updater"
	"eigentsar/storage/dataset"
	storage "eigentsar/storage/model"
)

// synth draws a subject-specific pattern (own stripe frequency/phase) plus seeded noise.
func synth(subject int, seed int64) []byte {
	rng := rand.New(rand.NewSource(seed))
	img := image.NewGray(image.Rect(0, 0, 90, 110))
	for y := 0; y < 110; y++ {
		for x := 0; x < 90; x++ {
			v := 128 + 90*sinish(float64(x*(subject+1))/9+float64(y*(3-subject%3))/13) + rng.NormFloat64()*8
			img.SetGray(x, y, color.Gray{uint8(max(0, min(255, int(v))))})
		}
	}
	var b bytes.Buffer
	png.Encode(&b, img)
	return b.Bytes()
}

func sinish(x float64) float64 { // cheap deterministic wave, avoids importing math twice
	x -= float64(int(x/6.2832)) * 6.2832
	if x < 3.1416 {
		return 1 - 2*(x-1.5708)*(x-1.5708)/2.4674
	}
	return -(1 - 2*(x-4.7124)*(x-4.7124)/2.4674)
}

func setup(t *testing.T) (dataset.Store, string) {
	root := t.TempDir()
	for s := 0; s < 4; s++ {
		dir := filepath.Join(root, "train", fmt.Sprintf("s%02d", s))
		os.MkdirAll(dir, 0o755)
		for i := 0; i < 4; i++ {
			os.WriteFile(filepath.Join(dir, fmt.Sprintf("%02d.png", i)), synth(s, int64(s*100+i)), 0o644)
		}
	}
	return dataset.Store{Root: root}, filepath.Join(root, "models", "model.gob")
}

func TestBuildRecognizeAndPersist(t *testing.T) {
	store, mpath := setup(t)
	m, err := builder.Build(store, 5)
	if err != nil || m.K != 5 || len(m.Projections) != 16 {
		t.Fatalf("build: %v", err)
	}
	for s := 0; s < 4; s++ {
		img, _, _ := loader.Decode(synth(s, int64(9000+s)))
		q, err := projection.ProjectImage(m, img)
		if err != nil {
			t.Fatal(err)
		}
		res, _ := recognition.Nearest(m, q, 0)
		if res.Subject != fmt.Sprintf("s%02d", s) {
			t.Errorf("subject %d recognised as %s", s, res.Subject)
		}
		if rej, _ := recognition.Nearest(m, q, 1e-9); rej.Known {
			t.Error("tiny threshold must reject")
		}
	}
	// save/load round trip and compatibility
	if err := storage.Save(mpath, m); err != nil {
		t.Fatal(err)
	}
	loaded, err := storage.Load(mpath)
	if err != nil || storage.Compatible(loaded, 5, m.Fingerprint) != nil {
		t.Fatalf("load: %v", err)
	}
	if storage.Compatible(loaded, 6, m.Fingerprint) == nil || storage.Compatible(loaded, 5, "other") == nil {
		t.Fatal("incompatibility must be detected")
	}
	os.WriteFile(mpath, []byte("garbage"), 0o644)
	if _, err := storage.Load(mpath); err == nil {
		t.Fatal("corrupt artifact must error")
	}
}

func TestBuildErrors(t *testing.T) {
	store := dataset.Store{Root: t.TempDir()}
	if _, err := builder.Build(store, 3); err == nil {
		t.Fatal("empty dataset must error")
	}
	store, _ = setup(t)
	if _, err := builder.Build(store, 0); err == nil {
		t.Fatal("k=0 must error")
	}
	if _, err := builder.Build(store, 100); err == nil {
		t.Fatal("k beyond rank must error")
	}
}

func TestUpdaterAddAndRollback(t *testing.T) {
	store, mpath := setup(t)
	mgr := updater.New(store, mpath, 5)
	if err := mgr.Init(); err != nil {
		t.Fatal(err)
	}
	old, _ := mgr.Model()
	if _, err := mgr.AddImage("s09", []byte("not an image")); err == nil {
		t.Fatal("invalid image must be rejected")
	}
	if _, err := mgr.AddImage("../evil", synth(7, 1)); err == nil {
		t.Fatal("unsafe subject must be rejected")
	}
	if _, err := mgr.AddImage("s09", synth(7, 1)); err != nil {
		t.Fatal(err)
	}
	nm, _ := mgr.Model()
	if nm == old || len(nm.Labels) != 17 {
		t.Fatal("model must be rebuilt with the new image")
	}
	if _, err := mgr.AddImage("s09", synth(7, 1)); err == nil {
		t.Fatal("duplicate must be rejected")
	}
	// restart: saved model is reused only while the dataset is unchanged
	mgr2 := updater.New(store, mpath, 5)
	if err := mgr2.Init(); err != nil {
		t.Fatal(err)
	}
	if m2, _ := mgr2.Model(); !m2.BuiltAt.Equal(nm.BuiltAt) {
		t.Fatal("compatible artifact should have been loaded, not rebuilt")
	}
}
