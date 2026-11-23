package integration

import (
	"encoding/json"
	"image/png"
	"net/http/httptest"
	"testing"

	"eigentsar/api"
	"eigentsar/model/updater"
)

func TestEigenfaceEndpoints(t *testing.T) {
	store, mpath := setup(t)
	mgr := updater.New(store, mpath, 4)
	if err := mgr.Init(); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer((&api.API{Mgr: mgr}).Routes())
	defer srv.Close()

	res, err := srv.Client().Get(srv.URL + "/api/model/eigenfaces")
	if err != nil || res.StatusCode != 200 {
		t.Fatalf("list: %v", err)
	}
	var d struct {
		K          int
		MeanUrl    string
		Components []struct {
			Singular float64
			Url      string
		}
	}
	json.NewDecoder(res.Body).Decode(&d)
	if d.K != 4 || len(d.Components) != 4 || d.Components[0].Singular < d.Components[3].Singular {
		t.Fatalf("bad listing %+v", d)
	}
	for _, u := range []string{d.MeanUrl, d.Components[0].Url} {
		r, _ := srv.Client().Get(srv.URL + u)
		img, err := png.Decode(r.Body)
		if err != nil || img.Bounds().Dx() != 64 || img.Bounds().Dy() != 80 {
			t.Fatalf("%s: %v", u, err)
		}
	}
	if r, _ := srv.Client().Get(srv.URL + "/api/model/eigenfaces/99.png"); r.StatusCode != 404 {
		t.Fatalf("want 404, got %d", r.StatusCode)
	}
}
