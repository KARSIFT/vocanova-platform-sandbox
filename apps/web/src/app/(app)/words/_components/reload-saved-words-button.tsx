"use client";

export function ReloadSavedWordsButton() {
  return (
    <button
      type="button"
      onClick={() => window.location.reload()}
      className="mt-3 inline-flex min-h-12 items-center font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
    >
      Try loading again
    </button>
  );
}
