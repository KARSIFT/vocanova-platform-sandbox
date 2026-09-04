import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";

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
    <div className="p-[var(--spacing-lg)]">
      <h1 className="text-2xl font-semibold text-neutral-900">Journey</h1>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        Choose a situation to explore practical vocabulary.
      </p>

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
            href="/reviews"
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
                className="block rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm hover:border-primary-300 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-primary-600"
              >
                <h2 className="text-lg font-semibold text-neutral-900">
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
    </div>
  );
}
