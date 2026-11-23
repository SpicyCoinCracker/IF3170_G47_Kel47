import { ModelStatus } from "../api/client";

export default function StatusBar({ status, error }: { status: ModelStatus | null; error: string }) {
  if (error) return <div className="status bad">{error}</div>;
  if (!status) return <div className="status">Connecting…</div>;
  if (status.building) return <div className="status busy">Updating model…</div>;
  if (!status.ready)
    return <div className="status bad">Model belum siap{status.lastError ? `: ${status.lastError}` : ""}</div>;
  return (
    <div className="status good">
      Model ready · {status.trainImages} images · {status.subjects} subjects · k={status.k} (rank {status.rank}) ·
      variance {(status.varianceRetained * 100).toFixed(1)}%
    </div>
  );
}
