import Link from "next/link";
import { notFound } from "next/navigation";

import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { getAdditionalDefinition } from "@/lib/word-definition";
import { PageContainer, Surface } from "@/ui/surface";
import { MeaningPicture } from "@/ui/meaning-picture";
import { WordPracticeDisclosure } from "../../_components/word-practice-disclosure";
import { ListenButton } from "@/ui/pronunciation";
import { SentenceFeedback } from "../../_components/sentence-feedback";
import { MeaningKnowledgeEditor } from "../../_components/meaning-knowledge-editor";
import { MeaningListEditor } from "../../_components/meaning-list-editor";
import { MeaningTeaching } from "../../_components/meaning-teaching";
import { MeaningComparison } from "../../_components/meaning-comparison";
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
  let currentUserResponse: Awaited<ReturnType<typeof client.getCurrentUser>>;
  try {
    [savedResponse, currentUserResponse] = await Promise.all([
      client.getSavedWord(userWordId),
      client.getCurrentUser(),
    ]);
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
  const additionalDefinition = getAdditionalDefinition(
    meaning.shortDefinition,
    meaning.learnerDefinition,
  );

  return (
    <PageContainer>
      <Link
        href="/words"
        className="inline-flex min-h-11 items-center text-base font-semibold text-primary-700 hover:text-primary-800"
      >
        Saved words
      </Link>

      <div className="mt-[var(--spacing-md)] flex flex-wrap items-start justify-between gap-[var(--spacing-md)]">
        <div>
          <h1 className="text-2xl font-semibold text-neutral-900">
            {savedWord.wordText}
          </h1>
          <ListenButton text={savedWord.wordText} />
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
        {additionalDefinition ? (
          <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
            {additionalDefinition}
          </p>
        ) : null}

        <MeaningPicture meaningId={meaning.id} />
        <MeaningTeaching meaning={meaning} />
        <details className="mt-4 border-t border-neutral-200 pt-2">
          <summary className="min-h-12 cursor-pointer content-center font-semibold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
            Personal tools
          </summary>
          <MeaningListEditor meaningId={meaning.id} />
          <MeaningKnowledgeEditor
            meaningId={meaning.id}
            initialKnown={meaning.selfReportedKnown}
          />
        </details>
      </Surface>

      <MeaningComparison
        word={wordResponse.data.word}
        currentMeaningId={meaning.id}
        canonicalPath={`/vocabulary/${encodeURIComponent(savedWord.wordSlug)}`}
      />

      <WordPracticeDisclosure>
        <SentenceFeedback
          targetWord={savedWord.wordText}
          attemptId={savedWord.userWordId}
          source="word_detail"
          userId={currentUserResponse.data.id}
          shortDefinition={savedWord.shortDefinition}
        />
        <Link
          href="/writing"
          className="mt-4 inline-flex min-h-12 items-center rounded-md font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Choose a writing topic
        </Link>
      </WordPracticeDisclosure>
    </PageContainer>
  );
}
