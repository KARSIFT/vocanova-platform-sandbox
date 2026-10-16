import type { VocabularyAttribution } from "@vocanova/api-client";

export function VocabularySources({
  attribution,
}: {
  attribution: VocabularyAttribution;
}) {
  return (
    <details
      id="word-sources"
      className="mt-8 border-t border-neutral-200 pt-2 text-sm text-neutral-600"
    >
      <summary className="min-h-11 cursor-pointer content-center rounded-lg font-semibold focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
        Word sources
      </summary>
      <p className="mt-3">{attribution.title}</p>
      <div className="mt-2 flex flex-wrap gap-x-5">
        <a
          href={attribution.sourceUrl}
          className="inline-flex min-h-11 items-center font-semibold text-primary-700 underline"
        >
          Vocabulary source
        </a>
        <a
          href={attribution.licenseUrl}
          className="inline-flex min-h-11 items-center font-semibold text-primary-700 underline"
        >
          Catalog license
        </a>
      </div>
      <pre className="mt-3 whitespace-pre-wrap break-words text-xs leading-relaxed">
        {attribution.wordnetLicense}
      </pre>
    </details>
  );
}
