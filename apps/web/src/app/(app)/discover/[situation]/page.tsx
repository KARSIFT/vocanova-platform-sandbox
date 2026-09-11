import Link from "next/link";
import { notFound } from "next/navigation";

import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer } from "@/ui/surface";

import { getSituationDetailView } from "./_components/situation-view";

interface SituationDiscoverPageProps {
  params: Promise<{ situation: string }>;
}

export default async function SituationDiscoverPage({
  params,
}: SituationDiscoverPageProps) {
  const { situation } = await params;
  const client = await createServerApiClient();
  let response: Awaited<ReturnType<typeof client.getJourneySituation>>;
  try {
    response = await client.getJourneySituation(situation);
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 404) {
      notFound();
    }
    requireAuthRedirect(error, `/discover/${situation}`);
  }

  const { situation: situationData, meanings } = response.data;
  const savedCount = meanings.filter((meaning) => meaning.saved).length;
  const nextMeaning = meanings.find((meaning) => !meaning.saved);

  return (
    <PageContainer>
      <Link
        href="/discover"
        className="inline-flex min-h-11 items-center text-base font-semibold text-primary-700 hover:text-primary-800"
      >
        Back to Journey
      </Link>
      <Eyebrow>
        {situationData.category.replaceAll("_", " ")}
        {situationData.levelBand ? ` · ${situationData.levelBand}` : ""}
      </Eyebrow>
      <h1 className="mt-[var(--spacing-xs)] text-3xl font-bold tracking-tight text-neutral-900">
        {situationData.title}
      </h1>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        {situationData.shortDescription}
      </p>

      {meanings.length > 0 ? (
        <section
          aria-label="Situation progress"
          className="mt-[var(--spacing-md)] rounded-[var(--radius-lg)] border border-secondary-100 bg-secondary-50 p-[var(--spacing-md)]"
        >
          <p className="text-sm font-semibold text-secondary-900">
            {savedCount} of {meanings.length}{" "}
            {meanings.length === 1 ? "word" : "words"} saved
          </p>
          <p className="mt-[var(--spacing-xs)] text-sm text-neutral-700">
            {nextMeaning
              ? "Choose one useful word to save, then practice it in a sentence when you’re ready."
              : "You’ve saved every word in this situation. Review them when you’re ready."}
          </p>
          <div className="mt-[var(--spacing-sm)] flex flex-wrap gap-[var(--spacing-sm)]">
            {nextMeaning ? (
              <Link
                href={`/discover/${situation}/${nextMeaning.wordSlug}`}
                className="inline-flex min-h-11 items-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-white hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
              >
                Continue with {nextMeaning.wordText}
              </Link>
            ) : null}
            {savedCount > 0 ? (
              <Link
                href="/words"
                className="inline-flex min-h-11 items-center rounded-md px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
              >
                View saved vocabulary
              </Link>
            ) : null}
          </div>
        </section>
      ) : null}

      {getSituationDetailView(meanings.length) === "empty" ? (
        <div className="flex flex-col items-center justify-center py-[var(--spacing-2xl)] text-center">
          <h2 className="text-xl font-semibold text-neutral-900">
            No words here yet
          </h2>
          <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
            We&apos;re still adding vocabulary for this situation. Try another
            situation for now.
          </p>
          <Link
            href="/discover"
            className="mt-[var(--spacing-lg)] inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Browse other situations
          </Link>
        </div>
      ) : (
        <ul className="mt-[var(--spacing-lg)] space-y-[var(--spacing-md)]">
          {meanings.map((meaning) => (
            <li key={meaning.meaningId}>
              <Link
                href={`/discover/${situation}/${meaning.wordSlug}`}
                className="block rounded-[var(--radius-lg)] border border-neutral-200 bg-white p-[var(--spacing-md)] shadow-sm transition hover:border-primary-300 hover:shadow-md"
              >
                <div className="flex items-start justify-between gap-[var(--spacing-md)]">
                  <div>
                    <h2 className="text-lg font-semibold text-neutral-900">
                      {meaning.wordText}
                    </h2>
                    <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
                      {meaning.shortDefinition}
                    </p>
                  </div>
                  {meaning.saved ? (
                    <span className="shrink-0 rounded-full bg-primary-100 px-[var(--spacing-sm)] py-[var(--spacing-xs)] text-sm font-semibold text-primary-800">
                      <span aria-hidden="true">✓</span> Saved
                    </span>
                  ) : null}
                </div>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </PageContainer>
  );
}
