import { useEffect, useState } from "react";
import { EigenfaceGallery, getEigenfaces, imageUrl } from "../api/client";

const STEP = 24;

// Reloads whenever `refreshKey` changes (i.e. after the model was rebuilt).
export default function EigenfacePage({ refreshKey }: { refreshKey: number }) {
  const [data, setData] = useState<EigenfaceGallery | null>(null);
  const [error, setError] = useState("");
  const [shown, setShown] = useState(STEP);

  useEffect(() => {
    getEigenfaces().then((d) => { setData(d); setError(""); }).catch((e) => setError(e.message));
  }, [refreshKey]);

  if (error) return <div className="alert error" role="alert">{error}</div>;
  if (!data) return <p className="muted">Loading…</p>;

  return (
    <section>
      <h2>Mean face</h2>
      <figure className="eigen"><img src={imageUrl(data.meanUrl)} alt="Mean face" /><figcaption>μ</figcaption></figure>
      <h2>Eigenfaces (top {Math.min(shown, data.k)} of {data.k})</h2>
      <p className="muted">Ordered by singular value. Pixels are min–max normalised per image for display only.</p>
      <div className="grid">
        {data.components.slice(0, shown).map((c) => (
          <figure key={c.index} className="eigen">
            <img src={imageUrl(c.url)} alt={`Eigenface ${c.index}`} loading="lazy" />
            <figcaption>#{c.index} · σ = {c.singular.toFixed(3)}</figcaption>
          </figure>
        ))}
      </div>
      {shown < data.k && <button onClick={() => setShown(shown + STEP)}>Show more</button>}
    </section>
  );
}
