import { LandingPage } from '@/components/landing/landing-page';
import { source } from '@/lib/source';
import { getMDXComponents } from '@/mdx-components';
import type { TOCItemType } from 'fumadocs-core/toc';
import { DocsBody, DocsDescription, DocsPage, DocsTitle } from 'fumadocs-ui/page';
import type { MDXContent } from 'mdx/types';
import { notFound } from 'next/navigation';

interface DocsPageProps {
  params: Promise<{ slug?: string[] }>;
}

export default async function Page(props: DocsPageProps) {
  const params = await props.params;
  const page = source.getPage(params.slug);

  if (!page) notFound();

  // Type assertion for MDX properties
  const pageData = page.data as typeof page.data & {
    body: MDXContent;
    toc: TOCItemType[];
    full?: boolean;
  };

  // Check if this is the overview page
  const isOverviewPage = params.slug && params.slug.length === 1 && params.slug[0] === 'overview';

  // If it's the overview page, render the custom landing page with full width
  if (isOverviewPage) {
    return (
      <DocsPage toc={[]} full>
        <LandingPage />
      </DocsPage>
    );
  }

  const MDX = pageData.body;

  return (
    <DocsPage toc={pageData.toc} full={pageData.full}>
      <DocsTitle>{pageData.title}</DocsTitle>
      <DocsDescription>{pageData.description}</DocsDescription>
      <DocsBody>
        <MDX components={getMDXComponents()} />
      </DocsBody>
    </DocsPage>
  );
}

export async function generateStaticParams() {
  return source.generateParams();
}

export async function generateMetadata(props: DocsPageProps) {
  const params = await props.params;
  const page = source.getPage(params.slug);
  if (!page) notFound();

  return {
    title: `${page.data.title} - Jan Platform`,
    description: page.data.description,
  };
}
