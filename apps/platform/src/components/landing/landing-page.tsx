import { Code2, Lock, MessageSquare, Zap } from 'lucide-react';
import Link from 'next/link';
import { CodeExample } from './code-example';
import { FeatureCard } from './feature-card';
import { Hero } from './hero';
import { ModelCard } from './model-card';

export function LandingPage() {
  return (
    <div>
      {/* Hero Section */}
      <Hero />

      {/* Main Content with Code Example */}
      <div className="relative">
        <div className="mx-auto max-w-7xl px-6 lg:px-8">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-12">
            {/* Left Content - 2/3 width */}
            <div className="lg:col-span-2 space-y-16">
              {/* Core Concepts */}
              <section className="flex flex-col md:flex-row md:items-start md:justify-between gap-12 w-full">
                <div className="w-full md:w-1/2 flex-shrink-0">
                  <h2 className="text-3xl tracking-tight text-gray-900 dark:text-white mb-10">
                    Developer quickstart
                  </h2>
                  <p className="text-lg text-gray-700 dark:text-gray-300 mb-6">
                    Make your first API request in minutes. Learn the basics of the Jan platform.
                  </p>
                  <Link
                    href="/docs/quickstart"
                    className="inline-block bg-gray-800 hover:bg-gray-700 transition-colors text-gray-100 px-6 py-3 rounded-full"
                  >
                    Get started
                  </Link>
                </div>
                <div className="flex-2 hidden md:block">
                  <CodeExample />
                </div>
              </section>

              {/* Core Concepts Section */}
              <section>
                <h2 className="text-3xl tracking-tight text-gray-900 dark:text-white mb-8">
                  Core concepts
                </h2>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                  <FeatureCard
                    title="Responses"
                    description="Handle streaming and batch responses with full control over generation"
                    icon={Zap}
                    href="/docs/api-reference/responses"
                  />
                  <FeatureCard
                    title="Authentication"
                    description="Secure your applications with OAuth, API keys, and token management"
                    icon={Lock}
                    href="/docs/api-reference/authentication"
                  />
                </div>
              </section>

              {/* Agents */}
              <section>
                <h2 className="text-3xl tracking-tight text-gray-900 dark:text-white mb-8">
                  Agents
                </h2>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                  <FeatureCard
                    title="Responses"
                    description="Build conversational agents with full context awareness"
                    icon={MessageSquare}
                    href="/docs/api-reference/responses"
                  />
                </div>
              </section>

              {/* Tools */}
              <section>
                <h2 className="text-3xl tracking-tight text-gray-900 dark:text-white mb-8">
                  Tools
                </h2>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                  <FeatureCard
                    title="API Keys"
                    description="Generate and manage API keys for secure access"
                    icon={Code2}
                    href="/docs/api-reference/authentication"
                  />
                </div>
              </section>

              {/* Models Gallery */}
              <section>
                <h2 className="text-3xl tracking-tight text-gray-900 dark:text-white mb-4">
                  Browse models
                </h2>
                <p className="text-lg text-gray-600 dark:text-gray-400 mb-8">
                  Choose from a wide range of AI models optimized for different tasks
                </p>
                <div className="grid grid-cols-1 gap-4">
                  <ModelCard
                    name="Jan-v1"
                    description="Our most advanced model for complex tasks and deep understanding"
                    badge="Recommended"
                  />
                  <ModelCard
                    name="Qwen3-Thinking"
                    description="Optimized for rapid responses and everyday conversations"
                  />
                  <ModelCard
                    name="QwenCoder 30B"
                    description="Excels at programming, logic, and multi-step reasoning"
                  />
                  <ModelCard
                    name="Gemma3 27B"
                    description="Ideal for deploying and scaling custom fine-tuned models"
                  />
                </div>
              </section>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
