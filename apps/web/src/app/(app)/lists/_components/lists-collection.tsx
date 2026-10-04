"use client";
import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import type { WordListsResponse } from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { Surface } from "@/ui/surface";
import { listWriteHeaders, useListAction } from "./use-list-action";
import { ListActionStatus } from "./list-action-status";
import { listButton, listInput, listLink, listPrimary } from "./list-styles";

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
        <Surface>
          <h2 className="text-xl font-bold text-neutral-900">
            Your lists are unavailable
          </h2>
          <p className="mt-2 text-neutral-700">
            Load your lists before creating or editing one.
          </p>
          <button
            type="button"
            className={`${listButton} mt-3`}
            disabled={action.busy}
            onClick={reload}
          >
            Try loading lists again
          </button>
        </Surface>
      ) : (
        <>
          <Surface aria-labelledby="create-list-heading">
            <h2
              id="create-list-heading"
              className="text-xl font-bold text-neutral-900"
            >
              Create a list
            </h2>
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
              <input
                id="new-list-name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                disabled={action.locked || data.items.length >= 50}
                maxLength={160}
                required
                aria-describedby="list-name-help"
                className={listInput}
              />
              <p id="list-name-help" className="mt-2 text-sm text-neutral-600">
                Use 1–80 characters. You can keep up to 50 personal lists.
              </p>
              <button
                type="submit"
                disabled={
                  action.locked ||
                  !name.trim() ||
                  Array.from(name.trim()).length > 80 ||
                  data.items.length >= 50
                }
                className={`${listPrimary} mt-3`}
              >
                Create list
              </button>
            </form>
            {data.items.length >= 50 && (
              <p className="mt-2 text-neutral-700">
                You have 50 lists. Delete a list you no longer need before
                creating another.
              </p>
            )}
          </Surface>
          <div
            className="my-5"
            role="search"
            aria-label="Search personal lists"
          >
            <label
              htmlFor="list-search"
              className="font-semibold text-neutral-900"
            >
              Search your lists
            </label>
            <input
              id="list-search"
              type="search"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              maxLength={100}
              className={listInput}
            />
          </div>
          {items.length ? (
            <ul
              aria-label="Personal lists"
              className="grid gap-4 sm:grid-cols-2"
            >
              {items.map((list) => (
                <li
                  key={list.id}
                  className="min-w-0 rounded-2xl border border-neutral-200 bg-white p-5"
                >
                  <h2 className="break-words text-xl font-bold text-neutral-900">
                    <Link
                      href={`/lists/${encodeURIComponent(list.id)}`}
                      className={listLink}
                    >
                      {list.name}
                    </Link>
                  </h2>
                  <p className="mt-2 text-neutral-700">
                    {list.memberCount}{" "}
                    {list.memberCount === 1 ? "meaning" : "meanings"} ·{" "}
                    {list.usableMemberCount} available for practice
                  </p>
                </li>
              ))}
            </ul>
          ) : (
            <Surface>
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
            </Surface>
          )}
        </>
      )}
      <ListActionStatus action={action} reload={reload} />
    </>
  );
}
