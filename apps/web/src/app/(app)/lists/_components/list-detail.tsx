"use client";
import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ApiResponseError, type WordListDetail } from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { Surface } from "@/ui/surface";
import { listWriteHeaders, useListAction } from "./use-list-action";
import { ListActionStatus } from "./list-action-status";
import { listButton, listInput, listLink, listPrimary } from "./list-styles";

export function ListDetail({
  listId,
  initialList,
  initialMissing,
}: {
  listId: string;
  initialList: WordListDetail | null;
  initialMissing: boolean;
}) {
  const router = useRouter();
  const [list, setList] = useState(initialList);
  const [missing, setMissing] = useState(initialMissing);
  const [name, setName] = useState(initialList?.name ?? "");
  const [query, setQuery] = useState("");
  const [confirmDelete, setConfirmDelete] = useState(false);
  const action = useListAction();
  const reload = () =>
    void action.reload(async () => {
      try {
        const { data } = await createApiClient().getWordList(listId);
        setList(data);
        setName(data.name);
        setMissing(false);
        setConfirmDelete(false);
      } catch (cause) {
        if (cause instanceof ApiResponseError && cause.status === 404) {
          setList(null);
          setMissing(true);
          return;
        }
        throw cause;
      }
    });
  function rename() {
    if (!list) return;
    const body = { name: name.trim(), expectedRevision: list.revision };
    const key = crypto.randomUUID();
    void action.run(async () => {
      const { data } = await createApiClient().putWordList(
        listId,
        body,
        key,
        await listWriteHeaders(),
      );
      setList(data);
      setName(data.name);
      return "List name confirmed.";
    });
  }
  function remove(meaningId: string) {
    if (!list) return;
    const revision = list.revision;
    const key = crypto.randomUUID();
    void action.run(async () => {
      const { data } = await createApiClient().deleteWordListMember(
        listId,
        meaningId,
        revision,
        key,
        await listWriteHeaders(),
      );
      setList(data);
      return "Current list membership confirmed.";
    });
  }
  function deleteList() {
    if (!list) return;
    const revision = list.revision;
    const key = crypto.randomUUID();
    void action.run(async () => {
      await createApiClient().deleteWordList(
        listId,
        revision,
        key,
        await listWriteHeaders(),
      );
      setList(null);
      setMissing(true);
      router.push("/lists");
      return "List deletion confirmed.";
    });
  }
  const members =
    list?.members.filter((member) =>
      `${member.wordText} ${member.shortDefinition}`
        .toLocaleLowerCase()
        .includes(query.trim().toLocaleLowerCase()),
    ) ?? [];
  return (
    <>
      {!list ? (
        <Surface className="mt-4">
          <h1 className="text-2xl font-bold text-neutral-900">
            {missing
              ? "This list is no longer available"
              : "Your list could not be loaded"}
          </h1>
          <p className="mt-2 text-neutral-700">
            {missing
              ? "It may have been deleted or belong to another account. Choose one of your current lists."
              : "Try loading its current status again."}
          </p>
          {missing ? (
            <Link href="/lists" className={`${listLink} mt-3`}>
              Choose a personal list
            </Link>
          ) : (
            <button
              type="button"
              onClick={reload}
              disabled={action.busy}
              className={`${listButton} mt-3`}
            >
              Try loading list again
            </button>
          )}
        </Surface>
      ) : (
        <>
          <header className="my-4">
            <h1 className="break-words text-3xl font-bold tracking-tight text-neutral-900">
              {list.name}
            </h1>
            <p className="mt-2 text-neutral-700">
              {list.memberCount}{" "}
              {list.memberCount === 1 ? "meaning" : "meanings"}.{" "}
              {list.usableMemberCount} available for typed and listening
              practice.
            </p>
            <p className="mt-2 text-sm text-neutral-600">
              List membership does not save words for scheduled review or change
              what you marked as known.
            </p>
            <div className="mt-3 flex flex-wrap gap-3">
              {list.usableMemberCount > 0 && (
                <Link
                  href={`/practice?list=${encodeURIComponent(list.id)}`}
                  className={listPrimary}
                >
                  Practice this list
                </Link>
              )}
              <Link href="/vocabulary" className={listButton}>
                Find words to add
              </Link>
            </div>
            {list.usableMemberCount === 0 && (
              <p className="mt-3 text-neutral-700">
                {list.memberCount
                  ? "These meanings are not yet available in focused practice. Add a supported meaning to start a list session."
                  : "Open a word and use Personal lists to add its meaning."}
              </p>
            )}
          </header>
          <Surface aria-labelledby="rename-list-heading">
            <h2
              id="rename-list-heading"
              className="text-lg font-bold text-neutral-900"
            >
              List settings
            </h2>
            <form
              onSubmit={(event) => {
                event.preventDefault();
                rename();
              }}
            >
              <label
                htmlFor="rename-list-name"
                className="mt-3 block font-semibold text-neutral-900"
              >
                List name
              </label>
              <input
                id="rename-list-name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                disabled={action.locked}
                required
                maxLength={160}
                className={listInput}
              />
              <button
                type="submit"
                disabled={
                  action.locked ||
                  !name.trim() ||
                  Array.from(name.trim()).length > 80 ||
                  name.trim() === list.name
                }
                className={`${listButton} mt-3`}
              >
                Rename list
              </button>
            </form>
            {!confirmDelete ? (
              <button
                type="button"
                disabled={action.locked}
                onClick={() => setConfirmDelete(true)}
                className={`${listButton} mt-3`}
              >
                Delete list
              </button>
            ) : (
              <div className="mt-4 border-t border-neutral-200 pt-4">
                <p className="text-neutral-700">
                  Delete this list and all its memberships? Your saved review
                  words and knowledge notes remain available.
                </p>
                <div className="mt-3 flex flex-wrap gap-3">
                  <button
                    type="button"
                    disabled={action.locked}
                    onClick={deleteList}
                    className={listButton}
                  >
                    Confirm delete list
                  </button>
                  <button
                    type="button"
                    disabled={action.locked}
                    onClick={() => setConfirmDelete(false)}
                    className={listButton}
                  >
                    Keep list
                  </button>
                </div>
              </div>
            )}
          </Surface>
          <div role="search" aria-label="Search list meanings" className="my-5">
            <label
              htmlFor="list-member-search"
              className="font-semibold text-neutral-900"
            >
              Search words and meanings in this list
            </label>
            <input
              id="list-member-search"
              type="search"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              maxLength={100}
              className={listInput}
            />
          </div>
          {members.length ? (
            <ul aria-label="List meanings" className="space-y-3">
              {members.map((member) => (
                <li
                  key={member.meaningId}
                  className="flex min-w-0 flex-wrap items-center justify-between gap-3 rounded-xl border border-neutral-200 bg-white p-4"
                >
                  <div className="min-w-0 flex-1">
                    <h2 className="break-words text-lg font-bold text-neutral-900">
                      <Link
                        href={`/vocabulary/${encodeURIComponent(member.wordSlug)}`}
                        className={listLink}
                      >
                        {member.wordText}
                      </Link>
                    </h2>
                    <p className="text-neutral-700">
                      {member.partOfSpeech} · {member.shortDefinition}
                    </p>
                    <p className="mt-1 text-sm text-neutral-600">
                      {member.practiceAvailable
                        ? "Available for list practice"
                        : "Organized here; focused practice is not yet available for this meaning"}
                    </p>
                  </div>
                  <button
                    type="button"
                    disabled={action.locked}
                    className={listButton}
                    onClick={() => remove(member.meaningId)}
                    aria-label={`Remove ${member.wordText} from this list`}
                  >
                    Remove from list
                  </button>
                </li>
              ))}
            </ul>
          ) : (
            <Surface>
              <h2 className="text-xl font-bold text-neutral-900">
                {query.trim()
                  ? "No meanings match your search"
                  : "Your list is empty"}
              </h2>
              <p className="mt-2 text-neutral-700">
                {query.trim()
                  ? "Try another word or part of its meaning."
                  : "Find a useful word, then explicitly add its meaning to this list."}
              </p>
            </Surface>
          )}
        </>
      )}
      <ListActionStatus action={action} reload={reload} />
    </>
  );
}
