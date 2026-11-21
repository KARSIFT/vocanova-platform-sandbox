"use client";

import { Button, TextInput } from "@/ui/controls";
import { useLayoutEffect, useRef, useState } from "react";
import Link from "@/ui/link";
import { LearnIcon } from "../../_components/learn-icons";
import { useRouter } from "next/navigation";
import type { WordListsResponse } from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { listWriteHeaders, useListAction } from "./use-list-action";
import { ListActionStatus } from "./list-action-status";
import { listLink } from "./list-styles";

export function ListsCollection({
  initialData,
}: {
  initialData: WordListsResponse | null;
}) {
  const router = useRouter();
  const [data, setData] = useState(initialData);
  const [name, setName] = useState("");
  const [query, setQuery] = useState("");
  const action = useListAction();
  const createRef = useRef<HTMLDetailsElement>(null);
  useLayoutEffect(() => {
    if (action.error && createRef.current) createRef.current.open = true;
  }, [action.error]);
  const items =
    data?.items.filter((item) =>
      item.name.toLocaleLowerCase().includes(query.trim().toLocaleLowerCase()),
    ) ?? [];
  const reload = () =>
    void action.reload(async () => {
      setData((await createApiClient().listWordLists()).data);
    });
  function create() {
    const id = crypto.randomUUID();
    const key = crypto.randomUUID();
    const body = { name: name.trim(), expectedRevision: 0 };
    void action.run(async () => {
      const { data: list } = await createApiClient().putWordList(
        id,
        body,
        key,
        await listWriteHeaders(),
      );
      setData((current) => ({
        items: [
          ...(current?.items.filter((item) => item.id !== list.id) ?? []),
          list,
        ],
      }));
      setName("");
      router.push(`/lists/${encodeURIComponent(list.id)}`);
      return "Your list is ready.";
    });
  }
  return (
    <>
      {!data ? (
        <section className="vn-empty">
          <h2 className="text-xl font-bold text-neutral-900">
            Your lists are unavailable
          </h2>
          <p className="mt-2 text-neutral-700">
            Load your lists before creating or editing one.
          </p>
          <Button
            type="button"
            variant="secondary"
            className="mt-3"
            disabled={action.busy}
            onClick={reload}
          >
            Try loading lists again
          </Button>
        </section>
      ) : (
        <>
          <div
            className="mb-4"
            role="search"
            aria-label="Search personal lists"
          >
            <label
              htmlFor="list-search"
              className="font-semibold text-neutral-900"
            >
              Search your lists
            </label>
            <TextInput
              id="list-search"
              type="search"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              maxLength={100}
              className="mt-2"
            />
          </div>
          {items.length ? (
            <ul aria-label="Personal lists" className="vn-list">
              {items.map((list) => (
                <li key={list.id}>
                  <Link
                    href={`/lists/${encodeURIComponent(list.id)}`}
                    className="vn-list-row flex-nowrap rounded-sm hover:bg-neutral-50 focus-visible:-outline-offset-4"
                    aria-labelledby={`list-name-${list.id}`}
                  >
                    <div className="min-w-0 [overflow-wrap:anywhere]">
                      <h2
                        id={`list-name-${list.id}`}
                        className="text-lg font-semibold text-neutral-900"
                      >
                        {list.name}
                      </h2>
                      <p className="mt-1 text-sm text-neutral-600">
                        {list.memberCount}{" "}
                        {list.memberCount === 1 ? "meaning" : "meanings"} ·{" "}
                        {list.usableMemberCount} available for practice
                      </p>
                    </div>
                    <span className="shrink-0 text-primary-700">
                      <LearnIcon name="arrow" />
                    </span>
                  </Link>
                </li>
              ))}
            </ul>
          ) : (
            <section className="vn-empty">
              <h2 className="text-xl font-bold text-neutral-900">
                {query.trim()
                  ? "No lists match your search"
                  : "Give your words a place"}
              </h2>
              <p className="mt-2 text-neutral-700">
                {query.trim()
                  ? "Try a different name or clear your search."
                  : "Create your first list, then open a word and choose which meaning to add."}
              </p>
              <Link href="/vocabulary" className={`${listLink} mt-3`}>
                Find useful words
              </Link>
            </section>
          )}
          <details
            id="create-list"
            ref={createRef}
            open={!data.items.length}
            className="mt-6 border-t border-neutral-200"
          >
            <summary className="vn-summary scroll-my-24">Create a list</summary>
            <form
              className="mt-3"
              onSubmit={(event) => {
                event.preventDefault();
                create();
              }}
            >
              <label
                htmlFor="new-list-name"
                className="font-semibold text-neutral-900"
              >
                List name
              </label>
              <TextInput
                id="new-list-name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                disabled={action.locked || data.items.length >= 50}
                maxLength={160}
                required
                aria-describedby="list-name-help"
                className="mt-2"
              />
              <p id="list-name-help" className="mt-2 text-sm text-neutral-600">
                Use 1–80 characters. You can keep up to 50 personal lists.
              </p>
              <Button
                type="submit"
                disabled={
                  action.locked ||
                  !name.trim() ||
                  Array.from(name.trim()).length > 80 ||
                  data.items.length >= 50
                }
                className="mt-3"
              >
                Create list
              </Button>
            </form>
            {data.items.length >= 50 && (
              <p className="mt-2 text-neutral-700">
                You have 50 lists. Delete a list you no longer need before
                creating another.
              </p>
            )}
          </details>
        </>
      )}
      <ListActionStatus action={action} reload={reload} />
    </>
  );
}
