import { ActionLink } from "@/ui/controls";

import type { LearnerSentence } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, PageHeader } from "@/ui/surface";

const PAGE_SIZE = 12;

interface SentenceHistoryPageProps {
  searchParams: Promise<{ after?: string }>;
}

export default async function SentenceHistoryPage({
  searchParams,
}: SentenceHistoryPageProps) {
  const { after } = await searchParams;
  const client = await createServerApiClient();
  let response: Awaited<ReturnType<typeof client.listLearnerSentences>>;
  try {
    response = await client.listLearnerSentences(
      {
        limit: PAGE_SIZE,
        ...(after ? { after } : {}),
      },
      { cache: "no-store" },
    );
  } catch (error) {
    requireAuthRedirect(
      error,
      `/progress/sentences${after ? `?after=${encodeURIComponent(after)}` : ""}`,
    );
  }

  const { items, hasMore, nextCursor } = response.data;

  return (
    <PageContainer>
      <PageHeader
        title="Sentence history"
        eyebrow="Your writing practice"
        back={{ href: "/progress", label: "Back to Progress" }}
        lede="Look back at your sentences and the feedback you received."
      />

      {items.length === 0 ? (
        after ? (
          <ExhaustedSentenceHistory />
        ) : (
          <EmptySentenceHistory />
        )
      ) : (
        <SentenceList items={items} />
      )}

      {hasMore && nextCursor ? (
        <ActionLink
          variant="secondary"
          href={`/progress/sentences?after=${encodeURIComponent(nextCursor)}`}
          className="mt-[var(--spacing-lg)]"
        >
          View older sentences
        </ActionLink>
      ) : null}
    </PageContainer>
  );
}

function EmptySentenceHistory() {
  return (
    <section
      className="vn-empty"
      aria-labelledby="empty-sentence-history-heading"
    >
      <h2
        id="empty-sentence-history-heading"
        className="text-xl font-semibold text-neutral-900"
      >
        Your sentence practice will appear here
      </h2>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        Write a sentence from Home or a saved word to keep a useful record of
        your feedback.
      </p>
      <ActionLink
        variant="primary"
        href="/home"
        className="mt-[var(--spacing-md)]"
      >
        Go to Home
      </ActionLink>
    </section>
  );
}

function ExhaustedSentenceHistory() {
  return (
    <section
      className="vn-empty"
      aria-labelledby="exhausted-sentence-history-heading"
    >
      <h2
        id="exhausted-sentence-history-heading"
        className="text-xl font-semibold text-neutral-900"
      >
        You&apos;ve reached the end of this history
      </h2>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        Newer sentence practice is available from the beginning of your history.
      </p>
      <ActionLink
        variant="primary"
        href="/progress/sentences"
        className="mt-[var(--spacing-md)]"
      >
        Back to newest sentences
      </ActionLink>
    </section>
  );
}

function SentenceList({ items }: Readonly<{ items: LearnerSentence[] }>) {
  return (
    <ol className="vn-list">
      {items.map((sentence) => (
        <li key={sentence.id} className="min-w-0 p-4 sm:p-6">
          <article aria-labelledby={`sentence-${sentence.id}`}>
            <div className="flex flex-wrap items-start justify-between gap-[var(--spacing-sm)]">
              <h2
                id={`sentence-${sentence.id}`}
                className="text-lg font-semibold text-neutral-900"
              >
                {formatProcessingState(sentence)}
              </h2>
              <time
                className="text-sm text-neutral-600"
                dateTime={sentence.createdAt}
              >
                {formatDate(sentence.createdAt)}
              </time>
            </div>
            <SentenceText
              label="Your sentence"
              value={sentence.originalSentence}
            />
            {sentence.processingStatus === "completed" ? (
              <CompletedFeedback sentence={sentence} />
            ) : (
              <p className="mt-[var(--spacing-md)] text-sm text-neutral-700">
                {getProcessingDescription(sentence.processingStatus)}
              </p>
            )}
          </article>
        </li>
      ))}
    </ol>
  );
}

function CompletedFeedback({
  sentence,
}: Readonly<{ sentence: LearnerSentence }>) {
  return (
    <div className="mt-3 space-y-3">
      {sentence.correctedSentence &&
      sentence.correctedSentence !== sentence.originalSentence ? (
        <SentenceText
          label="Suggested revision"
          revised
          value={sentence.correctedSentence}
        />
      ) : null}
      {sentence.headline &&
      sentence.headline !== formatProcessingState(sentence) ? (
        <p className="font-medium text-neutral-900">{sentence.headline}</p>
      ) : null}
      {sentence.explanation ? (
        <p className="text-base text-neutral-700">{sentence.explanation}</p>
      ) : null}
      {sentence.improvementTip ? (
        <p className="rounded-md bg-secondary-50 p-[var(--spacing-sm)] text-base text-secondary-900">
          <span className="font-semibold">Tip: </span>
          {sentence.improvementTip}
        </p>
      ) : null}
    </div>
  );
}

function SentenceText({
  label,
  value,
  revised,
}: Readonly<{ label: string; value: string; revised?: boolean }>) {
  return (
    <div
      className={`mt-3 rounded-xl p-3 ${revised ? "bg-primary-50" : "bg-neutral-50"}`}
    >
      <p className="text-sm font-medium text-neutral-700">{label}</p>
      <p className="mt-[var(--spacing-xs)] wrap-anywhere text-base text-neutral-900">
        {value}
      </p>
    </div>
  );
}

function formatProcessingState(sentence: LearnerSentence): string {
  if (sentence.processingStatus !== "completed") {
    return sentence.processingStatus === "pending"
      ? "Feedback is being prepared"
      : sentence.processingStatus === "failed"
        ? "Feedback wasn’t completed"
        : "Feedback was skipped";
  }
  switch (sentence.status) {
    case "correct":
      return "Looks good";
    case "needs_improvement":
      return "A small revision could help";
    case "incorrect":
      return "Let’s refine this sentence";
    default:
      return "Feedback complete";
  }
}

function getProcessingDescription(
  status: LearnerSentence["processingStatus"],
): string {
  switch (status) {
    case "pending":
      return "We’re still checking this sentence. Come back shortly for the full feedback.";
    case "failed":
      return "We couldn’t complete the feedback for this sentence. You can try another sentence from Home or a saved word.";
    case "skipped":
      return "This sentence was saved without automated feedback.";
    default:
      return "";
  }
}

function formatDate(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? "Saved recently"
    : `${new Intl.DateTimeFormat(undefined, {
        dateStyle: "medium",
        timeZone: "UTC",
      }).format(date)} UTC`;
}
