import { useEffect, useState } from "react";
import { recognize, RecognitionResponse } from "../api/client";
import UploadImage from "../components/UploadImage";
import RecognitionResult from "../components/RecognitionResult";

export default function RecognitionPage({ ready }: { ready: boolean }) {
  const [file, setFile] = useState<File | null>(null);
  const [threshold, setThreshold] = useState("0");
  const [result, setResult] = useState<RecognitionResponse | null>(null);
  const [queryUrl, setQueryUrl] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => () => { if (queryUrl) URL.revokeObjectURL(queryUrl); }, [queryUrl]);

  const run = async () => {
    if (!file) return setError("Select an image first.");
    const t = Number(threshold);
    if (!Number.isFinite(t) || t < 0) return setError("Threshold harus ≥ 0.");
    setBusy(true); setError("");
    try {
      setResult(await recognize(file, t));
      setQueryUrl(URL.createObjectURL(file));
    } catch (e) {
      setResult(null); setError((e as Error).message);
    } finally { setBusy(false); }
  };

  return (
    <section className="two-col">
      <div>
        <h2>Query image</h2>
        <UploadImage file={file} onChange={(f) => { setFile(f); setResult(null); }} onError={setError} />
        <label className="field">Unknown-face threshold (0 = off)
          <input type="number" min="0" step="any" value={threshold} onChange={(e) => setThreshold(e.target.value)} />
        </label>
        <button className="primary" disabled={!file || busy || !ready} onClick={run}>
          {busy ? "Recognizing…" : "Recognize"}
        </button>
        {!ready && <p className="muted">The model is not ready yet.</p>}
        <p className="muted">The query is never added to the dataset.</p>
      </div>
      <div>
        <h2>Result</h2>
        <RecognitionResult result={result} queryUrl={queryUrl} error={error} />
      </div>
    </section>
  );
}
