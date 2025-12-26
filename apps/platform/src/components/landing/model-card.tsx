interface ModelCardProps {
  name: string;
  description: string;
  badge?: string;
}

export function ModelCard({ name, description, badge }: ModelCardProps) {
  return (
    <div className="rounded-xl border p-6 transition-all bg-secondary/40">
      <div className="flex items-center gap-1 mb-3">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{name}</h3>
        {badge && (
          <span className="px-2.5 py-1 text-xs font-medium rounded-full bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300">
            {badge}
          </span>
        )}
      </div>
      <p className="text-sm text-muted-foreground">{description}</p>
    </div>
  );
}
