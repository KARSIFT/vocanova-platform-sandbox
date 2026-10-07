import type { DictionaryEntry } from "@vocanova/api-client";
import { ListenButton } from "@/ui/pronunciation";

export function DictionaryCard({ entry }: { entry: DictionaryEntry }) {
  const senses = entry.meanings.flatMap((meaning) =>
    meaning.definitions.map((definition) => ({
      ...definition,
      partOfSpeech: meaning.partOfSpeech,
    })),
  );
  const [first, ...others] = senses;
  if (!first) return null;
  return (
    <section
      aria-label="Dictionary result"
      className="mb-6 rounded-2xl border border-neutral-200 bg-white p-5 sm:p-7"
    >
      <p className="text-sm font-semibold text-neutral-600">Dictionary</p>
      <div className="mt-2 flex flex-wrap items-center justify-between gap-3">
        <h2 className="break-words text-3xl font-bold tracking-tight text-neutral-900">
          {entry.word}
        </h2>
      </div>
      <p className="mt-3 text-sm text-neutral-600">{first.partOfSpeech}</p>
      <p className="mt-2 max-w-prose text-lg leading-relaxed text-neutral-900">
        {first.definition}
      </p>
      {first.example && (
        <p className="mt-3 max-w-prose text-neutral-700">“{first.example}”</p>
      )}
      <ListenButton text={entry.word} showCaption={false} />
      {others.length > 0 && (
        <details className="mt-4 border-t border-neutral-200 pt-2">
          <summary className="min-h-11 cursor-pointer content-center font-semibold text-primary-700">
            More meanings
          </summary>
          <ul className="space-y-5 py-3">
            {others.map((sense, index) => (
              <li key={index}>
                <p className="text-sm text-neutral-600">{sense.partOfSpeech}</p>
                <p className="mt-1 max-w-prose text-neutral-900">
                  {sense.definition}
                </p>
                {sense.example && (
                  <p className="mt-2 text-neutral-700">“{sense.example}”</p>
                )}
              </li>
            ))}
          </ul>
        </details>
      )}
      <details className="mt-3 text-sm text-neutral-600">
        <summary className="min-h-11 cursor-pointer content-center">
          Dictionary source
        </summary>
        <a
          href={entry.attribution.providerUrl}
          className="inline-flex min-h-11 items-center font-semibold text-primary-700 underline underline-offset-4"
        >
          {entry.attribution.provider}
        </a>
        {entry.attribution.licenses.map((license) => (
          <div key={license.name}>
            <a
              href={license.url}
              className="inline-flex min-h-11 items-center text-primary-700 underline underline-offset-4"
            >
              {license.name} license
            </a>
            <p className="max-w-prose whitespace-pre-line break-words leading-relaxed">
              {license.text}
            </p>
          </div>
        ))}
      </details>
    </section>
  );
}
