import { Rocket } from 'lucide-react';
import Link from 'next/link';

interface ComingSoonProps {
  title: string;
  description?: string;
}

export function ComingSoon({ title, description }: ComingSoonProps) {
  return (
    <div className="flex h-full flex-col items-center justify-center p-8">
      <div className="mb-8 rounded-full bg-primary/10 p-6">
        <Rocket className="h-16 w-16 text-primary" />
      </div>
      <h1 className="mb-4 text-3xl font-bold">{title}</h1>
      <p className="mb-8 max-w-md text-center text-muted-foreground">
        {description || `We're working hard to bring you this feature. Stay tuned for updates!`}
      </p>
      <div className="flex gap-4">
        <Link
          href="/docs"
          className="inline-flex items-center justify-center rounded-md bg-primary px-6 py-3 text-sm font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90"
        >
          Back to Docs
        </Link>
        <Link
          href="/docs/quickstart"
          className="inline-flex h-10 items-center justify-center rounded-md bg-primary px-8 text-sm font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50"
        >
          View Documentation
        </Link>
      </div>
    </div>
  );
}
