import { imageUrl, RecognitionResponse } from "../api/client";

interface Props { result: RecognitionResponse | null; queryUrl: string | null; error: string }

export default function RecognitionResult({ result, queryUrl, error }: Props) {
  if (error) return <div className="alert error" role="alert">{error}</div>;
  if (!result) return <p className="muted">Hasil belum ada. Upload image dan tekan Recognize.</p>;
  return (
    <div className="result">
      <div className={`badge ${result.known ? "ok" : "warn"}`}>
        {result.known ? `Identified: ${result.subjectName}` : "Unknown face (distance above threshold)"}
      </div>
      <div className="pair">
        <figure>{queryUrl && <img src={queryUrl} alt="Query" />}<figcaption>Query</figcaption></figure>
        <figure>
          <img src={imageUrl(result.matchedImageUrl)} alt={`Matched training image of ${result.subjectName}`} />
          <figcaption>Closest training image ({result.subjectName})</figcaption>
        </figure>
      </div>
      <dl>
        <dt>Euclidean distance</dt><dd>{result.distance.toFixed(4)}</dd>
        <dt>Threshold</dt><dd>{result.threshold > 0 ? result.threshold : "disabled"}</dd>
        <dt>Components (k)</dt><dd>{result.k}</dd>
        <dt>Recognition time</dt><dd>{result.millis.toFixed(1)} ms</dd>
      </dl>
    </div>
  );
}
