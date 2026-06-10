import { redirect } from 'next/navigation';
import { setSession, isAuthenticated } from '@/lib/auth';
import { LockKeyhole, ArrowRight, ShieldCheck } from 'lucide-react';

export default async function LoginPage() {
  // Check if already authenticated
  if (await isAuthenticated()) {
    redirect('/home');
  }

  async function handleLogin(formData: FormData) {
    'use server';
    
    const password = formData.get('password') as string;
    const expectedPassword = process.env.DASHBOARD_PASSWORD;

    if (!password || !expectedPassword || password !== expectedPassword) {
      redirect('/login?error=1');
    }

    await setSession(password);
    redirect('/home');
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-6 relative overflow-hidden">
      {/* Decorative Background Elements */}
      <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-indigo-600/20 rounded-full blur-[120px] pointer-events-none"></div>
      <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-purple-600/20 rounded-full blur-[120px] pointer-events-none"></div>

      <div className="w-full max-w-[420px] animate-fade-in-up relative z-10">
        {/* Logo / Brand */}
        <div className="flex flex-col items-center mb-8 text-center">
          <div className="w-16 h-16 bg-gradient-to-tr from-indigo-500 to-purple-500 rounded-2xl flex items-center justify-center shadow-lg shadow-indigo-500/30 mb-6 rotate-3 hover:rotate-0 transition-transform duration-300">
            <ShieldCheck className="w-8 h-8 text-white" strokeWidth={2.5} />
          </div>
          <h1 className="text-3xl font-[family-name:var(--font-display)] font-bold text-white tracking-tight mb-2">
            Admin Access
          </h1>
          <p className="text-slate-400 text-sm">
            Authenticate to access the Manga API dashboard
          </p>
        </div>

        {/* Login Card */}
        <div className="glass-panel rounded-3xl p-8 sm:p-10 relative overflow-hidden">
          {/* subtle top border highlight */}
          <div className="absolute top-0 left-0 right-0 h-[1px] bg-gradient-to-r from-transparent via-indigo-500/50 to-transparent"></div>
          
          <ErrorMessage />

          <form action={handleLogin} className="space-y-6">
            <div className="space-y-2">
              <label htmlFor="password" className="block text-sm font-medium text-slate-300 ml-1">
                Security Passkey
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-500">
                  <LockKeyhole className="w-5 h-5" />
                </div>
                <input
                  type="password"
                  id="password"
                  name="password"
                  required
                  autoComplete="current-password"
                  placeholder="Enter your admin password"
                  className="glass-input w-full pl-12 pr-4 py-3.5 rounded-xl text-[15px]"
                />
              </div>
            </div>

            <button
              type="submit"
              className="btn-primary w-full py-3.5 mt-2 font-semibold text-[15px]"
            >
              <span>Authenticate</span>
              <ArrowRight className="w-4 h-4 ml-1" />
            </button>
          </form>
        </div>
        
        <p className="text-center text-slate-500 text-xs mt-8">
          &copy; {new Date().getFullYear()} Manga API. Secure Administrative Panel.
        </p>
      </div>
    </div>
  );
}

function ErrorMessage() {
  return (
    <div className="mb-6">
      <noscript>
        <div className="bg-red-500/10 border border-red-500/20 text-red-400 p-4 rounded-xl text-sm flex items-start gap-3">
          <div className="mt-0.5">⚠️</div>
          <div>Invalid password. Please try again.</div>
        </div>
      </noscript>
      <script dangerouslySetInnerHTML={{
        __html: `
          const urlParams = new URLSearchParams(window.location.search);
          if (urlParams.get('error') === '1') {
            document.write('<div class="bg-red-500/10 border border-red-500/20 text-red-400 p-4 rounded-xl text-sm flex items-start gap-3 animate-shake mb-6"><div class="mt-0.5">⚠️</div><div>Invalid password. Please try again.</div></div>');
          }
        `
      }} />
    </div>
  );
}
