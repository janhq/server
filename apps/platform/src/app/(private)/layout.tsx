import { Navbar } from '@/components/navbar';

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <Navbar />
      <main className="pt-14 min-h-screen">
        <div className="flex flex-1 w-full mx-auto max-w-(--fd-page-width) pt-(--fd-tocnav-height) pe-(--fd-toc-width)">
          <div className="flex min-w-0 w-full flex-col gap-4 pt-8 px-4 md:px-6 md:mx-auto">
            {children}
          </div>
        </div>
      </main>
    </>
  );
}
