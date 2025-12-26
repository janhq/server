'use client';

import { Check as CheckIcon, Copy as CopyIcon } from 'lucide-react';
import { useState } from 'react';

interface Example {
  title: string;
  language: string;
  code: string;
  highlightedHtml: string;
}

interface CodeExampleClientProps {
  examples: Example[];
}

export function CodeExampleClient({ examples }: CodeExampleClientProps) {
  const [selected, setSelected] = useState(0);
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(examples[selected].code);
      setCopied(true);
      setTimeout(() => setCopied(false), 1200);
    } catch (err) {
      setCopied(false);
    }
  };

  return (
    <div className="sticky top-24 rounded-2xl border overflow-hidden ">
      <div className="flex border-b ">
        {examples.map((example, idx) => (
          <button
            key={idx}
            onClick={() => setSelected(idx)}
            className={`px-4 py-3 text-sm font-medium transition-colors ${
              selected === idx
                ? 'bg-background border-b-2 border-blue-500'
                : 'text-gray-600 dark:text-gray-400 '
            }`}
          >
            {example.title}
          </button>
        ))}
      </div>
      <div className="relative py-6 bg-secondary/50">
        <button aria-label="Copy code" onClick={handleCopy} className="absolute top-4 right-4 ">
          <span className="inline-block relative size-4">
            <span
              className={`absolute inset-0 flex items-center justify-center ${
                copied ? 'scale-0 opacity-70' : 'scale-100 opacity-70'
              }`}
            >
              <CopyIcon />
            </span>
            <span
              className={`absolute inset-0 flex items-center justify-center  ${
                copied ? 'scale-100 opacity-100' : 'scale-0 opacity-0'
              }`}
            >
              <CheckIcon />
            </span>
          </span>
        </button>
        <div
          className="text-sm overflow-x-auto"
          dangerouslySetInnerHTML={{ __html: examples[selected].highlightedHtml }}
        />
      </div>
    </div>
  );
}
