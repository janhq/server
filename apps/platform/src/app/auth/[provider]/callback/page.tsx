'use client';

import { ProviderType } from '@/lib/auth/providers';
import { getSharedAuthService } from '@/lib/auth/service';
import { Loader2 } from 'lucide-react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { useEffect, useRef } from 'react';

export default function AuthCallbackPage() {
  const router = useRouter();
  const params = useParams();
  const searchParams = useSearchParams();
  const processedRef = useRef(false);

  useEffect(() => {
    if (processedRef.current) return;
    processedRef.current = true;

    const handleCallback = async () => {
      try {
        const providerId = params.provider as ProviderType;
        const code = searchParams.get('code') || '';
        const state = searchParams.get('state') || undefined;

        const authService = getSharedAuthService();
        await authService.handleProviderCallback(providerId, code, state);

        router.push('/docs');
      } catch (error) {
        console.error('Auth callback failed:', error);
        router.push('/');
      }
    };

    handleCallback();
  }, [params, searchParams, router]);

  return (
    <div className="flex h-screen items-center justify-center">
      <Loader2 className="h-8 w-8 animate-spin" />
    </div>
  );
}
