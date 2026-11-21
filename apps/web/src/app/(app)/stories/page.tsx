import { ActionLink } from "@/ui/controls";
import type { Metadata } from "next";
import Link from "@/ui/link";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, PageHeader } from "@/ui/surface";
import {
  primaryAction,
  secondaryAction,
  textLink,
} from "../practice/_components/practice-styles";

import styles from "./stories.module.css";

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
  const nextSessionId = library.items.find(
    (story) => story.latestSession?.status === "in_progress",
  )?.latestSession?.id;
  return (
    <PageContainer width="standard">
      <PageHeader
        title="Everyday stories"
        back={{ href: "/discover", label: "Back to Journey" }}
        lede="Follow a short dialogue, explore useful words, and check what you understand. Audio is optional."
      />
      {!library.items.length ? (
        <div className="vn-empty">
          <h2>No stories are available right now.</h2>
          <Link href="/courses" className={textLink}>
            Explore situations
          </Link>
        </div>
      ) : (
        <ul className="vn-list" role="list">
          {library.items.map((story) => (
            <li key={story.key}>
              <section
                className={styles.story}
                aria-labelledby={`story-${story.key}`}
              >
                <div className="min-w-0">
                  <p className="text-sm font-semibold text-neutral-600">
                    {story.situation} · {story.level}
                  </p>
                  <h2 id={`story-${story.key}`} className="section-title mt-1">
                    {story.title}
                  </h2>
                  <p className="mt-1 text-neutral-700">{story.description}</p>
                  <p className="mt-2 text-sm text-neutral-600">
                    {story.lineCount} lines · {story.questionCount} checks ·{" "}
                    {story.vocabulary.map((word) => word.wordText).join(", ")}
                  </p>
                  {story.latestSession && (
                    <p className="mt-2 text-sm font-semibold text-primary-700">
                      {story.latestSession.status === "completed"
                        ? "Story complete. Read or practise again."
                        : `${story.latestSession.completedSteps} of ${story.latestSession.totalSteps} steps saved.`}
                    </p>
                  )}
                </div>
                <div className={styles.storyActions}>
                  {story.latestSession?.status === "in_progress" ? (
                    <>
                      <Link
                        href={`/stories/session/${encodeURIComponent(story.latestSession.id)}`}
                        className={
                          story.latestSession.id === nextSessionId
                            ? primaryAction
                            : secondaryAction
                        }
                      >
                        Continue story
                      </Link>
                      <Link
                        href={`/stories/${encodeURIComponent(story.key)}`}
                        className={textLink}
                      >
                        Story details
                      </Link>
                    </>
                  ) : (
                    <ActionLink
                      variant="secondary"
                      href={`/stories/${encodeURIComponent(story.key)}`}
                      className={secondaryAction}
                    >
                      {story.latestSession?.status === "completed"
                        ? "Read and practise again"
                        : "Read and practise"}
                    </ActionLink>
                  )}
                </div>
              </section>
            </li>
          ))}
        </ul>
      )}
      <p className="mt-6 text-sm text-neutral-600">
        Story checks record your answers. They do not award points or change
        scheduled word reviews.
      </p>
    </PageContainer>
  );
}
