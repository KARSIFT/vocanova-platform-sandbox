import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";

import { PracticePlayer } from "../../_components/practice-player";

export const metadata: Metadata = {
  title: "Practice session — Vocanova",
  description:
    "Remember and use your vocabulary in a focused practice session.",
};

export default async function PracticeSessionPage({
  params,
}: {
  params: Promise<{ sessionId: string }>;
}) {
  const { sessionId } = await params;
  try {
    const { data } = await (
      await createServerApiClient()
    ).getPracticeSession(sessionId);
    return (
      <PageContainer className="max-w-[44rem]">
        <PracticePlayer key={data.id} initialSession={data} />
      </PageContainer>
    );
  } catch (error) {
    if (
      error instanceof ApiResponseError &&
      (error.status === 404 || error.status === 422)
    )
      notFound();
    requireAuthRedirect(
      error,
      `/practice/session/${encodeURIComponent(sessionId)}`,
    );
  }
}
