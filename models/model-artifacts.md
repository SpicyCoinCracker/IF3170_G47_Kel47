# Model artifacts
`model.gob` (gob, git-ignored) is written atomically by the server. It holds version, preprocessing size, k, mean,
basis, singular values, labels, image IDs, training projections and a dataset fingerprint (file IDs + sizes + mtimes).
At startup it is reused only if version, k and fingerprint match; otherwise (missing/corrupt/stale) the model is rebuilt.
Do not edit by hand.
