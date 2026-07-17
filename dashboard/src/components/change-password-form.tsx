'use client';

import { useState } from 'react';
import { KeyRound, Eye, EyeOff } from 'lucide-react';

export default function ChangePasswordForm() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [showOld, setShowOld] = useState(false);
  const [showNew, setShowNew] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError('');
    setSuccess('');

    const fd = new FormData(e.currentTarget);
    const old_password = fd.get('old_password') as string;
    const new_password = fd.get('new_password') as string;
    const confirm_password = fd.get('confirm_password') as string;

    if (new_password !== confirm_password) {
      setError('Password baru dan konfirmasi tidak sama');
      return;
    }
    if (new_password.length < 8) {
      setError('Password baru minimal 8 karakter');
      return;
    }

    setLoading(true);
    try {
      const res = await fetch('/dashboard/api/change-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ old_password, new_password }),
      });
      const data = await res.json();
      if (!data.success) {
        setError(data.message || 'Gagal mengubah password');
      } else {
        setSuccess('Password berhasil diubah');
        (e.target as HTMLFormElement).reset();
      }
    } catch {
      setError('Terjadi kesalahan, coba lagi');
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4 max-w-md">
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

      {/* Password Lama */}
      <div>
        <label className="block text-xs font-medium text-slate-400 mb-1">Password Lama</label>
        <div className="relative">
          <input
            name="old_password"
            type={showOld ? 'text' : 'password'}
            required
            placeholder="Masukkan password lama"
            className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 pr-10 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 text-sm"
          />
          <button
            type="button"
            onClick={() => setShowOld(v => !v)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
          >
            {showOld ? <EyeOff size={14} /> : <Eye size={14} />}
          </button>
        </div>
      </div>

      {/* Password Baru */}
      <div>
        <label className="block text-xs font-medium text-slate-400 mb-1">Password Baru</label>
        <div className="relative">
          <input
            name="new_password"
            type={showNew ? 'text' : 'password'}
            required
            minLength={8}
            placeholder="Minimal 8 karakter"
            className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 pr-10 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 text-sm"
          />
          <button
            type="button"
            onClick={() => setShowNew(v => !v)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
          >
            {showNew ? <EyeOff size={14} /> : <Eye size={14} />}
          </button>
        </div>
      </div>

      {/* Konfirmasi Password */}
      <div>
        <label className="block text-xs font-medium text-slate-400 mb-1">Konfirmasi Password Baru</label>
        <div className="relative">
          <input
            name="confirm_password"
            type={showConfirm ? 'text' : 'password'}
            required
            placeholder="Ulangi password baru"
            className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 pr-10 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 text-sm"
          />
          <button
            type="button"
            onClick={() => setShowConfirm(v => !v)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
          >
            {showConfirm ? <EyeOff size={14} /> : <Eye size={14} />}
          </button>
        </div>
      </div>

      <button
        type="submit"
        disabled={loading}
        className="w-full bg-indigo-600 hover:bg-indigo-700 disabled:bg-indigo-600/50 disabled:cursor-not-allowed text-white font-medium py-2.5 px-4 rounded-lg text-sm transition-colors flex items-center justify-center gap-2"
      >
        {loading ? (
          <>
            <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            Menyimpan...
          </>
        ) : (
          <>
            <KeyRound size={15} /> Ubah Password
          </>
        )}
      </button>
    </form>
  );
}
