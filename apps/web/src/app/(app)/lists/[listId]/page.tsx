import { ApiResponseError, type WordListDetail } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { listPageHref, listSearchQuery } from "@/lib/vocabulary-return";
import { PageContainer } from "@/ui/surface";
import { ListDetail } from "../_components/list-detail";

export default async function PersonalListPage({
  params,
  searchParams,
}: {
  params: Promise<{ listId: string }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const [{ listId }, search] = await Promise.all([params, searchParams]);
  // A bounded member filter restored when returning from a meaning.
  const initialQuery = listSearchQuery(search.q);
  const client = await createServerApiClient();
  let list: WordListDetail | null = null;
  let missing = false;
  try {
    list = (await client.getWordList(listId)).data;
  } catch (cause) {
    if (cause instanceof ApiResponseError && cause.status === 401)
      requireAuthRedirect(
        cause,
        listPageHref(listId, initialQuery) ??
          `/lists/${encodeURIComponent(listId)}`,
      );
    missing =
      cause instanceof ApiResponseError &&
      (cause.status === 404 || cause.status === 400 || cause.status === 422);
  }
  return (
    <PageContainer>
      <ListDetail
        key={`${listId}:${initialQuery}`}
        listId={listId}
        initialList={list}
        initialMissing={missing}
        initialQuery={initialQuery}
      />
    </PageContainer>
  );
}
