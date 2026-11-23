// Package api holds the HTTP layer only: routing, parsing, status codes, JSON errors.
// All numerics/recognition are delegated to model/*, image/* and storage/*.
//
// Errors are always {"error": "<message>"}. Images are exposed via safe URLs
// /api/images/{split}/{subject}/{file}, never filesystem paths.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eigentsar/image/loader"
	"eigentsar/model/projection"
	"eigentsar/model/recognition"
	"eigentsar/model/updater"
	"eigentsar/storage/dataset"
)

const maxUpload = 8 << 20

type API struct{ Mgr *updater.Manager }

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/subjects", a.subjects)
	mux.HandleFunc("GET /api/images/{split}/{subject}/{file}", a.image)
	mux.HandleFunc("GET /api/model/status", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, a.Mgr.Status()) })
	mux.HandleFunc("POST /api/model/rebuild", a.rebuild)
	mux.HandleFunc("POST /api/recognize", a.recognize)
	mux.HandleFunc("POST /api/dataset/images", a.addImage)
	mux.HandleFunc("GET /api/evaluate", a.evaluate)
	mux.HandleFunc("GET /api/model/eigenfaces", a.eigenfaces)
	mux.HandleFunc("GET /api/model/eigenfaces/{n}", a.eigenfaceImage)
	return middleware(mux)
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		defer func() {
			if p := recover(); p != nil {
				log.Printf("panic on %s %s: %v", r.Method, r.URL.Path, p)
				writeErr(w, 500, errors.New("internal error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

type imageDTO struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func toImage(id string) imageDTO { return imageDTO{id, "/api/images/" + id} }

func (a *API) subjects(w http.ResponseWriter, r *http.Request) {
	entries, err := a.Mgr.Store.List(dataset.Train)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	type subject struct {
		SubjectID   string     `json:"subjectId"`
		SubjectName string     `json:"subjectName"`
		Images      []imageDTO `json:"images"`
	}
	out := []subject{}
	for _, e := range entries { 
		if n := len(out); n == 0 || out[n-1].SubjectID != e.Subject {
			out = append(out, subject{e.Subject, e.Subject, []imageDTO{}})
		}
		last := &out[len(out)-1]
		last.Images = append(last.Images, toImage(e.ID()))
	}
	writeJSON(w, 200, out)
}

func (a *API) image(w http.ResponseWriter, r *http.Request) {
	e, err := a.Mgr.Store.Resolve(r.PathValue("split"), r.PathValue("subject"), r.PathValue("file"))
	switch {
	case errors.Is(err, dataset.ErrInvalid):
		writeErr(w, 400, errors.New("invalid image path"))
	case err != nil:
		writeErr(w, 404, errors.New("image not found"))
	default:
		http.ServeFile(w, r, a.Mgr.Store.Path(e))
	}
}


func readUpload(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		return nil, fmt.Errorf("upload invalid(max %d MB): %v", maxUpload>>20, err)
	}
	f, _, err := r.FormFile("image")
	if err != nil {
		return nil, errors.New(`missing file field "image"`)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil || len(data) == 0 {
		return nil, errors.New("upload kosong atau tidak bisa dibaca")
	}
	return data, nil
}

func (a *API) recognize(w http.ResponseWriter, r *http.Request) {
	data, err := readUpload(w, r)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	threshold := 0.0
	if s := r.FormValue("threshold"); s != "" {
		if threshold, err = strconv.ParseFloat(s, 64); err != nil || threshold < 0 {
			writeErr(w, 400, errors.New("threshold harus lebih dari >= 0"))
			return
		}
	}
	m, err := a.Mgr.Model()
	if err != nil {
		writeErr(w, 503, err)
		return
	}
	img, _, err := loader.Decode(data)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	start := time.Now()
	q, err := projection.ProjectImage(m, img)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	res, err := recognition.Nearest(m, q, threshold)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{
		"subjectId":       res.Subject,
		"subjectName":     res.Subject,
		"distance":        res.Distance,
		"matchedImage":    res.Image,
		"matchedImageUrl": toImage(res.Image).URL,
		"known":           res.Known,
		"threshold":       res.Threshold,
		"k":               m.K,
		"millis":          float64(time.Since(start).Microseconds()) / 1000,
	})
}

func (a *API) addImage(w http.ResponseWriter, r *http.Request) {
	data, err := readUpload(w, r)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	e, err := a.Mgr.AddImage(strings.TrimSpace(r.FormValue("subject")), data)
	switch {
	case err == nil:
		writeJSON(w, 201, map[string]any{"image": toImage(e.ID()), "subjectId": e.Subject, "status": a.Mgr.Status()})
	case errors.Is(err, dataset.ErrInvalid):
		writeErr(w, 400, fmt.Errorf("nama subyek invalid (letters, digits, '.', '_', '-'; max 64): %v", err))
	case errors.Is(err, dataset.ErrExists):
		writeErr(w, 409, err)
	case isDecodeErr(err):
		writeErr(w, 400, err)
	default: 
		writeErr(w, 422, fmt.Errorf("image tidak bisa ditambahkan, model tidak diubah: %v", err))
	}
}

func isDecodeErr(err error) bool {
	s := err.Error()
	return strings.HasPrefix(s, "unreadable/unsupported") || strings.HasPrefix(s, "corrupted image") || strings.HasPrefix(s, "dimensi image")
}

func (a *API) rebuild(w http.ResponseWriter, r *http.Request) {
	var body struct {
		K int `json:"k"`
	}
	if r.ContentLength != 0 {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
			writeErr(w, 400, errors.New(`body harus JSON-like {"k": 50}`))
			return
		}
	}
	if body.K < 0 {
		writeErr(w, 400, errors.New("k harus >= 1"))
		return
	}
	if err := a.Mgr.Rebuild(body.K); err != nil {
		writeErr(w, 422, err)
		return
	}
	writeJSON(w, 200, a.Mgr.Status())
}


func (a *API) evaluate(w http.ResponseWriter, r *http.Request) {
	m, err := a.Mgr.Model()
	if err != nil {
		writeErr(w, 503, err)
		return
	}
	entries, err := a.Mgr.Store.List(dataset.Validation)
	if err != nil || len(entries) == 0 {
		writeErr(w, 404, errors.New("validation split kosong/tidak bisa dibaca"))
		return
	}
	correct := 0
	var elapsed time.Duration
	for _, e := range entries {
		img, err := loader.Open(a.Mgr.Store.Path(e))
		if err != nil {
			writeErr(w, 500, fmt.Errorf("validation image %s: %v", e.ID(), err))
			return
		}
		start := time.Now()
		q, err := projection.ProjectImage(m, img)
		var res recognition.Result
		if err == nil {
			res, err = recognition.Nearest(m, q, 0)
		}
		elapsed += time.Since(start)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		if res.Subject == e.Subject {
			correct++
		}
	}
	writeJSON(w, 200, map[string]any{
		"k": m.K, "total": len(entries), "correct": correct,
		"accuracy":     float64(correct) / float64(len(entries)),
		"avgMillisPer": float64(elapsed.Microseconds()) / 1000 / float64(len(entries)),
	})
}
