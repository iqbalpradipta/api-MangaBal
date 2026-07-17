import { cookies } from 'next/headers';
import fs from 'fs';

const SESSION_COOKIE_NAME = 'manga_admin_session';
const SESSION_MAX_AGE = 60 * 60 * 24; // 24 hours
const PASSWORD_FILE = '/data/password';

// getStoredPassword baca dari file dulu, fallback ke env
export function getStoredPassword(): string {
  try {
    const content = fs.readFileSync(PASSWORD_FILE, 'utf-8').trim();
    if (content.length > 0) return content;
  } catch {
    // file tidak ada → pakai env
  }
  return process.env.DASHBOARD_PASSWORD ?? '';
}

export async function setSession(password: string) {
  const cookieStore = await cookies();
  cookieStore.set(SESSION_COOKIE_NAME, password, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    maxAge: SESSION_MAX_AGE,
    path: '/',
  });
}

export async function getSession(): Promise<string | undefined> {
  const cookieStore = await cookies();
  return cookieStore.get(SESSION_COOKIE_NAME)?.value;
}

export async function clearSession() {
  const cookieStore = await cookies();
  cookieStore.delete(SESSION_COOKIE_NAME);
}

export async function isAuthenticated(): Promise<boolean> {
  const session = await getSession();
  const expected = getStoredPassword();
  if (!expected) return false;
  return session === expected;
}

export async function changePassword(oldPassword: string, newPassword: string): Promise<{ ok: boolean; error?: string }> {
  const expected = getStoredPassword();
  if (oldPassword !== expected) {
    return { ok: false, error: 'Password lama salah' };
  }
  if (newPassword.length < 8) {
    return { ok: false, error: 'Password baru minimal 8 karakter' };
  }
  try {
    fs.mkdirSync('/data', { recursive: true });
    fs.writeFileSync(PASSWORD_FILE, newPassword, { encoding: 'utf-8', mode: 0o600 });
    return { ok: true };
  } catch (e) {
    return { ok: false, error: `Gagal menyimpan password: ${e}` };
  }
}
