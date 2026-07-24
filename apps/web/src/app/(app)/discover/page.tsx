// Placeholder local data for VOC-021 static UI; replace with real API wiring in a follow-up package.
const MOCK_DISCOVER_SITUATIONS = [
  {
    slug: "airport",
    title: "Airport",
    shortDescription:
      "Check in, get through security, and find your gate with confidence.",
  },
  {
    slug: "restaurant",
    title: "Restaurant",
    shortDescription:
      "Order food, ask about the menu, and handle the bill with ease.",
  },
  {
    slug: "hotel-check-in",
    title: "Hotel Check-in",
    shortDescription:
      "Check in and out smoothly and ask about hotel amenities.",
  },
  {
    slug: "job-interview",
    title: "Job Interview",
    shortDescription:
      "Talk about your experience and answer common interview questions.",
  },
  {
    slug: "daily-conversation",
    title: "Daily Conversation",
    shortDescription:
      "Handle everyday small talk and casual exchanges naturally.",
  },
  {
    slug: "work-meeting",
    title: "Work Meeting",
    shortDescription:
      "Follow along, share updates, and contribute to workplace discussions.",
  },
  {
    slug: "university-class",
    title: "University Class",
    shortDescription:
      "Understand lectures, ask questions, and join class discussion.",
  },
] as const;

export default function DiscoverPage() {
  return (
    <div className="p-[var(--spacing-lg)]">
      <h1 className="text-2xl font-semibold text-neutral-900">Journey</h1>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        Choose a situation to explore practical vocabulary.
      </p>

      <ul className="mt-[var(--spacing-lg)] grid grid-cols-1 gap-[var(--spacing-md)] sm:grid-cols-2">
        {MOCK_DISCOVER_SITUATIONS.map((situation) => (
          <li
            key={situation.slug}
            className="rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
          >
            <h2 className="text-lg font-semibold text-neutral-900">
              {situation.title}
            </h2>
            <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
              {situation.shortDescription}
            </p>
          </li>
        ))}
      </ul>
    </div>
  );
}
