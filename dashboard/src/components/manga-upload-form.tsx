'use client';

import { useState } from 'react';
import { Upload, X, Plus } from 'lucide-react';
import { api, Manga } from '@/lib/api';

interface Props {
  onSuccess?: (manga: Manga) => void;
}

const STATUS_OPTIONS = ['ongoing', 'completed', 'hiatus', 'dropped'];
const TYPE_OPTIONS = ['manga', 'manhwa', 'manhua', 'webtoon'];

export default function MangaUploadForm({ onSuccess }: Props) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [coverPreview, setCoverPreview] = useState('');
  const [genreInput, setGenreInput] = useState('');
  const [genres, setGenres] = useState<string[]>([]);

  function handleCoverChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setCoverPreview(URL.createObjectURL(file));
  }

  function addGenre() {
    const slug = genreInput.trim().toLowerCase().replace(/\s+/g, '-');
    if (!slug || genres.includes(slug)) return;
    setGenres([...genres, slug]);
    setGenreInput('');
  }

  function removeGenre(slug: string) {
    setGenres(genres.filter(g => g !== slug));
  }

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);

    const form = e.currentTarget;
    const fd = new FormData(form);

    // replace genres[] with current state
    fd.delete('genres');
    genres.forEach(g => fd.append('genres', g));

    try {
      const res = await api.createManga(fd);
      setSuccess(`Manga "${res.data?.title}" berhasil dibuat!`);
      form.reset();
      setCoverPreview('');
      setGenres([]);
      onSuccess?.(res.data as Manga);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Gagal membuat manga');
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

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Title */}
        <div className="md:col-span-2">
          <label className="block text-sm font-medium text-gray-300 mb-1">
            Judul <span className="text-red-400">*</span>
          </label>
          <input
            name="title"
            required
            placeholder="Masukkan judul manga"
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 text-sm"
          />
        </div>

        {/* Native Title */}
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">Judul Asli</label>
          <input
            name="native_title"
            placeholder="Judul dalam bahasa asli"
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 text-sm"
          />
        </div>

        {/* Author */}
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">Author</label>
          <input
            name="author"
            placeholder="Nama author"
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 text-sm"
          />
        </div>

        {/* Status */}
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">Status</label>
          <select
            name="status"
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-500 text-sm"
          >
            <option value="">-- Pilih Status --</option>
            {STATUS_OPTIONS.map(s => (
              <option key={s} value={s}>{s.charAt(0).toUpperCase() + s.slice(1)}</option>
            ))}
          </select>
        </div>

        {/* Type */}
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">Tipe</label>
          <select
            name="type"
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-500 text-sm"
          >
            <option value="">-- Pilih Tipe --</option>
            {TYPE_OPTIONS.map(t => (
              <option key={t} value={t}>{t.charAt(0).toUpperCase() + t.slice(1)}</option>
            ))}
          </select>
        </div>

        {/* Synopsis */}
        <div className="md:col-span-2">
          <label className="block text-sm font-medium text-gray-300 mb-1">Sinopsis</label>
          <textarea
            name="synopsis"
            rows={4}
            placeholder="Deskripsi singkat manga..."
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 text-sm resize-none"
          />
        </div>

        {/* Genres */}
        <div className="md:col-span-2">
          <label className="block text-sm font-medium text-gray-300 mb-1">Genre (slug)</label>
          <div className="flex gap-2 mb-2">
            <input
              value={genreInput}
              onChange={e => setGenreInput(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && (e.preventDefault(), addGenre())}
              placeholder="cth: action, romance"
              className="flex-1 bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 text-sm"
            />
            <button
              type="button"
              onClick={addGenre}
              className="bg-gray-700 hover:bg-gray-600 text-white px-3 py-2 rounded-lg text-sm flex items-center gap-1"
            >
              <Plus size={14} /> Tambah
            </button>
          </div>
          {genres.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {genres.map(g => (
                <span key={g} className="flex items-center gap-1 bg-blue-500/20 text-blue-300 border border-blue-500/30 px-2 py-1 rounded-full text-xs">
                  {g}
                  <button type="button" onClick={() => removeGenre(g)} className="hover:text-red-400">
                    <X size={10} />
                  </button>
                </span>
              ))}
            </div>
          )}
        </div>

        {/* Cover */}
        <div className="md:col-span-2">
          <label className="block text-sm font-medium text-gray-300 mb-1">Cover</label>
          <div className="flex items-start gap-4">
            <label className="flex flex-col items-center justify-center w-32 h-44 bg-gray-800 border-2 border-dashed border-gray-600 rounded-lg cursor-pointer hover:border-blue-500 transition-colors">
              {coverPreview ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={coverPreview} alt="cover preview" className="w-full h-full object-cover rounded-lg" />
              ) : (
                <>
                  <Upload size={20} className="text-gray-500 mb-1" />
                  <span className="text-xs text-gray-500">Upload Cover</span>
                </>
              )}
              <input name="cover" type="file" accept="image/*" className="hidden" onChange={handleCoverChange} />
            </label>
            {coverPreview && (
              <button
                type="button"
                onClick={() => { setCoverPreview(''); }}
                className="text-xs text-red-400 hover:text-red-300"
              >
                Hapus
              </button>
            )}
          </div>
        </div>
      </div>

      <button
        type="submit"
        disabled={loading}
        className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-blue-600/50 disabled:cursor-not-allowed text-white font-medium py-2.5 px-4 rounded-lg text-sm transition-colors flex items-center justify-center gap-2"
      >
        {loading ? (
          <>
            <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            Membuat...
          </>
        ) : (
          <>
            <Upload size={16} /> Buat Manga
          </>
        )}
      </button>
    </form>
  );
}
