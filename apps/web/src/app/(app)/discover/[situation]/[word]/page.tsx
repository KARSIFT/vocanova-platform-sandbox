import Link from "@/ui/link";
import { notFound } from "next/navigation";

import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { WordDetailContent } from "../../../_components/word-detail-content";

import { isWordInSituation } from "./_components/word-route";

interface WordDetailPageProps {
  params: Promise<{ situation: string; word: string }>;
}

export default async function WordDetailPage({ params }: WordDetailPageProps) {
  const { situation, word } = await params;
  const client = await createServerApiClient();
  let situationResponse: Awaited<ReturnType<typeof client.getJourneySituation>>;
  let response: Awaited<ReturnType<typeof client.getCanonicalWord>>;
  let currentUserResponse: Awaited<ReturnType<typeof client.getCurrentUser>>;
  // The word read starts with the situation read instead of after it. Results
  // are still checked in the original order: situation and membership first.
  const wordRead = Promise.allSettled([client.getCanonicalWord(word)]).then(
    ([result]) => result,
  );
  try {
    [situationResponse, currentUserResponse] = await Promise.all([
      client.getJourneySituation(situation),
      client.getCurrentUser(),
    ]);
    if (!isWordInSituation(situationResponse.data.meanings, word)) {
      notFound();
    }
    const wordResult = await wordRead;
    if (wordResult.status === "rejected") throw wordResult.reason;
    response = wordResult.value;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 404) {
      notFound();
    }
    requireAuthRedirect(error, `/discover/${situation}/${word}`);
  }

  const { word: wordData } = response.data;

  return (
    <PageContainer>
      <Link
        href={`/discover/${situation}`}
        className="inline-flex min-h-11 items-center text-base font-semibold text-primary-700 hover:text-primary-800"
      >
        Back to {situationResponse.data.situation.title}
      </Link>

      <WordDetailContent
        word={wordData}
        userId={currentUserResponse.data.id}
        source="journey"
      />
    </PageContainer>
  );
}
