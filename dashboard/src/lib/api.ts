export interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data?: T;
}

export interface IngestJob {
  id: string;
  type: string;
  slug?: string;
  status: string;
  progress?: number;
  message?: string;
  created_at: string;
  updated_at: string;
}

export interface Manga {
  id: string;
  slug: string;
  title: string;
  native_title?: string;
  author?: string;
  status?: string;
  type?: string;
  synopsis?: string;
  cover_preview_url?: string;
  cover_thumbnail_url?: string;
  source: string;
  genres?: { id: string; name: string; slug: string }[];
  created_at: string;
  updated_at: string;
}

export interface Chapter {
  id: string;
  manga_id: string;
  upstream_index: number;
  chapter_key: string;
  slug: string;
  title?: string;
  total_pages: number;
  created_at: string;
  updated_at: string;
}

export interface MangaPage {
  id: string;
  chapter_id: string;
  page_number: number;
  preview_url: string;
  download_url: string;
  thumbnail_url: string;
  mime_type: string;
  size: number;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8001/api/v1';
const ADMIN_TOKEN = process.env.NEXT_PUBLIC_ADMIN_TOKEN || '';

async function apiRequest<T = any>(
  endpoint: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> {
  const url = `${API_BASE_URL}${endpoint}`;
  
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    'X-Admin-Token': ADMIN_TOKEN,
    ...(options.headers || {}),
  };

  try {
    const response = await fetch(url, {
      ...options,
      headers,
    });

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('API request failed:', error);
    throw error;
  }
}

export const api = {
  // Ingest endpoints
  ingestAll: (opts?: { force?: boolean; missing_only?: boolean }) =>
    apiRequest('/admin/ingest/all', {
      method: 'POST',
      body: JSON.stringify(opts ?? {}),
    }),

  ingestSeries: (slug: string, opts?: { force?: boolean; missing_only?: boolean }) =>
    apiRequest('/admin/ingest/series', {
      method: 'POST',
      body: JSON.stringify({ slug, ...opts }),
    }),

  ingestChapter: (slug: string, chapterIndex: number, opts?: { force?: boolean; missing_only?: boolean }) =>
    apiRequest('/admin/ingest/chapter', {
      method: 'POST',
      body: JSON.stringify({ slug, chapter_index: chapterIndex, ...opts }),
    }),
  
  // Jobs endpoints
  getJobs: () =>
    apiRequest<IngestJob[]>('/admin/ingest/jobs'),
  
  getJob: (id: string) =>
    apiRequest<IngestJob>(`/admin/ingest/jobs/${id}`),
  
  cancelJob: (id: string) =>
    apiRequest(`/admin/ingest/jobs/${id}/cancel`, { method: 'POST' }),

  // Manual upload endpoints
  createManga: (form: FormData) =>
    apiRequest<Manga>('/admin/manga', {
      method: 'POST',
      headers: { 'X-Admin-Token': ADMIN_TOKEN },
      body: form,
    }),

  updateManga: (slug: string, body: Partial<Pick<Manga, 'title' | 'native_title' | 'author' | 'status' | 'type' | 'synopsis'> & { genres: string[] }>) =>
    apiRequest<Manga>(`/admin/manga/${slug}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),

  deleteManga: (slug: string) =>
    apiRequest(`/admin/manga/${slug}`, { method: 'DELETE' }),

  createChapter: (mangaSlug: string, body: { chapter_index: number; title?: string }) =>
    apiRequest<Chapter>(`/admin/manga/${mangaSlug}/chapters`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  deleteChapter: (mangaSlug: string, chapterIndex: number) =>
    apiRequest(`/admin/manga/${mangaSlug}/chapters/${chapterIndex}`, { method: 'DELETE' }),

  uploadPages: (mangaSlug: string, chapterIndex: number, form: FormData) =>
    apiRequest<MangaPage[]>(`/admin/manga/${mangaSlug}/chapters/${chapterIndex}/pages`, {
      method: 'POST',
      headers: { 'X-Admin-Token': ADMIN_TOKEN },
      body: form,
    }),
};
