import { cookies } from 'next/headers';

const SESSION_COOKIE_NAME = 'manga_admin_session';
const SESSION_MAX_AGE = 60 * 60 * 24; // 24 hours

export async function setSession(password: string) {
  const cookieStore = await cookies();
  cookieStore.set(SESSION_COOKIE_NAME, password, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    maxAge: SESSION_MAX_AGE,
    path: '/dashboard',
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
  const expectedPassword = process.env.DASHBOARD_PASSWORD;
  
  if (!expectedPassword) {
    return false;
  }
  
  return session === expectedPassword;
}
