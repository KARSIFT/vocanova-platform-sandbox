import type { WordDetail } from "@vocanova/api-client";

import {
  getAdditionalDefinition,
  getPrimaryDefinition,
} from "@/lib/word-definition";
import { MeaningPicture } from "@/ui/meaning-picture";
import { LearningIntro } from "@/ui/learning-guide";
import { ListenButton } from "@/ui/pronunciation";
import { WordPracticeDisclosure } from "./word-practice-disclosure";

import { wordDifficultyLabel } from "./word-difficulty-label";
import { MeaningSaveButton } from "../discover/[situation]/[word]/_components/meaning-save-button";
import { formatWordReviewState } from "../discover/[situation]/[word]/_components/word-review-state";
import { SentenceFeedback } from "./sentence-feedback";
import { MeaningKnowledgeEditor } from "./meaning-knowledge-editor";
import { MeaningListEditor } from "./meaning-list-editor";
import { MeaningTeaching } from "./meaning-teaching";
import { MeaningComparison, meaningAnchor } from "./meaning-comparison";

interface WordDetailContentProps {
  word: WordDetail;
  userId?: string;
  contextTitle?: string;
  source: "journey" | "search";
}

export function WordDetailContent({
  word: wordData,
  userId,
  contextTitle,
  source,
}: WordDetailContentProps) {
  const difficultyLabel = wordDifficultyLabel(wordData.difficultyLevel);
  return (
    <>
      <header className="mt-[var(--spacing-md)] min-w-0">
        {contextTitle ? (
          <p className="page-eyebrow mb-2">{contextTitle}</p>
        ) : null}
        {/* Set like a dictionary entry: headword, then its kind of word. */}
        <h1 className="headword text-[clamp(2.6rem,8vw,3.6rem)] [overflow-wrap:anywhere]">
          {wordData.text}
        </h1>
        {wordData.meanings.length === 0 && (
          <ListenButton text={wordData.text} />
        )}
        <p className="mt-[var(--spacing-sm)] flex flex-wrap items-center gap-x-3 gap-y-1 text-base">
          <span className="pos">{wordData.wordType.replaceAll("_", " ")}</span>
          {difficultyLabel ? (
            <span className="rounded-full bg-secondary-50 px-2.5 py-0.5 text-sm font-semibold text-secondary-800">
              {difficultyLabel}
            </span>
          ) : null}
        </p>
      </header>

      <MeaningComparison word={wordData} />

      <section
        aria-labelledby="word-meanings-heading"
        className="mt-[var(--spacing-lg)]"
      >
        <h2 id="word-meanings-heading" className="section-title">
          Meanings
        </h2>
        <ul className="mt-[var(--spacing-sm)] space-y-[var(--spacing-md)]">
          {wordData.meanings.map((meaning, index) => {
            const additionalDefinition = getAdditionalDefinition(
              getPrimaryDefinition(
                meaning.shortDefinition,
                meaning.learnerDefinition,
              ),
              meaning.shortDefinition,
            );
            const reviewState = formatWordReviewState(
              meaning.reviewState,
              meaning.due,
            );
            const writingAvailable = Boolean(
              meaning.saved && meaning.userWordId,
            );

            return (
              <li
                key={meaning.id}
                id={meaningAnchor(meaning.id)}
                tabIndex={-1}
                className="learning-surface scroll-mt-24 p-[var(--spacing-md)] sm:p-[var(--spacing-lg)]"
              >
                {/* Kind of word and the save toggle share one row so the
                    definition below keeps the full reading width. */}
                <div
                  id={
                    writingAvailable
                      ? undefined
                      : `sentence-practice-${meaning.id}`
                  }
                  className={`flex flex-wrap items-center justify-between gap-x-[var(--spacing-md)] gap-y-2${writingAvailable ? "" : " scroll-mt-24"}`}
                >
                  <p className="pos text-base">{meaning.partOfSpeech}</p>
                  <MeaningSaveButton
                    meaningId={meaning.id}
                    source={source}
                    initialSaved={meaning.saved}
                    wordText={wordData.text}
                    shortDefinition={meaning.shortDefinition}
                  />
                </div>
                <p className="mt-[var(--spacing-xs)] text-xl leading-snug text-neutral-900">
                  {getPrimaryDefinition(
                    meaning.shortDefinition,
                    meaning.learnerDefinition,
                  )}
                </p>
                {additionalDefinition ? (
                  <details className="mt-[var(--spacing-xs)] text-base text-neutral-600">
                    <summary className="min-h-11 cursor-pointer py-2 font-medium text-primary-700">
                      Reference definition
                    </summary>
                    <p>{additionalDefinition}</p>
                  </details>
                ) : null}
                {reviewState ? (
                  <p className="mt-[var(--spacing-xs)] text-sm font-medium text-primary-700">
                    {reviewState}
                  </p>
                ) : null}
                <ListenButton text={wordData.text} meaningId={meaning.id} />

                <MeaningPicture
                  meaningId={meaning.id}
                  image={meaning.image}
                  meaningRevision={meaning.meaningRevision}
                />
                <MeaningTeaching meaning={meaning} />

                {meaning.saved && meaning.userWordId ? (
                  <div className="mt-6">
                    <LearningIntro
                      title="Try this meaning in a sentence"
                      headingLevel={3}
                      description="Make your own example. Writing is optional, and you can edit your sentence after feedback."
                    />
                    <WordPracticeDisclosure
                      anchor={
                        index === 0
                          ? "sentence-practice"
                          : `sentence-practice-${meaning.id}`
                      }
                      alias={`sentence-practice-${meaning.id}`}
                    >
                      <SentenceFeedback
                        targetWord={wordData.text}
                        attemptId={meaning.userWordId}
                        source="word_detail"
                        userId={userId}
                        shortDefinition={meaning.shortDefinition}
                      />
                    </WordPracticeDisclosure>
                  </div>
                ) : null}

                <details className="mt-4 border-t border-neutral-200 pt-2">
                  <summary className="vn-summary">Personal tools</summary>
                  <MeaningListEditor meaningId={meaning.id} />
                  <MeaningKnowledgeEditor
                    meaningId={meaning.id}
                    initialKnown={meaning.selfReportedKnown}
                  />
                </details>
              </li>
            );
          })}
        </ul>
      </section>
    </>
  );
}
