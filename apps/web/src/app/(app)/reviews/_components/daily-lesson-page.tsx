import { notFound } from "next/navigation";
import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";

import { DailyLessonPlayer } from "./daily-lesson";

export async function DailyReviewsContent({
  sessionId,
}: { sessionId?: string } = {}) {
  try {
    const client = await createServerApiClient();
    const [{ data }, selected] = await Promise.all([
      client.getCurrentDailyLesson(),
      sessionId ? client.getDailyLesson(sessionId) : Promise.resolve(null),
    ]);
    const session = selected?.data ?? data.session;
    return (
      <PageContainer className="max-w-[44rem]">
        <DailyLessonPlayer
          key={session?.id ?? "new"}
          initialSession={session}
          initialDueCount={data.dueCount}
        />
      </PageContainer>
    );
  } catch (error) {
    if (
      sessionId &&
      error instanceof ApiResponseError &&
      [400, 404, 422].includes(error.status)
    )
      notFound();
    requireAuthRedirect(
      error,
      sessionId
        ? `/review/session?lesson=${encodeURIComponent(sessionId)}`
        : "/review",
    );
  }
}
