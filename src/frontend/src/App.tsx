import { useCallback, useEffect, useState } from "react";
import { getStatus, ModelStatus } from "./api/client";
import StatusBar from "./components/StatusBar";
import GalleryPage from "./pages/GalleryPage";
import RecognitionPage from "./pages/RecognitionPage";
import DatasetPage from "./pages/DatasetPage";
import EigenfacePage from "./pages/EigenfacePage";

type Tab = "eigen" | "gallery" | "recognize" | "dataset";

export default function App() {
  const [tab, setTab] = useState<Tab>("gallery");
  const [status, setStatus] = useState<ModelStatus | null>(null);
  const [statusError, setStatusError] = useState("");
  const [version, setVersion] = useState(0); // bump after dataset/model changes

  const refresh = useCallback(() => {
    getStatus().then((s) => { setStatus(s); setStatusError(""); }).catch((e) => setStatusError(e.message));
    setVersion((v) => v + 1);
  }, []);
  useEffect(refresh, [refresh]);

  const tabs: [Tab, string][] = [["gallery", "Gallery"], ["recognize", "Recognition"], ["eigen", "Eigenfaces"], ["dataset", "Dataset & Model"]];
  return (
    <>
      <header>
        <h1>EigenTsar</h1>
        <nav>
          {tabs.map(([id, label]) => (
            <button key={id} className={tab === id ? "active" : ""} onClick={() => setTab(id)}>{label}</button>
          ))}
        </nav>
      </header>
      <StatusBar status={status} error={statusError} />
      <main>
        {tab === "gallery" && <GalleryPage refreshKey={version} />}
        {tab === "recognize" && <RecognitionPage ready={!!status?.ready} />}
        {tab === "dataset" && <DatasetPage status={status} onModelChanged={refresh} />}
        {tab === "eigen" && <EigenfacePage refreshKey={version} />}
      </main>
    </>
  );
}
