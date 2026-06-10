import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const SESSION_COOKIE_NAME = 'manga_admin_session';
const LOGIN_PATH = '/login';
const DASHBOARD_PATH = '/dashboard';

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const session = request.cookies.get(SESSION_COOKIE_NAME)?.value;
  const expectedPassword = process.env.DASHBOARD_PASSWORD;
  
  const isAuthenticated = session && expectedPassword && session === expectedPassword;

  // Redirect authenticated users away from login page
  if (pathname === LOGIN_PATH && isAuthenticated) {
    return NextResponse.redirect(new URL(DASHBOARD_PATH, request.url));
  }

  // Protect dashboard route
  if (pathname.startsWith(DASHBOARD_PATH) && pathname !== LOGIN_PATH && !isAuthenticated) {
    return NextResponse.redirect(new URL(LOGIN_PATH, request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/', '/login'],
};
