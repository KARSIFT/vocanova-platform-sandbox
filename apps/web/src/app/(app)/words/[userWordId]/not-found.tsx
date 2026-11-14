import { ActionLink } from "@/ui/controls";

export default function SavedWordNotFound() {
  return (
    <div className="p-[var(--spacing-lg)] text-center">
      <h1 className="text-2xl font-semibold text-neutral-900">
        Saved word not found
      </h1>
      <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
        It may have been removed from your vocabulary.
      </p>
      <ActionLink
        variant="secondary"
        href="/words"
        className="mt-[var(--spacing-lg)] min-w-[var(--spacing-2xl)] text-base text-neutral-50"
      >
        Back to saved vocabulary
      </ActionLink>
    </div>
  );
}
