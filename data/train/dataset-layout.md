# Training split
Place the Georgia Tech (cropped) train split as `data/train/<subject>/<image>.{jpg,jpeg,png}`
(e.g. `data/train/s01/01.jpg`). Subject/file names: letters, digits, `.`, `_`, `-` (max 64, no separators).
Images added through `POST /api/dataset/images` are stored here as `<subject>/<content-hash>.<ext>`.
