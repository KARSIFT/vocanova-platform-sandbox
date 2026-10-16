import Link from "next/link";
import { PageContainer } from "@/ui/surface";
import KnowledgeMapPage from "../knowledge-map/page";

export default async function JourneyPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  return (
    <>
      <KnowledgeMapPage searchParams={searchParams} />
      <PageContainer className="max-w-[72rem] pt-0">
        <nav
          aria-label="More ways to learn"
          className="flex flex-wrap gap-x-6 border-t border-neutral-200 pt-3"
        >
          <Link
            href="/courses"
            className="inline-flex min-h-11 items-center font-semibold text-primary-700"
          >
            Situation courses
          </Link>
          <Link
            href="/stories"
            className="inline-flex min-h-11 items-center font-semibold text-primary-700"
          >
            Short stories
          </Link>
          <Link
            href="/lists"
            className="inline-flex min-h-11 items-center font-semibold text-primary-700"
          >
            Personal lists
          </Link>
        </nav>
      </PageContainer>
    </>
  );
}
