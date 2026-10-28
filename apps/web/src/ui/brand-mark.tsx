export function BrandSymbol({ size = "md" }: { size?: "sm" | "md" }) {
  return (
    <span
      aria-hidden="true"
      className={`relative grid shrink-0 place-items-center bg-[linear-gradient(145deg,var(--color-primary-600),var(--color-primary-800))] font-bold tracking-[-0.08em] text-white shadow-[0_6px_14px_-6px_rgb(30_58_138_/_0.6)] ${size === "sm" ? "size-8 rounded-[0.6rem] text-sm" : "size-9 rounded-[0.7rem] text-sm"}`}
    >
      V
      <span className="absolute -right-0.5 -top-0.5 size-2.5 rounded-full border-2 border-white bg-secondary-400" />
    </span>
  );
}

export function BrandMark() {
  return (
    <span className="inline-flex items-center gap-[var(--spacing-sm)]">
      <BrandSymbol />
      <span className="leading-tight">
        <span className="block text-lg font-bold tracking-[-0.035em] text-neutral-900">
          VocaNova
        </span>
      </span>
    </span>
  );
}
