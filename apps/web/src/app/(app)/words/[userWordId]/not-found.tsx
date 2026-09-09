import Link from "next/link";

export default function SavedWordNotFound() {
  return (
    <div className="p-[var(--spacing-lg)] text-center">
      <h1 className="text-2xl font-semibold text-neutral-900">
        Saved word not found
      </h1>
      <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
        It may have been removed from your vocabulary.
      </p>
      <Link
        href="/words"
        className="mt-[var(--spacing-lg)] inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
      >
        Back to saved vocabulary
      </Link>
    </div>
  );
}
