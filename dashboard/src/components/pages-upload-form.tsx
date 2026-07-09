'use client';

import { useState, useRef } from 'react';
import { Upload, X, ImageIcon } from 'lucide-react';
import { api, MangaPage } from '@/lib/api';

interface Props {
  mangaSlug: string;
  chapterIndex: number;
  onSuccess?: (pages: MangaPage[]) => void;
}

export default function PagesUploadForm({ mangaSlug, chapterIndex, onSuccess }: Props) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [previews, setPreviews] = useState<{ name: string; url: string }[]>([]);
  const [files, setFiles] = useState<File[]>([]);
  const inputRef = useRef<HTMLInputElement>(null);

  function handleFilesChange(e: React.ChangeEvent<HTMLInputElement>) {
    const selected = Array.from(e.target.files || []);
    if (!selected.length) return;

    // append, dedupe by name
    setFiles(prev => {
      const existing = new Set(prev.map(f => f.name));
      const newFiles = selected.filter(f => !existing.has(f.name));
      return [...prev, ...newFiles];
    });
    setPreviews(prev => {
      const existing = new Set(prev.map(p => p.name));
      const newPreviews = selected
        .filter(f => !existing.has(f.name))
        .map(f => ({ name: f.name, url: URL.createObjectURL(f) }));
      return [...prev, ...newPreviews];
    });

    // reset input so same files can be re-added after removal
    if (inputRef.current) inputRef.current.value = '';
  }

  function removeFile(name: string) {
    setFiles(prev => prev.filter(f => f.name !== name));
    setPreviews(prev => {
      const removed = prev.find(p => p.name === name);
      if (removed) URL.revokeObjectURL(removed.url);
      return prev.filter(p => p.name !== name);
    });
  }

  function moveFile(from: number, to: number) {
    if (to < 0 || to >= files.length) return;
    const newFiles = [...files];
    const newPreviews = [...previews];
    [newFiles[from], newFiles[to]] = [newFiles[to], newFiles[from]];
    [newPreviews[from], newPreviews[to]] = [newPreviews[to], newPreviews[from]];
    setFiles(newFiles);
    setPreviews(newPreviews);
  }

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (files.length === 0) {
      setError('Pilih minimal 1 file');
      return;
    }

    setLoading(true);
    const fd = new FormData();
    files.forEach(f => fd.append('files', f));

    try {
      const res = await api.uploadPages(mangaSlug, chapterIndex, fd);
      setSuccess(`${res.data?.length ?? 0} halaman berhasil diupload!`);
      setFiles([]);
      setPreviews([]);
      onSuccess?.(res.data as MangaPage[]);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Gagal upload halaman');
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="bg-red-500/10 border border-red-500/30 text-red-400 px-4 py-3 rounded-lg text-sm">
          {error}
        </div>
      )}
      {success && (
        <div className="bg-green-500/10 border border-green-500/30 text-green-400 px-4 py-3 rounded-lg text-sm">
          {success}
        </div>
      )}

      {/* Drop zone */}
      <label className="flex flex-col items-center justify-center w-full h-32 bg-gray-800 border-2 border-dashed border-gray-600 rounded-lg cursor-pointer hover:border-blue-500 transition-colors">
        <Upload size={24} className="text-gray-500 mb-2" />
        <span className="text-sm text-gray-400">Klik atau drag gambar ke sini</span>
        <span className="text-xs text-gray-600 mt-1">JPG, PNG, WEBP, GIF</span>
        <input
          ref={inputRef}
          type="file"
          accept="image/*"
          multiple
          className="hidden"
          onChange={handleFilesChange}
        />
      </label>

      {/* Preview grid */}
      {previews.length > 0 && (
        <div className="space-y-2">
          <p className="text-xs text-gray-500">{previews.length} file dipilih — drag nomor untuk reorder</p>
          <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-6 gap-2 max-h-72 overflow-y-auto pr-1">
            {previews.map((p, i) => (
              <div key={p.name} className="relative group">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={p.url}
                  alt={p.name}
                  className="w-full aspect-[3/4] object-cover rounded-lg border border-gray-700"
                />
                {/* page number badge */}
                <span className="absolute top-1 left-1 bg-black/70 text-white text-[10px] px-1.5 py-0.5 rounded">
                  {i + 1}
                </span>
                {/* actions */}
                <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity rounded-lg flex items-center justify-center gap-1">
                  <button
                    type="button"
                    onClick={() => moveFile(i, i - 1)}
                    disabled={i === 0}
                    className="bg-gray-700 hover:bg-gray-600 disabled:opacity-30 text-white text-xs px-1.5 py-1 rounded"
                  >
                    ←
                  </button>
                  <button
                    type="button"
                    onClick={() => removeFile(p.name)}
                    className="bg-red-600 hover:bg-red-700 text-white p-1 rounded"
                  >
                    <X size={10} />
                  </button>
                  <button
                    type="button"
                    onClick={() => moveFile(i, i + 1)}
                    disabled={i === previews.length - 1}
                    className="bg-gray-700 hover:bg-gray-600 disabled:opacity-30 text-white text-xs px-1.5 py-1 rounded"
                  >
                    →
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {previews.length === 0 && (
        <div className="flex items-center gap-2 text-gray-600 text-sm">
          <ImageIcon size={14} /> Belum ada gambar dipilih
        </div>
      )}

      <button
        type="submit"
        disabled={loading || files.length === 0}
        className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-blue-600/50 disabled:cursor-not-allowed text-white font-medium py-2.5 px-4 rounded-lg text-sm transition-colors flex items-center justify-center gap-2"
      >
        {loading ? (
          <>
            <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            Mengupload {files.length} halaman...
          </>
        ) : (
          <>
            <Upload size={16} /> Upload {files.length > 0 ? `${files.length} Halaman` : 'Halaman'}
          </>
        )}
      </button>
    </form>
  );
}
