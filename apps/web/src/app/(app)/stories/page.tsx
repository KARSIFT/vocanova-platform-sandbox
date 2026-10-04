import type { Metadata } from "next";
import Link from "next/link";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer, Surface } from "@/ui/surface";
import { textLink } from "../practice/_components/practice-styles";

export const metadata: Metadata = {
  title: "Stories — Vocanova",
  description: "Original short English dialogues for everyday situations.",
};
export default async function StoriesPage() {
  let library;
  try {
    library = (await (await createServerApiClient()).listStories()).data;
  } catch (error) {
    requireAuthRedirect(error, "/stories");
  }
  return (
    <PageContainer className="max-w-[64rem]">
      <Link href="/discover" className={textLink}>
        Back to Journey
      </Link>
      <header className="mb-6 mt-3">
        <Eyebrow>Read, notice, practise</Eyebrow>
        <h1 className="mt-2 text-3xl font-bold tracking-tight text-neutral-900">
          Everyday stories
        </h1>
        <p className="mt-3 max-w-[40rem] text-lg text-neutral-700">
          Follow an original short dialogue, explore useful words, and check
          what you understand. Audio is optional.
        </p>
        <p className="mt-2 text-sm text-neutral-600">
          A2–B1 · Saved progress · Repeat whenever you like
        </p>
      </header>
      {!library.items.length ? (
        <Surface>
          <p>No stories are available right now.</p>
          <Link href="/discover" className={textLink}>
            Explore situations
          </Link>
        </Surface>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {library.items.map((story) => (
            <Surface key={story.key} aria-labelledby={`story-${story.key}`}>
              <Eyebrow>
                {story.situation} · {story.level}
              </Eyebrow>
              <h2
                id={`story-${story.key}`}
                className="mt-2 text-xl font-bold text-neutral-900"
              >
                {story.title}
              </h2>
              <p className="mt-2 text-neutral-700">{story.description}</p>
              <p className="mt-2 text-sm text-neutral-600">
                {story.lineCount} lines · {story.questionCount} checks ·{" "}
                {story.vocabulary.map((word) => word.wordText).join(", ")}
              </p>
              {story.latestSession && (
                <p className="mt-3 font-semibold text-neutral-900">
                  {story.latestSession.status === "completed"
                    ? "Story complete. Read or practise again."
                    : `${story.latestSession.completedSteps} of ${story.latestSession.totalSteps} steps saved.`}
                </p>
              )}
              {story.latestSession?.status === "in_progress" && (
                <Link
                  href={`/stories/session/${encodeURIComponent(story.latestSession.id)}`}
                  className={`${textLink} mt-3 mr-4`}
                >
                  Continue story
                </Link>
              )}
              <Link
                href={`/stories/${encodeURIComponent(story.key)}`}
                className={`${textLink} mt-3`}
              >
                Read and practise
              </Link>
            </Surface>
          ))}
        </div>
      )}
      <p className="mt-6 text-sm text-neutral-600">
        Story checks record your answers. They do not award points or change
        scheduled word reviews.
      </p>
    </PageContainer>
  );
}
