import Link from "next/link";
import { notFound } from "next/navigation";

// Placeholder local data for VOC-022 static UI; replace with real API wiring in a follow-up package.
const MOCK_SITUATION_WORD_LISTS = {
  airport: {
    title: "Airport",
    words: [
      {
        term: "boarding pass",
        meaning: "A document that lets you get on your flight.",
        isSaved: true,
      },
      {
        term: "gate",
        meaning: "The place where you board your plane.",
        isSaved: false,
      },
      {
        term: "security check",
        meaning: "The screening process before entering the departure area.",
        isSaved: false,
      },
      {
        term: "luggage",
        meaning: "Bags and suitcases you take on a trip.",
        isSaved: true,
      },
      { term: "customs", meaning: "A border check for goods.", isSaved: false },
      { term: "layover", meaning: "A stop between flights.", isSaved: false },
    ],
  },
  restaurant: {
    title: "Restaurant",
    words: [
      {
        term: "menu",
        meaning: "A list of food and drinks you can order.",
        isSaved: true,
      },
      {
        term: "reservation",
        meaning: "A booking for a table.",
        isSaved: false,
      },
      {
        term: "appetizer",
        meaning: "A small dish eaten before the main meal.",
        isSaved: false,
      },
      {
        term: "bill",
        meaning: "The statement showing how much you need to pay.",
        isSaved: true,
      },
      {
        term: "tip",
        meaning: "Extra money given to thank someone for service.",
        isSaved: false,
      },
      {
        term: "take-out",
        meaning: "Food to eat somewhere else.",
        isSaved: false,
      },
    ],
  },
  "hotel-check-in": {
    title: "Hotel Check-in",
    words: [
      {
        term: "reservation",
        meaning: "A booking for a room at a hotel.",
        isSaved: true,
      },
      {
        term: "front desk",
        meaning: "The hotel counter for guest help.",
        isSaved: false,
      },
      {
        term: "key card",
        meaning: "An electronic card used to open your room.",
        isSaved: false,
      },
      {
        term: "amenities",
        meaning: "Useful hotel services or features.",
        isSaved: true,
      },
      {
        term: "wake-up call",
        meaning: "A scheduled phone call to wake you up.",
        isSaved: false,
      },
      {
        term: "check-out",
        meaning: "The process of leaving a hotel.",
        isSaved: false,
      },
    ],
  },
  "job-interview": {
    title: "Job Interview",
    words: [
      {
        term: "resume",
        meaning: "A summary of your work experience and skills.",
        isSaved: true,
      },
      {
        term: "qualifications",
        meaning: "Skills and experience for a job.",
        isSaved: false,
      },
      {
        term: "references",
        meaning: "People who can speak about your work.",
        isSaved: false,
      },
      {
        term: "salary expectations",
        meaning: "The pay range you hope to receive.",
        isSaved: true,
      },
      {
        term: "cover letter",
        meaning: "A letter for a job application.",
        isSaved: false,
      },
      {
        term: "follow-up",
        meaning: "A message sent after an interview.",
        isSaved: false,
      },
    ],
  },
  "daily-conversation": {
    title: "Daily Conversation",
    words: [
      {
        term: "small talk",
        meaning: "Light conversation about everyday topics.",
        isSaved: true,
      },
      {
        term: "catch up",
        meaning: "To talk after time apart.",
        isSaved: false,
      },
      {
        term: "weekend plans",
        meaning: "Activities planned for the weekend.",
        isSaved: false,
      },
      {
        term: "greeting",
        meaning: "Words used to say hello to someone.",
        isSaved: true,
      },
      {
        term: "farewell",
        meaning: "Words used to say goodbye to someone.",
        isSaved: false,
      },
      {
        term: "casual",
        meaning: "Relaxed and informal in style or conversation.",
        isSaved: false,
      },
    ],
  },
  "work-meeting": {
    title: "Work Meeting",
    words: [
      {
        term: "agenda",
        meaning: "A list of topics planned for a meeting.",
        isSaved: true,
      },
      {
        term: "action items",
        meaning: "Tasks assigned after a meeting.",
        isSaved: false,
      },
      {
        term: "deadline",
        meaning: "The latest time something must be finished.",
        isSaved: false,
      },
      {
        term: "follow-up",
        meaning: "A later message or action.",
        isSaved: true,
      },
      {
        term: "stakeholder",
        meaning: "A person interested in a project.",
        isSaved: false,
      },
      {
        term: "brainstorm",
        meaning: "To share ideas freely to solve a problem.",
        isSaved: false,
      },
    ],
  },
  "university-class": {
    title: "University Class",
    words: [
      {
        term: "lecture",
        meaning: "A lesson given by an instructor to a class.",
        isSaved: true,
      },
      {
        term: "assignment",
        meaning: "Work a teacher asks students to complete.",
        isSaved: false,
      },
      {
        term: "syllabus",
        meaning: "A document describing a course.",
        isSaved: false,
      },
      {
        term: "office hours",
        meaning: "Times to meet an instructor.",
        isSaved: true,
      },
      {
        term: "group project",
        meaning: "An assignment with other students.",
        isSaved: false,
      },
      {
        term: "deadline",
        meaning: "When work must be submitted.",
        isSaved: false,
      },
    ],
  },
} as const;

type SituationSlug = keyof typeof MOCK_SITUATION_WORD_LISTS;

export default async function SituationDiscoverPage({
  params,
}: {
  params: Promise<{ situation: string }>;
}) {
  const { situation } = await params;
  const situationWords = MOCK_SITUATION_WORD_LISTS[situation as SituationSlug];

  if (!situationWords) {
    notFound();
  }

  return (
    <div className="p-[var(--spacing-lg)]">
      <Link
        href="/discover"
        className="text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
      >
        Back to Journey
      </Link>
      <h1 className="mt-[var(--spacing-md)] text-2xl font-semibold text-neutral-900">
        {situationWords.title}
      </h1>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        Words already saved are marked below.
      </p>

      <ul className="mt-[var(--spacing-lg)] space-y-[var(--spacing-md)]">
        {situationWords.words.map((word) => (
          <li
            key={word.term}
            className="rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
          >
            <div className="flex items-start justify-between gap-[var(--spacing-md)]">
              <div>
                <p className="text-lg font-semibold text-neutral-900">
                  {word.term}
                </p>
                <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
                  {word.meaning}
                </p>
              </div>
              {word.isSaved ? (
                <span className="shrink-0 rounded-full bg-primary-100 px-[var(--spacing-sm)] py-[var(--spacing-xs)] text-sm font-semibold text-primary-800">
                  <span aria-hidden="true">✓</span> Saved
                </span>
              ) : null}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
