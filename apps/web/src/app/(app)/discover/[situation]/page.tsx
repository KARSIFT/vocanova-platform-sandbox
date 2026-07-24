import Link from "next/link";
import { notFound } from "next/navigation";

import {
  MOCK_SITUATION_WORD_LISTS,
  type SituationSlug,
} from "./_lib/mock-word-data";

export default async function SituationDiscoverPage({
  params,
}: {
  params: Promise<{ situation: string }>;
}) {
  const { situation } = await params;
  const situationWords = MOCK_SITUATION_WORD_LISTS[situation as SituationSlug];

  if (!situationWords) {
    notFound();
  }

  return (
    <div className="p-[var(--spacing-lg)]">
      <Link
        href="/discover"
        className="text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
      >
        Back to Journey
      </Link>
      <h1 className="mt-[var(--spacing-md)] text-2xl font-semibold text-neutral-900">
        {situationWords.title}
      </h1>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        Words already saved are marked below.
      </p>

      <ul className="mt-[var(--spacing-lg)] space-y-[var(--spacing-md)]">
        {situationWords.words.map((word) => (
          <li key={word.wordSlug}>
            <Link
              href={`/discover/${situation}/${word.wordSlug}`}
              className="block rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm hover:border-primary-300 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-primary-600"
            >
              <div className="flex items-start justify-between gap-[var(--spacing-md)]">
                <div>
                  <p className="text-lg font-semibold text-neutral-900">
                    {word.term}
                  </p>
                  <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
                    {word.meaning}
                  </p>
                </div>
                {word.isSaved ? (
                  <span className="shrink-0 rounded-full bg-primary-100 px-[var(--spacing-sm)] py-[var(--spacing-xs)] text-sm font-semibold text-primary-800">
                    <span aria-hidden="true">✓</span> Saved
                  </span>
                ) : null}
              </div>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
