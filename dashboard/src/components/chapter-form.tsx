'use client';

import { useState } from 'react';
import { Plus } from 'lucide-react';
import { api, Chapter } from '@/lib/api';

interface Props {
  mangaSlug: string;
  onSuccess?: (chapter: Chapter) => void;
}

export default function ChapterForm({ mangaSlug, onSuccess }: Props) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);

    const fd = new FormData(e.currentTarget);
    const chapterIndex = parseInt(fd.get('chapter_index') as string, 10);
    const title = (fd.get('title') as string).trim();

    if (!chapterIndex || chapterIndex <= 0) {
      setError('Chapter index harus lebih dari 0');
      setLoading(false);
      return;
    }

    try {
      const res = await api.createChapter(mangaSlug, { chapter_index: chapterIndex, title });
      setSuccess(`Chapter ${chapterIndex} berhasil dibuat!`);
      (e.target as HTMLFormElement).reset();
      onSuccess?.(res.data as Chapter);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Gagal membuat chapter');
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
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">
            Chapter Index <span className="text-red-400">*</span>
          </label>
          <input
            name="chapter_index"
            type="number"
            min={1}
            required
            placeholder="cth: 1"
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 text-sm"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">Judul Chapter</label>
          <input
            name="title"
            placeholder="cth: Awal Perjalanan"
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 text-sm"
          />
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
            <Plus size={16} /> Buat Chapter
          </>
        )}
      </button>
    </form>
  );
}
