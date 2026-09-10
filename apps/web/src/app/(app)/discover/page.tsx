import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer } from "@/ui/surface";

import { getDiscoverListView } from "./_components/discover-view";

export default async function DiscoverPage() {
  const client = await createServerApiClient();
  let response: Awaited<ReturnType<typeof client.listJourneySituations>>;
  try {
    response = await client.listJourneySituations();
  } catch (error) {
    requireAuthRedirect(error, "/discover");
  }

  const { items } = response.data;

  return (
    <PageContainer>
      <Eyebrow>Learn in context</Eyebrow>
      <h1 className="mt-[var(--spacing-xs)] text-3xl font-bold tracking-tight text-neutral-900">
        Journey
      </h1>
      <div className="flex flex-wrap items-end justify-between gap-[var(--spacing-md)]">
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          Choose a familiar moment, then collect words you will actually use.
        </p>
        <Link
          href="/words"
          className="inline-flex min-h-[var(--spacing-2xl)] items-center rounded-md px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          View saved vocabulary
        </Link>
      </div>

      {getDiscoverListView(items.length) === "empty" ? (
        <div className="flex flex-col items-center justify-center py-[var(--spacing-2xl)] text-center">
          <h2 className="text-xl font-semibold text-neutral-900">
            New situations are on the way
          </h2>
          <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
            We don&apos;t have any situations to explore just yet. Check back
            soon, or head to your reviews in the meantime.
          </p>
          <Link
            href="/review"
            className="mt-[var(--spacing-lg)] inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Go to Reviews
          </Link>
        </div>
      ) : (
        <ul className="mt-[var(--spacing-lg)] grid grid-cols-1 gap-[var(--spacing-md)] sm:grid-cols-2">
          {items.map((situation) => (
            <li key={situation.slug}>
              <Link
                href={`/discover/${situation.slug}`}
                className="group block min-h-32 rounded-[var(--radius-lg)] border border-neutral-200 bg-white p-[var(--spacing-lg)] shadow-sm transition hover:-translate-y-0.5 hover:border-primary-300 hover:shadow-md"
              >
                <span className="inline-flex rounded-lg bg-secondary-100 px-[var(--spacing-sm)] py-1 text-xs font-bold text-secondary-800">
                  {situation.category.replaceAll("_", " ")}
                  {situation.levelBand ? ` · ${situation.levelBand}` : ""}
                </span>
                <h2 className="mt-[var(--spacing-sm)] text-xl font-bold tracking-tight text-neutral-900 group-hover:text-primary-800">
                  {situation.title}
                </h2>
                <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
                  {situation.shortDescription}
                </p>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </PageContainer>
  );
}
