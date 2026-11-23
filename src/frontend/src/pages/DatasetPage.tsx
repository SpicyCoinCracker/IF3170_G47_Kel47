import { useState } from "react";
import { addImage, evaluate, Evaluation, ModelStatus, rebuild } from "../api/client";
import UploadImage from "../components/UploadImage";

interface Props { status: ModelStatus | null; onModelChanged: () => void }

export default function DatasetPage({ status, onModelChanged }: Props) {
  const [file, setFile] = useState<File | null>(null);
  const [subject, setSubject] = useState("");
  const [k, setK] = useState("");
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);
  const [busy, setBusy] = useState(false);
  const [evalRes, setEvalRes] = useState<Evaluation | null>(null);

  const wrap = async (fn: () => Promise<string>) => {
    setBusy(true); setMsg(null);
    try { setMsg({ ok: true, text: await fn() }); onModelChanged(); }
    catch (e) { setMsg({ ok: false, text: (e as Error).message }); }
    finally { setBusy(false); }
  };

  const add = () => {
    if (!file) return setMsg({ ok: false, text: "Pilih sebuah image dahulu." });
    if (!/^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/.test(subject.trim()))
      return setMsg({ ok: false, text: "Subject: letters, digits, '.', '_', '-' (max 64), e.g. s51." });
    wrap(async () => {
      const r = await addImage(file, subject.trim());
      setFile(null);
      return `Ditambahkan ke ${r.subjectId}. Model me-rebuilt: ${r.status.trainImages} images, ${r.status.subjects} subjects.`;
    });
  };

  const rebuildK = () => {
    const n = k === "" ? undefined : Number(k);
    if (n !== undefined && (!Number.isInteger(n) || n < 1)) return setMsg({ ok: false, text: "k harus integer ≥ 1." });
    wrap(async () => { const s = await rebuild(n); setEvalRes(null); return `Model di-rebuilt dengan k=${s.k}.`; });
  };

  return (
    <section className="two-col">
      <div>
        <h2>Add image to dataset</h2>
        <UploadImage file={file} onChange={setFile} onError={(t) => t && setMsg({ ok: false, text: t })} />
        <label className="field">Subject (existing or new)
          <input value={subject} onChange={(e) => setSubject(e.target.value)} placeholder="s01" />
        </label>
        <button className="primary" disabled={busy} onClick={add}>{busy ? "Model sedang diupdate…" : "Tambah dan update model"}</button>
      </div>
      <div>
        <h2>Model</h2>
        <label className="field">Number of eigenfaces k (current: {status?.k ?? "–"}, max {status?.rank ?? "–"})
          <input type="number" min="1" value={k} onChange={(e) => setK(e.target.value)} />
        </label>
        <button disabled={busy} onClick={rebuildK}>Rebuild model</button>
        <button disabled={busy || !status?.ready} onClick={() => wrap(async () => { setEvalRes(await evaluate()); return "Validation selesai."; })}>
          Evaluate on validation
        </button>
        {evalRes && (
          <p>k={evalRes.k}: accuracy <b>{(evalRes.accuracy * 100).toFixed(1)}%</b> ({evalRes.correct}/{evalRes.total}),
            {" "}{evalRes.avgMillisPer.toFixed(2)} ms/query</p>
        )}
      </div>
      {msg && <div className={`alert ${msg.ok ? "okmsg" : "error"} span`} role="status">{msg.text}</div>}
    </section>
  );
}
