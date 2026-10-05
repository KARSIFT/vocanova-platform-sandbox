import Link from "next/link";
import { notFound } from "next/navigation";

import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer } from "@/ui/surface";

import { getSituationDetailView } from "./_components/situation-view";
import { getConversationPractice } from "./_components/conversation-context-content";
import { ConversationContextPractice } from "./_components/conversation-context-practice";
import { formatLevelBand } from "../_components/level-band";
import { UnitGuidebook } from "./_components/unit-guidebook";
import { getSituationOutcome } from "../../_components/situation-presentation";
import { unitGuides } from "./_components/unit-guide-content";

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
  // Secondary catalogs and examples cannot hide the existing situation words.
  // Authentication errors still redirect instead of resembling missing content.
  async function optional<T>(request: Promise<T>): Promise<T | null> {
    try {
      return await request;
    } catch (error) {
      if (error instanceof ApiResponseError && error.status === 401) {
        requireAuthRedirect(error, `/discover/${situation}`);
      }
      return null;
    }
  }
  const guideWords = meanings.filter((meaning) =>
    unitGuides[situation]?.phrases.some(
      (phrase) => phrase.word === meaning.wordText,
    ),
  );
  const [lessonResponse, storyResponse, wordResponses] = await Promise.all([
    optional(client.listLessons()),
    optional(client.listStories()),
    Promise.all(
      guideWords.map((word) =>
        optional(client.getCanonicalWord(word.wordSlug)),
      ),
    ),
  ]);
  const examples: Record<string, string> = {};
  for (let index = 0; index < guideWords.length; index++) {
    const guideWord = guideWords[index];
    if (!guideWord) continue;
    const matchingMeaning = wordResponses[index]?.data.word.meanings.find(
      (meaning) => meaning.id === guideWord.meaningId,
    );
    const example = matchingMeaning?.examples[0]?.exampleText;
    if (example) examples[guideWord.meaningId] = example;
  }
  const conversationPractice = getConversationPractice(
    situationData.id,
    meanings,
  );
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
        {situationData.levelBand
          ? ` · ${formatLevelBand(situationData.levelBand)}`
          : ""}
      </Eyebrow>
      <h1 className="mt-[var(--spacing-xs)] text-3xl font-bold tracking-tight text-neutral-900">
        {situationData.title}
      </h1>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        {getSituationOutcome(situation) ?? situationData.shortDescription}
      </p>

      <nav
        aria-label="This situation"
        className="mt-4 flex flex-wrap gap-x-5 border-y border-neutral-200 py-2"
      >
        {meanings.length > 0 && (
          <a
            href="#situation-words"
            className="inline-flex min-h-11 items-center font-semibold text-primary-700"
          >
            Words to use
          </a>
        )}
        {unitGuides[situation] && (
          <a
            href="#unit-guide"
            className="inline-flex min-h-11 items-center font-semibold text-primary-700"
          >
            A quick guide
          </a>
        )}
        <Link
          href={`/writing?situation=${encodeURIComponent(situation)}`}
          className="inline-flex min-h-11 items-center font-semibold text-primary-700"
        >
          Write about this situation
        </Link>
      </nav>
      {lessonResponse?.data.items.some(
        (lesson) => lesson.situationSlug === situation,
      ) && (
        <section aria-labelledby="situation-lessons-heading" className="mt-6">
          <h2
            id="situation-lessons-heading"
            className="text-xl font-bold text-neutral-900"
          >
            Lessons in this situation
          </h2>
          <ul className="mt-3 divide-y divide-neutral-200">
            {lessonResponse.data.items
              .filter((lesson) => lesson.situationSlug === situation)
              .map((lesson) => (
                <li
                  key={lesson.key}
                  className="flex flex-wrap items-center justify-between gap-2 py-3"
                >
                  <div className="min-w-0">
                    <h3 className="font-semibold text-neutral-900">
                      {lesson.title}
                    </h3>
                    <p className="mt-1 text-sm text-neutral-600">
                      {lesson.status === "completed"
                        ? "Completed"
                        : lesson.status === "in_progress"
                          ? `${lesson.completedSteps} of ${lesson.stepCount} steps`
                          : `${lesson.wordCount} words`}
                    </p>
                  </div>
                  <Link
                    href={`/learn/${encodeURIComponent(lesson.key)}`}
                    className="inline-flex min-h-11 items-center font-semibold text-primary-700"
                  >
                    {lesson.status === "in_progress"
                      ? "Continue lesson"
                      : lesson.status === "completed"
                        ? "Revisit words"
                        : "Start lesson"}
                    <span className="sr-only">: {lesson.title}</span>
                  </Link>
                </li>
              ))}
          </ul>
        </section>
      )}

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

      <UnitGuidebook
        situationSlug={situation}
        meanings={meanings}
        examples={examples}
        lessons={lessonResponse?.data.items ?? null}
        stories={storyResponse?.data.items ?? null}
      />

      {conversationPractice.length > 0 ? (
        <ConversationContextPractice
          key={situationData.id}
          cases={conversationPractice}
          situationSlug={situation}
        />
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
        <ul
          id="situation-words"
          aria-label="Words in this situation"
          tabIndex={-1}
          className="scroll-mt-24 mt-[var(--spacing-lg)] divide-y divide-neutral-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          {meanings.map((meaning) => (
            <li key={meaning.meaningId}>
              <Link
                href={`/discover/${situation}/${meaning.wordSlug}`}
                className="block min-h-11 rounded-lg py-4 px-2 transition-colors hover:bg-primary-50"
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
