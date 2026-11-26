import { notFound } from "next/navigation";

import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import {
  savedWordHref,
  savedWordsRestartHref,
  savedWordsReturnHref,
} from "@/lib/vocabulary-return";
import {
  getAdditionalDefinition,
  getPrimaryDefinition,
} from "@/lib/word-definition";
import { PageContainer, PageHeader, Surface } from "@/ui/surface";
import { MeaningPicture } from "@/ui/meaning-picture";
import { LearningIntro } from "@/ui/learning-guide";
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
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

export default async function SavedWordDetailPage({
  params,
  searchParams,
}: SavedWordDetailPageProps) {
  const [{ userWordId }, query] = await Promise.all([params, searchParams]);
  // Restore a validated, filtered/paged Saved collection; else the default.
  const savedBack = savedWordsReturnHref(query.fromSaved);
  const authReturn = savedWordHref(userWordId, savedBack);
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
    requireAuthRedirect(error, authReturn);
  }

  const savedWord = savedResponse.data;
  let wordResponse: Awaited<ReturnType<typeof client.getCanonicalWord>>;
  try {
    wordResponse = await client.getCanonicalWord(savedWord.wordSlug);
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 404) {
      notFound();
    }
    requireAuthRedirect(error, authReturn);
  }

  const meaning = wordResponse.data.word.meanings.find(
    (candidate) => candidate.id === savedWord.meaningId,
  );
  if (!meaning) {
    notFound();
  }
  const statusLabel = formatSavedWordStatus(
    savedWord.status,
    meaning.reviewState,
  );
  const additionalDefinition = getAdditionalDefinition(
    getPrimaryDefinition(meaning.shortDefinition, meaning.learnerDefinition),
    meaning.shortDefinition,
  );

  return (
    <PageContainer>
      <PageHeader
        title={<span className="headword">{savedWord.wordText}</span>}
        back={{ href: savedBack ?? "/words", label: "Saved words" }}
        actions={
          <RemoveSavedWordButton
            meaningId={savedWord.meaningId}
            wordText={savedWord.wordText}
            redirectTo={savedWordsRestartHref(savedBack)}
          />
        }
      >
        <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1">
          <span className="pos">{savedWord.partOfSpeech}</span>
          {statusLabel && (
            <span className="text-sm font-semibold text-primary-700">
              {statusLabel}
            </span>
          )}
          <ListenButton
            text={savedWord.wordText}
            meaningId={savedWord.meaningId}
          />
        </div>
      </PageHeader>

      <Surface aria-labelledby="saved-meaning-heading">
        <h2 id="saved-meaning-heading" className="section-title">
          Meaning
        </h2>
        <p className="mt-2 text-lg leading-snug text-neutral-900 [overflow-wrap:anywhere]">
          {getPrimaryDefinition(
            meaning.shortDefinition,
            meaning.learnerDefinition,
          )}
        </p>
        {additionalDefinition ? (
          <details className="mt-[var(--spacing-xs)] text-base text-neutral-600">
            <summary className="vn-summary text-sm">
              Reference definition
            </summary>
            <p>{additionalDefinition}</p>
          </details>
        ) : null}

        <MeaningPicture
          meaningId={meaning.id}
          image={meaning.image}
          meaningRevision={meaning.meaningRevision}
        />
        <MeaningTeaching meaning={meaning} />
      </Surface>
      <section aria-labelledby="saved-word-practice-heading" className="mt-8">
        <LearningIntro
          id="saved-word-practice-heading"
          title="Try this meaning in a sentence"
          headingLevel={2}
          description="Make your own example. Writing is optional, and you can edit your sentence after feedback."
        />
        <WordPracticeDisclosure>
          <SentenceFeedback
            targetWord={savedWord.wordText}
            attemptId={savedWord.userWordId}
            source="word_detail"
            userId={currentUserResponse.data.id}
            shortDefinition={savedWord.shortDefinition}
          />
        </WordPracticeDisclosure>
      </section>

      <details className="mt-6 border-t border-neutral-200">
        <summary className="vn-summary">Personal tools</summary>
        <MeaningListEditor meaningId={meaning.id} />
        <MeaningKnowledgeEditor
          meaningId={meaning.id}
          initialKnown={meaning.selfReportedKnown}
        />
      </details>

      <MeaningComparison
        word={wordResponse.data.word}
        currentMeaningId={meaning.id}
        canonicalPath={`/vocabulary/${encodeURIComponent(savedWord.wordSlug)}`}
      />
    </PageContainer>
  );
}
