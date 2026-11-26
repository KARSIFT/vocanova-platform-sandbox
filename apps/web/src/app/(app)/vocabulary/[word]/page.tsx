import Link from "@/ui/link";
import { discoveryReturnHref } from "../../discovery/discovery-detail-route";
import { notFound } from "next/navigation";
import {
  listReturnHref,
  vocabularySearchReturnHref,
} from "@/lib/vocabulary-return";

import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { WordDetailContent } from "../../_components/word-detail-content";

export default async function VocabularyWordPage({
  params,
  searchParams,
}: {
  params: Promise<{ word: string }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const [{ word }, query] = await Promise.all([params, searchParams]);
  const finished = query.finished === "1";
  const back = discoveryReturnHref(query.fromDiscovery, finished);
  // Discovery keeps precedence; otherwise restore a validated search state.
  const searchBack = back ? null : vocabularySearchReturnHref(query.fromSearch);
  // Then a validated personal list (with its filter) opened from that list.
  const listBack = back || searchBack ? null : listReturnHref(query.fromList);
  const returnQuery = back
    ? `?fromDiscovery=${query.fromDiscovery}${finished ? "&finished=1" : ""}`
    : searchBack
      ? `?${new URLSearchParams({ fromSearch: searchBack })}`
      : listBack
        ? `?${new URLSearchParams({ fromList: listBack })}`
        : "";
  const client = await createServerApiClient();
  let wordResponse: Awaited<ReturnType<typeof client.getCanonicalWord>>;
  let userResponse: Awaited<ReturnType<typeof client.getCurrentUser>>;
  try {
    [wordResponse, userResponse] = await Promise.all([
      client.getCanonicalWord(word),
      client.getCurrentUser(),
    ]);
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 404) {
      notFound();
    }
    requireAuthRedirect(
      error,
      `/vocabulary/${encodeURIComponent(word)}${returnQuery}`,
    );
  }

  return (
    <PageContainer>
      <Link
        href={back ?? searchBack ?? listBack ?? "/vocabulary"}
        className="page-back"
      >
        <svg aria-hidden="true" viewBox="0 0 24 24">
          <path
            d="m15 5-7 7 7 7"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
        {back
          ? "Back to Discovery"
          : listBack
            ? "Back to list"
            : "Back to word search"}
      </Link>
      <WordDetailContent
        word={wordResponse.data.word}
        userId={userResponse.data.id}
        source="search"
      />
    </PageContainer>
  );
}
