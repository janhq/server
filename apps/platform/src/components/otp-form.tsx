'use client';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field';
import { InputOTP, InputOTPGroup, InputOTPSlot } from '@/components/ui/input-otp';
import { useAuthStore } from '@/store/auth-store';
import { useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';

// Dummy OTP for testing (in production, this would be verified on the backend)
const DUMMY_OTP = '123456';

export function OTPForm({ ...props }: React.ComponentProps<typeof Card>) {
  const login = useAuthStore((state) => state.login);
  const searchParams = useSearchParams();
  const [otp, setOtp] = useState('');
  const [email, setEmail] = useState('');
  const [countdown, setCountdown] = useState(60);
  const [canResend, setCanResend] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    // Get email from URL params
    const emailParam = searchParams.get('email');
    if (emailParam) {
      setEmail(emailParam);
    }

    // Start countdown timer
    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          setCanResend(true);
          clearInterval(timer);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [searchParams]);

  const handleResend = () => {
    if (!canResend) return;

    // Simulate sending OTP again
    console.log('Resending OTP to:', email);

    // Reset countdown
    setCountdown(60);
    setCanResend(false);
    setError('');

    // Start new countdown
    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          setCanResend(true);
          clearInterval(timer);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    // Verify OTP (using dummy value for now)
    if (otp === DUMMY_OTP) {
      // OTP is correct, log the user in
      login({
        email: email,
      });

      // Redirect to docs
      window.location.href = '/docs';
    } else {
      setError('Invalid OTP. Please try again.');
    }
  };

  return (
    <Card {...props}>
      <CardHeader className="text-center">
        <CardTitle className="text-xl">Enter verification code</CardTitle>
        <CardDescription>
          We sent a 6-digit code to{' '}
          <span className="font-medium text-foreground">{email || 'your email'}</span>. Please check
          your mail.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="otp" className="sr-only">
                Verification code
              </FieldLabel>
              <InputOTP maxLength={6} id="otp" required value={otp} onChange={setOtp}>
                <InputOTPGroup className="gap-2.5 *:data-[slot=input-otp-slot]:rounded-md *:data-[slot=input-otp-slot]:border">
                  <InputOTPSlot index={0} />
                  <InputOTPSlot index={1} />
                  <InputOTPSlot index={2} />
                  <InputOTPSlot index={3} />
                  <InputOTPSlot index={4} />
                  <InputOTPSlot index={5} />
                </InputOTPGroup>
              </InputOTP>
              <FieldDescription className="text-center">
                Enter the 6-digit code sent to your email.
              </FieldDescription>
              {error && (
                <FieldDescription className="text-center text-red-500">{error}</FieldDescription>
              )}
            </Field>
            <Button type="submit" disabled={otp.length !== 6}>
              Verify
            </Button>
            <FieldDescription className="text-center">
              Didn&apos;t receive the code?{' '}
              {canResend ? (
                <button
                  type="button"
                  onClick={handleResend}
                  className="font-medium text-primary hover:underline"
                >
                  Resend
                </button>
              ) : (
                <span className="text-muted-foreground">Resend in {countdown}s</span>
              )}
            </FieldDescription>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  );
}
