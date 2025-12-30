'use client';

import { ArrowRight, Box, Code2, Terminal } from 'lucide-react';
import Link from 'next/link';

export default function HomePage() {
  return (
    <div className="flex flex-col min-h-[calc(100vh-3.5rem)]">
      <main className="flex-1">
        <section className="space-y-6 pb-8 pt-6 md:pb-12 md:pt-10 lg:py-32">
          <div className="container flex flex-col items-center gap-4 text-center">
            <h1 className="font-heading text-3xl sm:text-5xl md:text-6xl lg:text-7xl">
              Jan Platform
            </h1>
            <p className="leading-normal text-muted-foreground sm:text-xl sm:leading-8">
              Self-hosted agentic AI platform powered by local models
            </p>
            <div className="space-x-4">
              <Link
                href="/docs/quickstart"
                className="inline-flex h-11 items-center justify-center rounded-md bg-primary px-8 text-sm font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50"
              >
                Get Started
              </Link>
              <Link
                href="/docs/api-reference"
                className="inline-flex h-11 items-center justify-center rounded-md border border-input bg-background px-8 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50"
              >
                API Reference
              </Link>
            </div>
          </div>
        </section>

        <section className="container space-y-6 py-8 md:py-12 lg:py-24">
          <div className="mx-auto grid justify-center gap-4 sm:grid-cols-2 md:grid-cols-3">
            <Link
              href="/docs/quickstart"
              className="relative overflow-hidden rounded-lg border bg-background p-2 transition-colors hover:bg-accent hover:text-accent-foreground group"
            >
              <div className="flex h-[180px] flex-col justify-between rounded-md p-6">
                <Code2 className="h-12 w-12" />
                <div className="space-y-2">
                  <h3 className="font-bold flex items-center gap-2">
                    Quickstart
                    <ArrowRight className="h-4 w-4 opacity-0 group-hover:opacity-100 transition-opacity" />
                  </h3>
                  <p className="text-sm text-muted-foreground">
                    Get started with Jan Server in minutes
                  </p>
                </div>
              </div>
            </Link>
            <Link
              href="/docs/architecture"
              className="relative overflow-hidden rounded-lg border bg-background p-2 transition-colors hover:bg-accent hover:text-accent-foreground group"
            >
              <div className="flex h-[180px] flex-col justify-between rounded-md p-6">
                <Box className="h-12 w-12" />
                <div className="space-y-2">
                  <h3 className="font-bold flex items-center gap-2">
                    Architecture
                    <ArrowRight className="h-4 w-4 opacity-0 group-hover:opacity-100 transition-opacity" />
                  </h3>
                  <p className="text-sm text-muted-foreground">
                    Detailed system architecture and design decisions
                  </p>
                </div>
              </div>
            </Link>
            <Link
              href="/docs/api-reference"
              className="relative overflow-hidden rounded-lg border bg-background p-2 transition-colors hover:bg-accent hover:text-accent-foreground group"
            >
              <div className="flex h-[180px] flex-col justify-between rounded-md p-6">
                <Terminal className="h-12 w-12" />
                <div className="space-y-2">
                  <h3 className="font-bold flex items-center gap-2">
                    API Reference
                    <ArrowRight className="h-4 w-4 opacity-0 group-hover:opacity-100 transition-opacity" />
                  </h3>
                  <p className="text-sm text-muted-foreground">
                    Explore the API reference documentation.
                  </p>
                </div>
              </div>
            </Link>
          </div>
        </section>
      </main>
    </div>
  );
}
