const paths = {
  practice: "m10 5 9 7-9 7V5ZM5 5v14",
  stories:
    "M12 6c-3-2-6-2-9-1v14c3-1 6-1 9 1m0-14c3-2 6-2 9-1v14c-3-1-6-1-9 1V6Z",
  situations: "M3 20h18M5 20V8h14v12M3 8l9-5 9 5M9 20v-6h6v6",
  lists: "M9 6h11M9 12h11M9 18h11M4 6h.01M4 12h.01M4 18h.01",
  plan: "M7 3v4m10-4v4M4 10h16M4 5h16v16H4V5Zm4 9 3 3 5-5",
  discover: "m12 3 3 6 6 3-6 3-3 6-3-6-6-3 6-3 3-6Z",
  history: "M4 5v5h5M4 10a8 8 0 1 1 1 7m7-10v5l3 2",
  listen: "m11 4-6 5H2v6h3l6 5V4Zm4 4a6 6 0 0 1 0 8m3-11a10 10 0 0 1 0 14",
  profile: "M16 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0ZM4 21v-2a8 8 0 0 1 16 0v2",
  security: "M12 3 4 6v6c0 4 4 7 8 9 4-2 8-5 8-9V6l-8-3Zm-4 9 3 3 5-6",
  arrow: "m9 5 7 7-7 7",
} as const;

/** Decorative icons accompany visible learning-navigation labels. */
export function LearnIcon({ name }: { name: keyof typeof paths }) {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      className="size-5 shrink-0 fill-none stroke-current stroke-[1.8]"
    >
      <path d={paths[name]} strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
