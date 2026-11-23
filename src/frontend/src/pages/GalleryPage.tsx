import { useEffect, useState } from "react";
import { getSubjects, imageUrl, Subject } from "../api/client";

const PER_PAGE = 12;

export default function GalleryPage({ refreshKey }: { refreshKey: number }) {
  const [subjects, setSubjects] = useState<Subject[] | null>(null);
  const [error, setError] = useState("");
  const [page, setPage] = useState(0);
  const [open, setOpen] = useState<Subject | null>(null);

  useEffect(() => {
    getSubjects().then(setSubjects).catch((e) => setError(e.message));
  }, [refreshKey]);

  if (error) return <div className="alert error" role="alert">{error}</div>;
  if (!subjects) return <p className="muted">Loading…</p>;
  if (subjects.length === 0) return <p className="muted">Dataset training kosong.</p>;

  const pages = Math.ceil(subjects.length / PER_PAGE);
  const cur = Math.min(page, pages - 1);
  const shown = subjects.slice(cur * PER_PAGE, (cur + 1) * PER_PAGE);

  return (
    <section>
      <h2>Subjects ({subjects.length})</h2>
      <div className="grid">
        {shown.map((s) => (
          <button key={s.subjectId} className="card" onClick={() => setOpen(s)}>
            <img src={imageUrl(s.images[0].url)} alt={`Subject ${s.subjectName}`} loading="lazy" />
            <span>{s.subjectName}</span>
            <small>{s.images.length} images</small>
          </button>
        ))}
      </div>
      <div className="pager">
        <button disabled={cur === 0} onClick={() => setPage(cur - 1)}>Previous</button>
        <span>Page {cur + 1} / {pages}</span>
        <button disabled={cur >= pages - 1} onClick={() => setPage(cur + 1)}>Next</button>
      </div>
      {open && (
        <div className="modal" role="dialog" aria-label={`Images of ${open.subjectName}`} onClick={() => setOpen(null)}>
          <div className="modal-body" onClick={(e) => e.stopPropagation()}>
            <h3>{open.subjectName}</h3>
            <div className="grid small">
              {open.images.map((im) => <img key={im.id} src={imageUrl(im.url)} alt={im.id} loading="lazy" />)}
            </div>
            <button onClick={() => setOpen(null)}>Close</button>
          </div>
        </div>
      )}
    </section>
  );
}
