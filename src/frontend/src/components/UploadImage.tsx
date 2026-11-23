import { useEffect, useState } from "react";

const MAX = 8 * 1024 * 1024;

interface Props {
  file: File | null;
  onChange: (f: File | null) => void;
  onError: (msg: string) => void;
}

export default function UploadImage({ file, onChange, onError }: Props) {
  const [preview, setPreview] = useState<string | null>(null);

  useEffect(() => {
    if (!file) return setPreview(null);
    const url = URL.createObjectURL(file);
    setPreview(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);

  const pick = (f: File | undefined) => {
    if (!f) return;
    if (!["image/png", "image/jpeg"].includes(f.type)) return onError("Format file salah(PNG/JPEG)");
    if (f.size > MAX) return onError("Image lebih besar dari 8 MB.");
    onError("");
    onChange(f);
  };

  return (
    <div className="upload">
      <label className="drop">
        <input type="file" accept="image/png,image/jpeg" onChange={(e) => pick(e.target.files?.[0])} />
        {preview ? <img src={preview} alt="Selected query" /> : <span>Pilih image muka(PNG/JPEG)</span>}
      </label>
      {file && <p className="muted">{file.name}</p>}
    </div>
  );
}
