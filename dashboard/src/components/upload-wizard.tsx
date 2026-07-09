'use client';

import { useState } from 'react';
import { CheckCircle2, Lock, BookOpen, FileText, Image, ChevronRight } from 'lucide-react';
import MangaUploadForm from './manga-upload-form';
import ChapterForm from './chapter-form';
import PagesUploadForm from './pages-upload-form';
import { Manga, Chapter } from '@/lib/api';

type Step = 'manga' | 'chapter' | 'pages';

interface SessionCtx {
  manga?: { slug: string; title: string };
  chapter?: { index: number; title: string };
}

export default function UploadWizard() {
  const [step, setStep] = useState<Step>('manga');
  const [ctx, setCtx] = useState<SessionCtx>({});

  // derived
  const mangaReady = !!ctx.manga;
  const chapterReady = !!ctx.chapter;

  const steps = [
    {
      key: 'manga' as Step,
      icon: BookOpen,
      label: 'Buat Manga',
      sublabel: ctx.manga ? ctx.manga.title : 'Belum dibuat',
      done: mangaReady,
      locked: false,
    },
    {
      key: 'chapter' as Step,
      icon: FileText,
      label: 'Buat Chapter',
      sublabel: ctx.chapter
        ? `Chapter ${ctx.chapter.index}${ctx.chapter.title ? ` — ${ctx.chapter.title}` : ''}`
        : 'Belum dibuat',
      done: chapterReady,
      locked: !mangaReady,
    },
    {
      key: 'pages' as Step,
      icon: Image,
      label: 'Upload Halaman',
      sublabel: ctx.manga && ctx.chapter
        ? `${ctx.manga.title} › Chapter ${ctx.chapter.index}`
        : 'Tunggu chapter selesai',
      done: false,
      locked: !mangaReady || !chapterReady,
    },
  ];

  function handleMangaSuccess(manga: Manga) {
    setCtx(prev => ({ ...prev, manga: { slug: manga.slug, title: manga.title } }));
    setStep('chapter');
  }

  function handleChapterSuccess(chapter: Chapter) {
    setCtx(prev => ({
      ...prev,
      chapter: {
        index: chapter.upstream_index,
        title: chapter.title || '',
      },
    }));
    setStep('pages');
  }

  function handlePagesSuccess() {
    // reset chapter only — keep manga for next chapter upload
    setCtx(prev => ({ ...prev, chapter: undefined }));
    setStep('chapter');
  }

  function reset() {
    setCtx({});
    setStep('manga');
  }

  return (
    <div className="space-y-6">

      {/* Step bar */}
      <div className="glass-panel rounded-2xl overflow-hidden">
        <div className="grid grid-cols-3 divide-x divide-slate-700/50">
          {steps.map((s, i) => {
            const Icon = s.icon;
            const isActive = step === s.key;
            return (
              <button
                key={s.key}
                onClick={() => !s.locked && setStep(s.key)}
                disabled={s.locked}
                className={`flex items-start gap-3 p-4 transition-colors text-left ${
                  isActive
                    ? 'bg-indigo-500/10 border-b-2 border-indigo-500'
                    : s.done
                    ? 'hover:bg-slate-800/50 cursor-pointer border-b-2 border-emerald-500/40'
                    : s.locked
                    ? 'opacity-40 cursor-not-allowed border-b-2 border-transparent'
                    : 'hover:bg-slate-800/50 cursor-pointer border-b-2 border-transparent'
                }`}
              >
                {/* number / check / lock */}
                <div className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 mt-0.5 ${
                  isActive
                    ? 'bg-indigo-600 text-white'
                    : s.done
                    ? 'bg-emerald-500/20 text-emerald-400'
                    : 'bg-slate-700 text-slate-400'
                }`}>
                  {s.done && !isActive
                    ? <CheckCircle2 className="w-4 h-4" />
                    : s.locked
                    ? <Lock className="w-3.5 h-3.5" />
                    : <Icon className="w-4 h-4" />}
                </div>

                <div className="min-w-0">
                  <div className="flex items-center gap-1">
                    <span className={`text-xs font-semibold ${isActive ? 'text-indigo-300' : s.done ? 'text-emerald-400' : 'text-slate-300'}`}>
                      Langkah {i + 1}
                    </span>
                  </div>
                  <p className={`text-sm font-medium mt-0.5 ${isActive ? 'text-white' : 'text-slate-300'}`}>
                    {s.label}
                  </p>
                  <p className={`text-xs mt-0.5 truncate ${s.done && !isActive ? 'text-emerald-400/80' : 'text-slate-500'}`}>
                    {s.sublabel}
                  </p>
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Context banner — tampil di step chapter & pages */}
      {step !== 'manga' && (
        <div className="flex items-center gap-3 px-4 py-3 bg-slate-800/60 border border-slate-700/50 rounded-xl text-sm">
          <span className="text-slate-400">Konteks aktif:</span>
          {ctx.manga && (
            <span className="flex items-center gap-1.5 bg-indigo-500/10 border border-indigo-500/20 text-indigo-300 px-2.5 py-1 rounded-lg text-xs font-medium">
              <BookOpen className="w-3.5 h-3.5" />
              {ctx.manga.title}
            </span>
          )}
          {ctx.chapter && (
            <>
              <ChevronRight className="w-3.5 h-3.5 text-slate-600" />
              <span className="flex items-center gap-1.5 bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 px-2.5 py-1 rounded-lg text-xs font-medium">
                <FileText className="w-3.5 h-3.5" />
                Chapter {ctx.chapter.index}{ctx.chapter.title ? ` — ${ctx.chapter.title}` : ''}
              </span>
            </>
          )}
          <button
            onClick={reset}
            className="ml-auto text-xs text-slate-500 hover:text-red-400 transition-colors"
          >
            Reset semua
          </button>
        </div>
      )}

      {/* Step content */}
      <div className="glass-panel rounded-2xl p-6 sm:p-8">

        {/* Step 1: Manga */}
        {step === 'manga' && (
          <>
            <div className="mb-6">
              <h3 className="text-lg font-semibold text-slate-100 flex items-center gap-2">
                <BookOpen className="w-5 h-5 text-indigo-400" /> Langkah 1 — Buat Manga
              </h3>
              <p className="text-slate-400 text-sm mt-1">
                Isi informasi manga. Setelah berhasil, kamu akan otomatis lanjut ke pembuatan chapter.
              </p>
            </div>
            <MangaUploadForm onSuccess={handleMangaSuccess} />
          </>
        )}

        {/* Step 2: Chapter */}
        {step === 'chapter' && ctx.manga && (
          <>
            <div className="mb-6">
              <h3 className="text-lg font-semibold text-slate-100 flex items-center gap-2">
                <FileText className="w-5 h-5 text-indigo-400" /> Langkah 2 — Buat Chapter
              </h3>
              <p className="text-slate-400 text-sm mt-1">
                Tambah chapter untuk manga <span className="text-indigo-300 font-medium">{ctx.manga.title}</span>.
                Setelah berhasil, kamu akan lanjut upload halaman.
              </p>
            </div>
            <ChapterForm mangaSlug={ctx.manga.slug} onSuccess={handleChapterSuccess} />
          </>
        )}

        {/* Step 3: Pages */}
        {step === 'pages' && ctx.manga && ctx.chapter && (
          <>
            <div className="mb-6">
              <h3 className="text-lg font-semibold text-slate-100 flex items-center gap-2">
                <Image className="w-5 h-5 text-indigo-400" /> Langkah 3 — Upload Halaman
              </h3>
              <p className="text-slate-400 text-sm mt-1">
                Upload gambar untuk{' '}
                <span className="text-indigo-300 font-medium">{ctx.manga.title}</span>
                {' › '}
                <span className="text-emerald-300 font-medium">
                  Chapter {ctx.chapter.index}{ctx.chapter.title ? ` — ${ctx.chapter.title}` : ''}
                </span>.
                Urutan gambar = urutan halaman.
              </p>
            </div>
            <PagesUploadForm
              mangaSlug={ctx.manga.slug}
              chapterIndex={ctx.chapter.index}
              onSuccess={handlePagesSuccess}
            />
            <p className="mt-4 text-xs text-slate-500 text-center">
              Setelah upload, kamu akan otomatis kembali ke langkah 2 untuk tambah chapter berikutnya.
            </p>
          </>
        )}
      </div>
    </div>
  );
}
