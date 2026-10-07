// Small visual cues for the authored situation catalogue. The adjacent title
// supplies the accessible name; these icons never encode learning progress.
const paths: Record<string, string> = {
  airport: "m3 13 7-3V4l2-2 2 2v6l7 3v2l-7-1v5l2 2H8l2-2v-5l-7 1z",
  restaurant: "M4 3v6a3 3 0 0 0 6 0V3M7 3v18M20 21V3c-4 0-5 7-5 10h5",
  "hotel-check-in": "M3 21V4h18v17M8 21v-5h8v5M7 8h2m6 0h2M7 12h2m6 0h2",
  "job-interview":
    "M5 8h14a2 2 0 0 1 2 2v9H3v-9a2 2 0 0 1 2-2Zm3 0V4h8v4M3 13h18M10 13v3h4v-3",
  "daily-conversation": "M4 3h16v12h-9l-5 4v-4H4zM8 7h8M8 11h5",
  "work-meeting": "M3 3h18v13H3zM7 21l5-5 5 5M8 8h8M8 12h5",
  "university-class": "m2 8 10-5 10 5-10 5zM6 10v7c4 3 8 3 12 0v-7M22 8v9",
  shopping: "M5 7h14l2 14H3zM8 8V6a4 4 0 0 1 8 0v2",
  "home-and-renting": "m2 11 10-9 10 9M5 9v12h14V9M9 21v-7h6v7",
  "public-transport":
    "M6 3h12a2 2 0 0 1 2 2v13H4V5a2 2 0 0 1 2-2ZM4 11h16M8 3v8M7 15h1m8 0h1M6 18v3m12-3v3",
  deliveries: "m3 7 9-4 9 4v11l-9 4-9-4zM3 7l9 4 9-4M12 11v11M7 5l10 4",
  "everyday-payments": "M3 5h18v14H3zM3 10h18M7 15h4",
  "everyday-services": "M5 3h14v18H5zM8 7h8M8 11h8M8 15h5",
  "health-appointments": "M8 3h8v5h5v8h-5v5H8v-5H3V8h5z",
  "phone-calls":
    "M5 3h4l2 5-3 2a14 14 0 0 0 6 6l2-3 5 2v4c0 2-3 3-5 2C8 19 5 16 3 8 2 6 3 3 5 3Z",
  "email-and-online-tasks": "M3 5h18v14H3zM3 5l9 8 9-8",
  "everyday-work": "M3 5h18v12H3zM8 21h8M12 17v4M7 9h5m-5 4h9",
};

const aliases: Record<string, string> = {
  "ordering-at-a-cafe": "restaurant",
  "navigating-an-airport": "airport",
  transport: "public-transport",
  work: "everyday-work",
  friends: "daily-conversation",
  hotel: "hotel-check-in",
};

export function SituationIcon({ slug }: { slug: string }) {
  const path = paths[aliases[slug] ?? slug];
  if (!path) return null;
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      className="size-6 shrink-0 text-primary-700"
    >
      <path d={path} />
    </svg>
  );
}
