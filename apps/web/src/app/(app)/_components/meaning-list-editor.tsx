"use client";

import { useId, useState } from "react";
import Link from "next/link";
import type { WordListDetail, WordListSummary } from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import {
  listWriteHeaders,
  useListAction,
} from "../lists/_components/use-list-action";
import { ListActionStatus } from "../lists/_components/list-action-status";
import {
  listButton,
  listInput,
  listLink,
} from "../lists/_components/list-styles";

export function MeaningListEditor({ meaningId }: { meaningId: string }) {
  const id = useId();
  const [open, setOpen] = useState(false);
  const [lists, setLists] = useState<WordListSummary[] | null>(null);
  const [selectedId, setSelectedId] = useState("");
  const [detail, setDetail] = useState<WordListDetail | null>(null);
  const action = useListAction();
  const member =
    detail?.members.some((item) => item.meaningId === meaningId) ?? false;
  const reload = () => {
    setOpen(true);
    void action.reload(async () => {
      const { data } = await createApiClient().listWordLists();
      setLists(data.items);
      if (selectedId && data.items.some((item) => item.id === selectedId)) {
        setDetail((await createApiClient().getWordList(selectedId)).data);
      } else {
        setSelectedId("");
        setDetail(null);
      }
    });
  };
  function choose(listId: string) {
    if (action.locked) return;
    setSelectedId(listId);
    setDetail(null);
    if (listId)
      void action.reload(async () => {
        setDetail((await createApiClient().getWordList(listId)).data);
      });
  }
  function changeMembership() {
    if (!detail) return;
    const listId = detail.id;
    const revision = detail.revision;
    const key = crypto.randomUUID();
    const remove = member;
    void action.run(async () => {
      const client = createApiClient();
      const init = await listWriteHeaders();
      const { data } = remove
        ? await client.deleteWordListMember(
            listId,
            meaningId,
            revision,
            key,
            init,
          )
        : await client.putWordListMember(
            listId,
            meaningId,
            { expectedRevision: revision },
            key,
            init,
          );
      setDetail(data);
      setLists(
        (current) =>
          current?.map((item) => (item.id === data.id ? data : item)) ?? null,
      );
      // Exact retries return current status, which may differ after another edit.
      return data.members.some((item) => item.meaningId === meaningId)
        ? "This meaning is currently in the list."
        : "This meaning is currently outside the list.";
    });
  }
  return (
    <section
      aria-labelledby={`${id}-heading`}
      className="mt-6 border-t border-neutral-200 pt-5"
    >
      <h3
        id={`${id}-heading`}
        className="text-lg font-semibold text-neutral-900"
      >
        Personal lists
      </h3>
      <p className="mt-2 text-sm text-neutral-600">
        Organize this meaning without changing scheduled review or your known
        status.
      </p>
      {!open ? (
        <button type="button" className={`${listButton} mt-3`} onClick={reload}>
          Choose personal lists
        </button>
      ) : (
        <>
          {lists && lists.length > 0 && (
            <div className="mt-3">
              <label
                htmlFor={`${id}-select`}
                className="font-semibold text-neutral-900"
              >
                Choose a list for this meaning
              </label>
              <select
                id={`${id}-select`}
                value={selectedId}
                disabled={action.locked}
                onChange={(event) => choose(event.target.value)}
                className={listInput}
              >
                <option value="">Choose a personal list</option>
                {lists.map((list) => (
                  <option key={list.id} value={list.id}>
                    {list.name}
                  </option>
                ))}
              </select>
              {detail && (
                <>
                  <p className="mt-2 text-neutral-700">
                    {member
                      ? "This meaning is in this list."
                      : "This meaning is not in this list."}
                  </p>
                  <button
                    type="button"
                    className={`${listButton} mt-3`}
                    disabled={
                      action.locked || (!member && detail.memberCount >= 500)
                    }
                    onClick={changeMembership}
                  >
                    {member
                      ? "Remove meaning from list"
                      : "Add meaning to list"}
                  </button>
                  {!member && detail.memberCount >= 500 && (
                    <p className="mt-2 text-neutral-700">
                      This list has reached its 500-meaning limit.
                    </p>
                  )}
                  <Link
                    href={`/lists/${encodeURIComponent(detail.id)}`}
                    className={`${listLink} ml-3`}
                  >
                    Open list
                  </Link>
                </>
              )}
            </div>
          )}
          {lists?.length === 0 && (
            <p className="mt-3 text-neutral-700">
              Create a personal list first, then return here to add this
              meaning.
            </p>
          )}
          {!lists && !action.busy && (
            <button
              type="button"
              className={`${listButton} mt-3`}
              onClick={reload}
            >
              Try loading lists again
            </button>
          )}
          <Link href="/lists" className={`${listLink} mt-3`}>
            Manage personal lists
          </Link>
        </>
      )}
      <ListActionStatus action={action} reload={reload} />
    </section>
  );
}
