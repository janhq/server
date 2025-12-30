import type { NextRequest } from 'next/server';
import { NextResponse } from 'next/server';

export function proxy(request: NextRequest) {
  const authStorage = request.cookies.get('auth-storage');

  // Parse the auth storage cookie to check if user is logged in
  let isLoggedIn = false;
  if (authStorage) {
    try {
      const authData = JSON.parse(authStorage.value);
      // Check for direct isLoggedIn property (new structure)
      isLoggedIn = authData?.isLoggedIn === true;
    } catch (error) {
      // Invalid auth data, treat as not logged in
      isLoggedIn = false;
    }
  }

  const { pathname } = request.nextUrl;

  // Protected routes - require authentication
  if (pathname.startsWith('/app')) {
    if (!isLoggedIn) {
      const loginUrl = new URL('/', request.url);
      return NextResponse.redirect(loginUrl);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - docs (documentation pages)
     */
    '/((?!api|_next/static|_next/image|favicon.ico|docs).*)',
  ],
};
