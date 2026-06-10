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
  ingestAll: () => 
    apiRequest('/admin/ingest/all', { method: 'POST' }),
  
  ingestSeries: (slug: string) =>
    apiRequest('/admin/ingest/series', {
      method: 'POST',
      body: JSON.stringify({ slug }),
    }),
  
  ingestChapter: (slug: string, chapterIndex: number) =>
    apiRequest('/admin/ingest/chapter', {
      method: 'POST',
      body: JSON.stringify({ slug, chapter_index: chapterIndex }),
    }),
  
  // Jobs endpoints
  getJobs: () =>
    apiRequest<IngestJob[]>('/admin/ingest/jobs'),
  
  getJob: (id: string) =>
    apiRequest<IngestJob>(`/admin/ingest/jobs/${id}`),
  
  cancelJob: (id: string) =>
    apiRequest(`/admin/ingest/jobs/${id}/cancel`, { method: 'POST' }),
};
