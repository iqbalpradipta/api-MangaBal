import { redirect } from 'next/navigation';
import { isAuthenticated } from '@/lib/auth';

export default async function HomePage() {
  // Middleware will handle the redirect, but we can also do it here
  const authenticated = await isAuthenticated();
  
  if (authenticated) {
    redirect('/dashboard');
  } else {
    redirect('/login');
  }
}
