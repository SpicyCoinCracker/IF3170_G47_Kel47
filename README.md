# EigenTsar (backend)

Eigenface face recognition. Go backend, all matrix/Jacobi-eigen/SVD/PCA/projection/nearest-neighbour code is in-house
(standard library only; no external dependencies). Frontend is not covered here.

## Requirements / run / test
- Go >= 1.22
- Put the dataset in `data/train/<subject>/*.jpg` and `data/validation/...` (see `data/*/dataset-layout.md`).
- Run (from `src/backend`): `go run ./cmd/server -addr :8080 -data ../../data -models ../../models -k 50`
- Test (from `test/`): `go test ./...` (separate Go module that imports the backend module)

## Pipeline / assumptions
- Preprocessing (identical for train/validation/query): grayscale (0.299/0.587/0.114) → area-average resize to **64x80** (WxH, aspect ratio not kept) → scale to [0,1] → column-stacking vectorization (d = 5120).
- Model: μ = mean of train columns; A = X − μ; SVD via Gram matrix AᵀA + cyclic Jacobi (converges when off-diagonal norm ≤ 1e-13·‖G‖F, max 100 sweeps); components with γ ≤ 1e-10·γmax are dropped (rank). `1 ≤ k ≤ rank ≤ N−1`, otherwise an error is returned.
- Recognition: Euclidean distance to every training projection (1-NN). **Tie-break:** training images are ordered by `split/subject/file`; the first one wins on exactly equal distance.
- Unknown-face bonus: optional `threshold` form field; `0` (default) disables it; distance > threshold ⇒ `known:false`.
- Queries are never added to the dataset. Dataset additions are only via `POST /api/dataset/images`, which rebuilds the whole model (mean, basis, projections) and swaps it in atomically; on failure the file is removed and the old model stays. Adding images needs a dataset that can already satisfy k (≥ k+1 images and variance).
- Startup: loads `models/model.gob` if version/k/dataset fingerprint match, else rebuilds. If that fails the server still starts and reports `ready:false`.
- Build time ≈ 12 s for 550 images on a laptop (Jacobi on a 550x550 Gram matrix).

## API (errors are always `{"error": "..."}`)
| Endpoint | Notes |
|---|---|
| `GET /api/subjects` | `[{subjectId, subjectName, images:[{id,url}]}]` (train split) |
| `GET /api/images/{split}/{subject}/{file}` | serves an image (`train`/`validation`) |
| `GET /api/model/status` | ready, k, rank, subjects, varianceRetained, lastError, … |
| `POST /api/recognize` | multipart `image` (PNG/JPEG ≤ 8 MB), optional `threshold` → `{subjectId, subjectName, distance, matchedImage, matchedImageUrl, known, threshold, k, millis}`; 503 if model not ready |
| `POST /api/dataset/images` | multipart `image` + `subject` (existing or new) → 201 with new status; 400 invalid, 409 duplicate, 422 rebuild failed |
| `POST /api/model/rebuild` | optional JSON `{"k": 80}` |
| `GET /api/evaluate` | accuracy + avg ms/query on the validation split for the current k (use with rebuild for k comparison) |

## Limitations
No eigenface-gallery endpoint, no face detection, PNG/JPEG only; Gram-matrix SVD loses precision for tiny singular values; rebuilds block the add request until finished.

## Frontend (`src/frontend`, React + TypeScript + Vite)
- Requires Node >= 18. `cd src/frontend && npm install && npm run dev` (open http://localhost:5173; `/api` is proxied to the backend on :8080 — start the backend first).
- Production build: `npm run build` (output `dist/`); set `VITE_API_BASE` to the backend URL if not served from the same origin (the backend allows CORS).
- Pages: Gallery (paged subjects, click for all images), Recognition (preview, optional threshold, result with matched image + distance), Dataset & Model (add image to existing/new subject, change k and rebuild, validation accuracy).
