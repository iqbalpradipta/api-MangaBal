import { NextRequest, NextResponse } from 'next/server';
import { isAuthenticated, changePassword, setSession } from '@/lib/auth';

export async function POST(req: NextRequest) {
  if (!(await isAuthenticated())) {
    return NextResponse.json({ success: false, message: 'Unauthorized' }, { status: 401 });
  }

  const { old_password, new_password } = await req.json();

  if (!old_password || !new_password) {
    return NextResponse.json({ success: false, message: 'Field tidak boleh kosong' }, { status: 400 });
  }

  const result = await changePassword(old_password, new_password);
  if (!result.ok) {
    return NextResponse.json({ success: false, message: result.error }, { status: 400 });
  }

  // update session cookie dengan password baru supaya tidak logout
  await setSession(new_password);

  return NextResponse.json({ success: true, message: 'Password berhasil diubah' });
}
