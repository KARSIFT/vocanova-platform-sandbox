"use client";

import { ActionLink, Button, TextInput } from "@/ui/controls";
import { useLayoutEffect, useRef, useState } from "react";
import Link from "@/ui/link";
import { useRouter } from "next/navigation";
import { ApiResponseError, type WordListDetail } from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { PageHeader } from "@/ui/surface";
import { getPrimaryDefinition } from "@/lib/word-definition";
import { listPageHref } from "@/lib/vocabulary-return";
import { listWriteHeaders, useListAction } from "./use-list-action";
import { ListActionStatus } from "./list-action-status";
import { listLink } from "./list-styles";
import { meaningAnchor } from "../../_components/meaning-comparison";

export function ListDetail({
  listId,
  initialList,
  initialMissing,
  initialQuery = "",
}: {
  listId: string;
  initialList: WordListDetail | null;
  initialMissing: boolean;
  initialQuery?: string;
}) {
  const router = useRouter();
  const [list, setList] = useState(initialList);
  const [missing, setMissing] = useState(initialMissing);
  const [name, setName] = useState(initialList?.name ?? "");
  const [query, setQuery] = useState(initialQuery);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const action = useListAction();
  const settingsRef = useRef<HTMLDetailsElement>(null);
  const managementRequest = useRef(false);
  useLayoutEffect(() => {
    if (action.error && managementRequest.current && settingsRef.current)
      settingsRef.current.open = true;
  }, [action.error]);
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
    managementRequest.current = true;
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
    managementRequest.current = false;
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
    managementRequest.current = true;
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
      `${member.wordText} ${member.shortDefinition} ${member.learnerDefinition ?? ""}`
        .toLocaleLowerCase()
        .includes(query.trim().toLocaleLowerCase()),
    ) ?? [];
  // Carry this list and its current filter so word detail can return here.
  const listReturn = list ? listPageHref(list.id, query) : null;
  const memberHref = (wordSlug: string, meaningId: string) =>
    `/vocabulary/${encodeURIComponent(wordSlug)}${
      listReturn ? `?${new URLSearchParams({ fromList: listReturn })}` : ""
    }#${meaningAnchor(meaningId)}`;
  return (
    <>
      {!list ? (
        <div>
          <PageHeader
            title={
              missing
                ? "This list is no longer available"
                : "Your list could not be loaded"
            }
            back={{ href: "/lists", label: "Back to personal lists" }}
            lede={
              missing
                ? "It may have been deleted or belong to another account. Choose one of your current lists."
                : "Try loading its current status again."
            }
          />
          {missing ? (
            <Link href="/lists" className={`${listLink} mt-3`}>
              Choose a personal list
            </Link>
          ) : (
            <Button
              type="button"
              onClick={reload}
              disabled={action.busy}
              variant="secondary"
              className="mt-3"
            >
              Try loading list again
            </Button>
          )}
        </div>
      ) : (
        <>
          <PageHeader
            title={list.name}
            back={{ href: "/lists", label: "Back to personal lists" }}
            lede={`${list.memberCount} ${list.memberCount === 1 ? "meaning" : "meanings"} · ${list.usableMemberCount} available for practice`}
            actions={
              <>
                {list.usableMemberCount > 0 && (
                  <ActionLink
                    href={`/practice?list=${encodeURIComponent(list.id)}`}
                  >
                    Practice this list
                  </ActionLink>
                )}
                <Link href="/vocabulary" className="vn-text-link">
                  Find words to add
                </Link>
              </>
            }
          />
          <p className="mb-4 text-sm text-neutral-600">
            List membership does not save words for scheduled review or change
            what you marked as known.
          </p>
          {list.memberCount > 0 && list.usableMemberCount === 0 && (
            <p className="mb-4 text-sm text-neutral-600">
              Add a meaning available for practice to start a list session.
            </p>
          )}
          <div role="search" aria-label="Search list meanings" className="mb-4">
            <label
              htmlFor="list-member-search"
              className="font-semibold text-neutral-900"
            >
              Search words and meanings in this list
            </label>
            <TextInput
              id="list-member-search"
              type="search"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              maxLength={100}
              className="mt-2"
            />
          </div>
          {members.length ? (
            <ul aria-label="List meanings" className="vn-list">
              {members.map((member) => (
                <li key={member.meaningId} className="vn-list-row">
                  <div className="min-w-0 flex-[1_1_17rem] [overflow-wrap:anywhere]">
                    <h2 className="headword text-xl">
                      <Link
                        href={memberHref(member.wordSlug, member.meaningId)}
                        className={listLink}
                      >
                        {member.wordText}
                      </Link>
                    </h2>
                    <p className="text-neutral-700">
                      {member.partOfSpeech} ·{" "}
                      {getPrimaryDefinition(
                        member.shortDefinition,
                        member.learnerDefinition,
                      )}
                    </p>
                    <p className="mt-1 text-sm text-neutral-600">
                      {member.practiceAvailable
                        ? "Available for list practice"
                        : "Organized here; focused practice is not yet available for this meaning"}
                    </p>
                  </div>
                  <Button
                    type="button"
                    disabled={action.locked}
                    variant="secondary"
                    onClick={() => remove(member.meaningId)}
                    aria-label={`Remove ${member.wordText} from this list`}
                  >
                    Remove from list
                  </Button>
                </li>
              ))}
            </ul>
          ) : (
            <section className="vn-empty">
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
            </section>
          )}
          <details
            id="list-settings"
            ref={settingsRef}
            className="mt-6 border-t border-neutral-200"
          >
            <summary className="vn-summary scroll-my-24">List settings</summary>
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
              <TextInput
                id="rename-list-name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                disabled={action.locked}
                required
                maxLength={160}
                className="mt-2"
              />
              <Button
                type="submit"
                disabled={
                  action.locked ||
                  !name.trim() ||
                  Array.from(name.trim()).length > 80 ||
                  name.trim() === list.name
                }
                variant="secondary"
                className="mt-3"
              >
                Rename list
              </Button>
            </form>
            {!confirmDelete ? (
              <Button
                type="button"
                disabled={action.locked}
                onClick={() => setConfirmDelete(true)}
                variant="secondary"
                className="mt-3"
              >
                Delete list
              </Button>
            ) : (
              <div className="mt-4 border-t border-neutral-200 pt-4">
                <p className="text-neutral-700">
                  Delete this list and all its memberships? Your saved review
                  words and knowledge notes remain available.
                </p>
                <div className="mt-3 flex flex-wrap gap-3">
                  <Button
                    type="button"
                    disabled={action.locked}
                    onClick={deleteList}
                    variant="secondary"
                  >
                    Confirm delete list
                  </Button>
                  <Button
                    type="button"
                    disabled={action.locked}
                    onClick={() => setConfirmDelete(false)}
                    variant="secondary"
                  >
                    Keep list
                  </Button>
                </div>
              </div>
            )}
          </details>
        </>
      )}
      <ListActionStatus action={action} reload={reload} />
    </>
  );
}
