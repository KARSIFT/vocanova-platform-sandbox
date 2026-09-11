"use client";

import { buildIdentity } from "@/lib/build-identity";

function formatBuildTime(value: string): string {
  return value === "unknown"
    ? value
    : value.replace("T", " ").replace(/\.\d{3}Z$/, " UTC");
}

const environmentLabel = {
  development: "Development",
  production: "Production",
  staging: "Staging",
  test: "Test",
  unknown: "Unknown",
} as const;

export function BuildIdentity() {
  const commit =
    buildIdentity.commit === "unknown"
      ? buildIdentity.commit
      : buildIdentity.commit.slice(0, 12);

  return (
    <section
      aria-labelledby="about-vocanova-heading"
      className="mt-[var(--spacing-lg)] rounded-[var(--radius-lg)] border border-neutral-200 bg-white p-[var(--spacing-md)] shadow-sm sm:p-[var(--spacing-lg)]"
    >
      <h2
        id="about-vocanova-heading"
        className="text-lg font-semibold text-neutral-900"
      >
        About VocaNova
      </h2>
      <dl className="mt-[var(--spacing-sm)] grid gap-x-[var(--spacing-lg)] gap-y-[var(--spacing-xs)] text-sm sm:grid-cols-[auto_1fr]">
        <dt className="font-medium text-neutral-700">Version</dt>
        <dd className="font-mono text-neutral-900">{buildIdentity.version}</dd>
        <dt className="font-medium text-neutral-700">Environment</dt>
        <dd className="text-neutral-900">
          {environmentLabel[buildIdentity.environment]}
        </dd>
        <dt className="font-medium text-neutral-700">Commit</dt>
        <dd className="font-mono text-neutral-900">{commit}</dd>
        <dt className="font-medium text-neutral-700">Built (UTC)</dt>
        <dd className="text-neutral-900">
          {formatBuildTime(buildIdentity.builtAt)}
        </dd>
      </dl>
    </section>
  );
}
