import type { WordMeaning } from "@vocanova/api-client";
import { formatNoteType } from "@/lib/word-definition";
import { ListenButton } from "@/ui/pronunciation";

/** Render only the authored content attached to this specific meaning. */
export function MeaningTeaching({ meaning }: { meaning: WordMeaning }) {
  const groups = new Map<string, WordMeaning["usageNotes"]>();
  for (const note of meaning.usageNotes) {
    if (!note.noteText.trim()) continue;
    const group = groups.get(note.noteType) ?? [];
    group.push(note);
    groups.set(note.noteType, group);
  }
  return (
    <>
      {meaning.examples.length > 0 && (
        <div className="mt-5">
          <h3 className="text-lg font-semibold text-neutral-900">
            Example sentences
          </h3>
          <ul className="mt-2 space-y-3 text-neutral-700">
            {meaning.examples.map((example) => (
              <li
                key={example.id}
                className="rounded-xl border border-neutral-200 bg-neutral-50 p-3"
              >
                <p>{example.exampleText}</p>
                <ListenButton
                  text={example.exampleText}
                  label={`example: ${example.exampleText}`}
                  showCaption={false}
                />
              </li>
            ))}
          </ul>
        </div>
      )}
      {groups.size > 0 && (
        <div className="mt-5">
          <h3 className="text-lg font-semibold text-neutral-900">
            Usage notes
          </h3>
          <ul className="mt-3 grid gap-3 sm:grid-cols-2">
            {[...groups].map(([type, notes]) => (
              <li
                key={type}
                className="min-w-0 rounded-xl border border-neutral-200 p-4"
              >
                <h4 className="font-semibold text-neutral-900">
                  {formatNoteType(type)}
                </h4>
                <ul className="mt-2 space-y-3 text-neutral-700">
                  {notes.map((note) => (
                    <li key={note.id}>
                      <p
                        className={
                          type === "collocation"
                            ? "font-medium text-neutral-900"
                            : undefined
                        }
                      >
                        {note.noteText}
                      </p>
                      {type === "collocation" && (
                        <ListenButton
                          text={note.noteText}
                          label={`combination: ${note.noteText}`}
                        />
                      )}
                    </li>
                  ))}
                </ul>
              </li>
            ))}
          </ul>
        </div>
      )}
    </>
  );
}
