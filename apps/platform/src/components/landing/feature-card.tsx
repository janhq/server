import { LucideIcon } from 'lucide-react';
import Link from 'next/link';

interface FeatureCardProps {
  title: string;
  description: string;
  icon: LucideIcon;
  href: string;
}

export function FeatureCard({ title, description, icon: Icon, href }: FeatureCardProps) {
  return (
    <Link href={href} className="group relative rounded-2xl border bg-secondary/40 p-6">
      <div className="flex items-center gap-4 mb-4">
        <div className="rounded-lg border p-2">
          <Icon className="size-4 text-gray-900 dark:text-white" />
        </div>
      </div>
      <h3 className="text-lg font-semibold mb-2 transition-colors">{title}</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">{description}</p>
    </Link>
  );
}
