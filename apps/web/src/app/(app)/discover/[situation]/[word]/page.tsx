import Link from "next/link";
import { notFound } from "next/navigation";

import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { SentenceFeedback } from "../../../_components/sentence-feedback";

import { MeaningSaveButton } from "./_components/meaning-save-button";
import { isWordInSituation } from "./_components/word-route";

interface WordDetailPageProps {
  params: Promise<{ situation: string; word: string }>;
}

export default async function WordDetailPage({ params }: WordDetailPageProps) {
  const { situation, word } = await params;
  const client = await createServerApiClient();
  let situationResponse: Awaited<ReturnType<typeof client.getJourneySituation>>;
  let response: Awaited<ReturnType<typeof client.getCanonicalWord>>;
  try {
    situationResponse = await client.getJourneySituation(situation);
    if (!isWordInSituation(situationResponse.data.meanings, word)) {
      notFound();
    }
    response = await client.getCanonicalWord(word);
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 404) {
      notFound();
    }
    requireAuthRedirect(error, `/discover/${situation}/${word}`);
  }

  const { word: wordData } = response.data;

  return (
    <div className="p-[var(--spacing-lg)]">
      <Link
        href={`/discover/${situation}`}
        className="text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
      >
        Back to Journey
      </Link>

      <div className="mt-[var(--spacing-md)]">
        <h1 className="text-2xl font-semibold text-neutral-900">
          {wordData.text}
        </h1>
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          {wordData.wordType}
          {wordData.difficultyLevel ? ` · ${wordData.difficultyLevel}` : null}
        </p>
      </div>

      <section className="mt-[var(--spacing-lg)]">
        <h2 className="text-xl font-semibold text-neutral-900">Meanings</h2>
        <ul className="mt-[var(--spacing-sm)] space-y-[var(--spacing-md)]">
          {wordData.meanings.map((meaning) => (
            <li
              key={meaning.id}
              className="rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
            >
              <div className="flex flex-wrap items-start justify-between gap-[var(--spacing-md)]">
                <div className="min-w-0 flex-1">
                  <p className="font-medium text-neutral-900">
                    {meaning.partOfSpeech}
                  </p>
                  <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
                    {meaning.shortDefinition}
                  </p>
                  {meaning.learnerDefinition ? (
                    <p className="mt-[var(--spacing-xs)] text-base text-neutral-600">
                      {meaning.learnerDefinition}
                    </p>
                  ) : null}
                </div>
                <MeaningSaveButton
                  meaningId={meaning.id}
                  source="journey"
                  initialSaved={meaning.saved}
                  wordText={wordData.text}
                  shortDefinition={meaning.shortDefinition}
                />
              </div>

              {meaning.examples.length > 0 ? (
                <div className="mt-[var(--spacing-md)]">
                  <h3 className="text-base font-semibold text-neutral-900">
                    Example sentences
                  </h3>
                  <ul className="mt-[var(--spacing-xs)] list-disc space-y-[var(--spacing-xs)] pl-[var(--spacing-lg)] text-base text-neutral-700">
                    {meaning.examples.map((example) => (
                      <li key={example.id}>{example.exampleText}</li>
                    ))}
                  </ul>
                </div>
              ) : null}

              {meaning.usageNotes.length > 0 ? (
                <div className="mt-[var(--spacing-md)]">
                  <h3 className="text-base font-semibold text-neutral-900">
                    Usage notes
                  </h3>
                  <ul className="mt-[var(--spacing-xs)] space-y-[var(--spacing-sm)]">
                    {meaning.usageNotes.map((note) => (
                      <li key={note.id}>
                        <h4 className="text-sm font-semibold text-neutral-800">
                          {formatNoteType(note.noteType)}
                        </h4>
                        <p className="text-base text-neutral-700">
                          {note.noteText}
                        </p>
                      </li>
                    ))}
                  </ul>
                </div>
              ) : null}

              {meaning.saved && meaning.userWordId ? (
                <SentenceFeedback
                  targetWord={wordData.text}
                  attemptId={meaning.userWordId}
                  source="word_detail"
                  shortDefinition={meaning.shortDefinition}
                />
              ) : null}
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}

function formatNoteType(noteType: string): string {
  return noteType
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}
