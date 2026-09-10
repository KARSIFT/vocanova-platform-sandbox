import Link from "next/link";
import { notFound } from "next/navigation";

import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, Surface } from "@/ui/surface";
import { SentenceFeedback } from "../../_components/sentence-feedback";
import { RemoveSavedWordButton } from "../_components/remove-saved-word-button";
import { formatSavedWordStatus } from "../_components/saved-word-view";

interface SavedWordDetailPageProps {
  params: Promise<{ userWordId: string }>;
}

export default async function SavedWordDetailPage({
  params,
}: SavedWordDetailPageProps) {
  const { userWordId } = await params;
  const client = await createServerApiClient();

  let savedResponse: Awaited<ReturnType<typeof client.getSavedWord>>;
  try {
    savedResponse = await client.getSavedWord(userWordId);
  } catch (error) {
    if (
      error instanceof ApiResponseError &&
      (error.status === 404 || error.status === 422)
    ) {
      notFound();
    }
    requireAuthRedirect(error, `/words/${userWordId}`);
  }

  const savedWord = savedResponse.data;
  let wordResponse: Awaited<ReturnType<typeof client.getCanonicalWord>>;
  try {
    wordResponse = await client.getCanonicalWord(savedWord.wordSlug);
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 404) {
      notFound();
    }
    requireAuthRedirect(error, `/words/${userWordId}`);
  }

  const meaning = wordResponse.data.word.meanings.find(
    (candidate) => candidate.id === savedWord.meaningId,
  );
  if (!meaning) {
    notFound();
  }
  const statusLabel = formatSavedWordStatus(savedWord.status);

  return (
    <PageContainer>
      <Link
        href="/words"
        className="inline-flex min-h-11 items-center text-base font-semibold text-primary-700 hover:text-primary-800"
      >
        Back to saved vocabulary
      </Link>

      <div className="mt-[var(--spacing-md)] flex flex-wrap items-start justify-between gap-[var(--spacing-md)]">
        <div>
          <h1 className="text-2xl font-semibold text-neutral-900">
            {savedWord.wordText}
          </h1>
          <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
            {savedWord.partOfSpeech}
          </p>
          {statusLabel ? (
            <p className="mt-[var(--spacing-xs)] text-sm font-semibold text-primary-700">
              {statusLabel}
            </p>
          ) : null}
        </div>
        <RemoveSavedWordButton
          meaningId={savedWord.meaningId}
          wordText={savedWord.wordText}
          redirectTo="/words"
        />
      </div>

      <Surface className="mt-[var(--spacing-lg)]">
        <h2 className="text-xl font-semibold text-neutral-900">Meaning</h2>
        <p className="mt-[var(--spacing-sm)] text-base text-neutral-800">
          {meaning.shortDefinition}
        </p>
        {meaning.learnerDefinition ? (
          <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
            {meaning.learnerDefinition}
          </p>
        ) : null}

        {meaning.examples.length > 0 ? (
          <div className="mt-[var(--spacing-md)]">
            <h3 className="text-lg font-semibold text-neutral-900">
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
            <h3 className="text-lg font-semibold text-neutral-900">
              Usage notes
            </h3>
            <ul className="mt-[var(--spacing-xs)] space-y-[var(--spacing-sm)]">
              {meaning.usageNotes.map((note) => (
                <li key={note.id}>
                  <h4 className="text-sm font-semibold text-neutral-800">
                    {formatNoteType(note.noteType)}
                  </h4>
                  <p className="text-base text-neutral-700">{note.noteText}</p>
                </li>
              ))}
            </ul>
          </div>
        ) : null}
      </Surface>

      <SentenceFeedback
        targetWord={savedWord.wordText}
        attemptId={savedWord.userWordId}
        source="word_detail"
        shortDefinition={savedWord.shortDefinition}
      />
    </PageContainer>
  );
}

function formatNoteType(noteType: string): string {
  return noteType
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}
