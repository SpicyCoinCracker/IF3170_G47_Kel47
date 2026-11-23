// All backend calls live here. Base URL: VITE_API_BASE (default "" = same origin / dev proxy).
const BASE: string = import.meta.env.VITE_API_BASE ?? "";

export interface ImageRef { id: string; url: string }
export interface Subject { subjectId: string; subjectName: string; images: ImageRef[] }
export interface ModelStatus {
  ready: boolean; building: boolean; k: number; rank: number; width: number; height: number;
  trainImages: number; subjects: number; varianceRetained: number; builtAt?: string;
  buildMillis: number; configuredK: number; lastError?: string; saveError?: string;
}
export interface RecognitionResponse {
  subjectId: string; subjectName: string; distance: number; matchedImage: string;
  matchedImageUrl: string; known: boolean; threshold: number; k: number; millis: number;
}
export interface Evaluation { k: number; total: number; correct: number; accuracy: number; avgMillisPer: number }
export interface AddResponse { image: ImageRef; subjectId: string; status: ModelStatus }

export const imageUrl = (path: string) => BASE + path;

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(BASE + path, init);
  } catch {
    throw new Error("Backend tidak bisa diakses. Cek status server");
  }
  const body = await res.json().catch(() => null);
  if (!res.ok) throw new Error(body?.error ?? `Request failed (HTTP ${res.status})`);
  return body as T;
}

export const getSubjects = () => request<Subject[]>("/api/subjects");
export const getStatus = () => request<ModelStatus>("/api/model/status");
export const evaluate = () => request<Evaluation>("/api/evaluate");
export const rebuild = (k?: number) =>
  request<ModelStatus>("/api/model/rebuild", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(k ? { k } : {}),
  });

export function recognize(file: File, threshold: number) {
  const f = new FormData();
  f.append("image", file);
  if (threshold > 0) f.append("threshold", String(threshold));
  return request<RecognitionResponse>("/api/recognize", { method: "POST", body: f });
}

export function addImage(file: File, subject: string) {
  const f = new FormData();
  f.append("image", file);
  f.append("subject", subject);
  return request<AddResponse>("/api/dataset/images", { method: "POST", body: f });
}

export interface Eigenface { index: number; singular: number; url: string }
export interface EigenfaceGallery { k: number; meanUrl: string; components: Eigenface[] }
export const getEigenfaces = () => request<EigenfaceGallery>("/api/model/eigenfaces");
