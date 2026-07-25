import Link from "next/link";
import { notFound } from "next/navigation";

import {
  MOCK_SITUATION_WORD_LISTS,
  type SituationSlug,
} from "../_lib/mock-word-data";

export default async function WordDetailPage({
  params,
}: {
  params: Promise<{ situation: string; word: string }>;
}) {
  const { situation, word } = await params;
  const situationWords = MOCK_SITUATION_WORD_LISTS[situation as SituationSlug];

  if (!situationWords) {
    notFound();
  }

  const wordDetails = situationWords.words.find(
    (candidate) => candidate.wordSlug === word,
  );

  if (!wordDetails) {
    notFound();
  }

  return (
    <div className="p-[var(--spacing-lg)]">
      <Link
        href={`/discover/${situation}`}
        className="text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
      >
        Back to {situationWords.title}
      </Link>

      <div className="mt-[var(--spacing-md)] flex flex-wrap items-start justify-between gap-[var(--spacing-md)]">
        <div>
          <h1 className="text-2xl font-semibold text-neutral-900">
            {wordDetails.term}
          </h1>
          <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
            Placeholder word details for VOC-022 static UI; replace with real
            API wiring in a follow-up package.
          </p>
        </div>
        <button
          type="button"
          disabled
          className="min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] rounded-md border border-neutral-200 bg-neutral-50 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-700"
        >
          {wordDetails.isSaved ? "Saved" : "Not saved"} (mock state)
        </button>
      </div>

      <section className="mt-[var(--spacing-lg)]">
        <h2 className="text-xl font-semibold text-neutral-900">Meanings</h2>
        <ul className="mt-[var(--spacing-sm)] space-y-[var(--spacing-sm)]">
          {wordDetails.meanings.map((meaning) => (
            <li key={`${meaning.partOfSpeech}-${meaning.definition}`}>
              <p className="font-medium text-neutral-900">
                {meaning.partOfSpeech}
              </p>
              <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
                {meaning.definition}
              </p>
            </li>
          ))}
        </ul>
      </section>

      <section className="mt-[var(--spacing-lg)]">
        <h2 className="text-xl font-semibold text-neutral-900">
          Example sentences
        </h2>
        <ul className="mt-[var(--spacing-sm)] list-disc space-y-[var(--spacing-sm)] pl-[var(--spacing-lg)] text-base text-neutral-700">
          {wordDetails.examples.map((example) => (
            <li key={example}>{example}</li>
          ))}
        </ul>
      </section>

      <section className="mt-[var(--spacing-lg)]">
        <h2 className="text-xl font-semibold text-neutral-900">Usage notes</h2>
        <ul className="mt-[var(--spacing-sm)] space-y-[var(--spacing-md)]">
          <li>
            <h3 className="text-lg font-semibold text-neutral-900">
              Collocation
            </h3>
            <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
              {wordDetails.usageNotes.collocation}
            </p>
          </li>
          <li>
            <h3 className="text-lg font-semibold text-neutral-900">Register</h3>
            <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
              {wordDetails.usageNotes.register}
            </p>
          </li>
          <li>
            <h3 className="text-lg font-semibold text-neutral-900">
              Common mistakes
            </h3>
            <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
              {wordDetails.usageNotes.commonMistake}
            </p>
          </li>
        </ul>
      </section>
    </div>
  );
}
