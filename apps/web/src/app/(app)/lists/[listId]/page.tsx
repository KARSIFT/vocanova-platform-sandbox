import { ApiResponseError, type WordListDetail } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { ListDetail } from "../_components/list-detail";

export default async function PersonalListPage({
  params,
}: {
  params: Promise<{ listId: string }>;
}) {
  const { listId } = await params;
  const client = await createServerApiClient();
  let list: WordListDetail | null = null;
  let missing = false;
  try {
    list = (await client.getWordList(listId)).data;
  } catch (cause) {
    if (cause instanceof ApiResponseError && cause.status === 401)
      requireAuthRedirect(cause, `/lists/${encodeURIComponent(listId)}`);
    missing =
      cause instanceof ApiResponseError &&
      (cause.status === 404 || cause.status === 400 || cause.status === 422);
  }
  return (
    <PageContainer>
      <ListDetail
        key={listId}
        listId={listId}
        initialList={list}
        initialMissing={missing}
      />
    </PageContainer>
  );
}
