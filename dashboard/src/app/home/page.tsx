'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { api, type IngestJob } from '@/lib/api';
import { 
  LogOut, 
  Library, 
  BookOpen, 
  FileText, 
  Activity,
  RefreshCw,
  Terminal,
  ShieldAlert,
  CheckCircle2,
  Clock,
  Loader2,
  LayoutDashboard
} from 'lucide-react';

export default function DashboardPage() {
  const router = useRouter();
  const [jobs, setJobs] = useState<IngestJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [responses, setResponses] = useState<Record<string, { message: string; success: boolean }>>({});
  const [isRefreshing, setIsRefreshing] = useState(false);

  const loadJobs = async (manual = false) => {
    if (manual) setIsRefreshing(true);
    try {
      const result = await api.getJobs();
      if (result.data) {
        const jobsList = (result.data as any).data || result.data;
        setJobs(Array.isArray(jobsList) ? jobsList : []);
      }
    } catch (error) {
      console.error('Failed to load jobs:', error);
    } finally {
      setLoading(false);
      if (manual) setTimeout(() => setIsRefreshing(false), 500);
    }
  };

  useEffect(() => {
    loadJobs();
    const interval = setInterval(() => loadJobs(false), 5000); 
    return () => clearInterval(interval);
  }, []);

  const handleLogout = async () => {
    await fetch('/api/logout', { method: 'POST' });
    router.push('/login');
    router.refresh();
  };

  const showResponse = (key: string, message: string, success: boolean) => {
    setResponses(prev => ({ ...prev, [key]: { message, success } }));
    setTimeout(() => {
      setResponses(prev => {
        const newResponses = { ...prev };
        delete newResponses[key];
        return newResponses;
      });
    }, 8000);
  };

  const handleIngestAll = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const result = await api.ingestAll();
      showResponse('ingestAll', JSON.stringify(result, null, 2), result.success);
      setTimeout(() => loadJobs(true), 1000);
    } catch (error) {
      showResponse('ingestAll', `Error: ${error}`, false);
    }
  };

  const handleIngestSeries = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const slug = formData.get('slug') as string;

    try {
      const result = await api.ingestSeries(slug);
      showResponse('ingestSeries', JSON.stringify(result, null, 2), result.success);
      setTimeout(() => loadJobs(true), 1000);
      form.reset();
    } catch (error) {
      showResponse('ingestSeries', `Error: ${error}`, false);
    }
  };

  const handleIngestChapter = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const slug = formData.get('slug') as string;
    const chapterIndex = parseInt(formData.get('chapter_index') as string);

    try {
      const result = await api.ingestChapter(slug, chapterIndex);
      showResponse('ingestChapter', JSON.stringify(result, null, 2), result.success);
      setTimeout(() => loadJobs(true), 1000);
      form.reset();
    } catch (error) {
      showResponse('ingestChapter', `Error: ${error}`, false);
    }
  };

  return (
    <div className="min-h-screen pb-12 relative overflow-hidden">
      {/* Background Ornaments */}
      <div className="absolute top-0 left-[20%] w-[50%] h-[300px] bg-indigo-600/10 rounded-full blur-[120px] pointer-events-none"></div>

      {/* Header */}
      <header className="sticky top-0 z-40 backdrop-blur-xl bg-slate-950/60 border-b border-slate-800/80 shadow-sm">
        <div className="max-w-7xl mx-auto px-6 h-20 flex justify-between items-center">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-600/20">
              <LayoutDashboard className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="text-xl font-[family-name:var(--font-display)] font-bold text-slate-100 tracking-tight leading-tight">
                Manga API
              </h1>
              <p className="text-[11px] text-indigo-400 font-medium uppercase tracking-wider">Control Center</p>
            </div>
          </div>
          <button
            onClick={handleLogout}
            className="flex items-center gap-2 text-sm font-medium text-slate-400 hover:text-white bg-slate-800/50 hover:bg-slate-800 px-4 py-2.5 rounded-lg border border-slate-700/50 transition-all active:scale-95"
          >
            <LogOut className="w-4 h-4" />
            <span>Sign Out</span>
          </button>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-6 pt-10">
        
        <div className="mb-10 animate-fade-in-up" style={{ animationDelay: '0.1s' }}>
          <h2 className="text-2xl font-[family-name:var(--font-display)] font-bold text-slate-100 mb-2">Ingestion Operations</h2>
          <p className="text-slate-400 text-sm">Trigger manual data synchronization from external sources to the database.</p>
        </div>

        {/* Action Cards Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-12">
          {/* Ingest All Card */}
          <ActionCard
            title="Ingest All Series"
            icon={<Library className="w-6 h-6 text-white" />}
            description="Trigger a complete synchronization of all available manga series and their chapters."
            onSubmit={handleIngestAll}
            response={responses.ingestAll}
            delay="0.2s"
          />

          {/* Ingest Series Card */}
          <ActionCard
            title="Ingest Specific Series"
            icon={<BookOpen className="w-6 h-6 text-white" />}
            description="Fetch and synchronize metadata and chapters for a specific manga series."
            onSubmit={handleIngestSeries}
            response={responses.ingestSeries}
            delay="0.3s"
          >
            <div className="mb-4 space-y-1.5">
              <label htmlFor="seriesSlug" className="block text-xs font-medium text-slate-400 ml-1">
                Series Slug
              </label>
              <input
                type="text"
                id="seriesSlug"
                name="slug"
                placeholder="e.g., one-piece"
                required
                className="glass-input w-full px-4 py-2.5 rounded-xl text-sm"
              />
            </div>
          </ActionCard>

          {/* Ingest Chapter Card */}
          <ActionCard
            title="Ingest Chapter Content"
            icon={<FileText className="w-6 h-6 text-white" />}
            description="Fetch individual pages and assets for a specific chapter within a series."
            onSubmit={handleIngestChapter}
            response={responses.ingestChapter}
            delay="0.4s"
          >
            <div className="grid grid-cols-2 gap-3 mb-4">
              <div className="space-y-1.5 col-span-2 sm:col-span-1">
                <label htmlFor="chapterSlug" className="block text-xs font-medium text-slate-400 ml-1">
                  Series Slug
                </label>
                <input
                  type="text"
                  id="chapterSlug"
                  name="slug"
                  placeholder="one-piece"
                  required
                  className="glass-input w-full px-4 py-2.5 rounded-xl text-sm"
                />
              </div>
              <div className="space-y-1.5 col-span-2 sm:col-span-1">
                <label htmlFor="chapterIndex" className="block text-xs font-medium text-slate-400 ml-1">
                  Chapter Index
                </label>
                <input
                  type="number"
                  id="chapterIndex"
                  name="chapter_index"
                  placeholder="e.g., 1084"
                  required
                  className="glass-input w-full px-4 py-2.5 rounded-xl text-sm"
                />
              </div>
            </div>
          </ActionCard>
        </div>

        {/* Jobs Section */}
        <div className="glass-panel rounded-2xl p-6 sm:p-8 animate-fade-in-up" style={{ animationDelay: '0.5s' }}>
          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-8 pb-6 border-b border-slate-800">
            <div>
              <h2 className="text-xl font-[family-name:var(--font-display)] font-bold text-slate-100 flex items-center gap-2">
                <Activity className="w-5 h-5 text-indigo-400" />
                Active & Recent Jobs
              </h2>
              <p className="text-slate-400 text-xs mt-1">Live monitoring of background ingestion tasks</p>
            </div>
            <button
              onClick={() => loadJobs(true)}
              disabled={isRefreshing}
              className="flex items-center gap-2 bg-slate-800/80 hover:bg-slate-700 text-slate-200 text-sm px-4 py-2 rounded-lg border border-slate-700 transition-all disabled:opacity-50"
            >
              <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
              <span>Refresh</span>
            </button>
          </div>

          {loading && jobs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-slate-500">
              <Loader2 className="w-8 h-8 animate-spin text-indigo-500 mb-4" />
              <p className="text-sm">Fetching job history...</p>
            </div>
          ) : jobs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-slate-500 border border-dashed border-slate-700 rounded-xl bg-slate-900/20">
              <Terminal className="w-10 h-10 text-slate-600 mb-3" />
              <p>No jobs found in the system.</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4">
              {jobs.map((job) => (
                <JobItem key={job.id} job={job} />
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

interface ActionCardProps {
  title: string;
  icon: React.ReactNode;
  description: string;
  onSubmit: (e: React.FormEvent<HTMLFormElement>) => void;
  response?: { message: string; success: boolean };
  children?: React.ReactNode;
  delay: string;
}

function ActionCard({ title, icon, description, onSubmit, response, children, delay }: ActionCardProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    setIsSubmitting(true);
    await onSubmit(e);
    setIsSubmitting(false);
  };

  return (
    <div 
      className="glass-panel rounded-2xl p-7 flex flex-col h-full relative group animate-fade-in-up"
      style={{ animationDelay: delay }}
    >
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-1/2 h-1 bg-gradient-to-r from-transparent via-indigo-500/50 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
      
      <div className="flex items-center gap-4 mb-4">
        <div className="w-12 h-12 bg-gradient-to-br from-slate-800 to-slate-900 border border-slate-700 rounded-xl flex items-center justify-center shadow-inner relative overflow-hidden group-hover:border-indigo-500/50 transition-colors">
          <div className="absolute inset-0 bg-indigo-500/10 opacity-0 group-hover:opacity-100 transition-opacity"></div>
          {icon}
        </div>
        <h3 className="text-lg font-[family-name:var(--font-display)] font-semibold text-slate-100">{title}</h3>
      </div>
      
      <p className="text-slate-400 text-sm mb-6 leading-relaxed flex-grow">{description}</p>
      
      <form onSubmit={handleSubmit} className="mt-auto">
        {children}
        <button
          type="submit"
          disabled={isSubmitting}
          className="btn-primary w-full py-2.5 mt-2 text-sm disabled:opacity-70 disabled:cursor-not-allowed"
        >
          {isSubmitting ? (
            <><Loader2 className="w-4 h-4 animate-spin" /> Processing...</>
          ) : (
            <>Execute Operation</>
          )}
        </button>
      </form>

      {response && (
        <div className={`mt-4 p-3 rounded-xl border text-xs relative overflow-hidden ${
          response.success 
            ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400' 
            : 'bg-red-500/10 border-red-500/20 text-red-400'
        }`}>
          <div className="font-semibold mb-1 flex items-center gap-1.5">
            {response.success ? <CheckCircle2 className="w-3.5 h-3.5" /> : <ShieldAlert className="w-3.5 h-3.5" />}
            {response.success ? 'Operation Successful' : 'Operation Failed'}
          </div>
          <div className="font-mono opacity-80 whitespace-pre-wrap max-h-32 overflow-y-auto custom-scrollbar">
            {response.message}
          </div>
        </div>
      )}
    </div>
  );
}

function JobItem({ job }: { job: IngestJob }) {
  const statusConfig = {
    pending: { color: 'text-amber-400', bg: 'bg-amber-400/10', border: 'border-amber-400/20' },
    running: { color: 'text-indigo-400', bg: 'bg-indigo-400/10', border: 'border-indigo-400/20' },
    completed: { color: 'text-emerald-400', bg: 'bg-emerald-400/10', border: 'border-emerald-400/20' },
    failed: { color: 'text-rose-400', bg: 'bg-rose-400/10', border: 'border-rose-400/20' },
    cancelled: { color: 'text-slate-400', bg: 'bg-slate-400/10', border: 'border-slate-400/20' },
  };

  const status = job.status.toLowerCase() as keyof typeof statusConfig;
  const config = statusConfig[status] || statusConfig.cancelled;

  return (
    <div className="group bg-slate-900/40 hover:bg-slate-800/60 border border-slate-700/50 hover:border-slate-600/80 rounded-xl p-5 transition-all">
      <div className="flex flex-col sm:flex-row justify-between items-start gap-4 mb-4">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <span className={`px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider border ${config.bg} ${config.color} ${config.border}`}>
              {job.status}
            </span>
            <span className="text-slate-500 text-xs font-mono">{job.id}</span>
          </div>
          <h4 className="text-slate-200 font-medium text-sm mt-2 flex items-center gap-2">
            {job.type} <span className="text-slate-600">•</span> {job.slug || 'System-wide'}
          </h4>
        </div>
        
        <div className="flex items-center gap-1.5 text-slate-500 text-xs">
          <Clock className="w-3.5 h-3.5" />
          {new Date(job.created_at).toLocaleString(undefined, { 
            month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' 
          })}
        </div>
      </div>
      
      {job.message && (
        <div className="text-sm text-slate-400 bg-slate-950/50 p-3 rounded-lg border border-slate-800/50 mt-3 font-mono text-xs">
          {job.message}
        </div>
      )}
      
      {job.progress !== undefined && job.status === 'running' && (
        <div className="mt-5">
          <div className="flex justify-between text-xs mb-1.5">
            <span className="text-slate-400 font-medium">Progress</span>
            <span className="text-indigo-400 font-mono font-bold">{job.progress}%</span>
          </div>
          <div className="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-indigo-500 to-purple-500 transition-all duration-500 relative"
              style={{ width: `${job.progress}%` }}
            >
              <div className="absolute inset-0 bg-white/20 animate-pulse"></div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
