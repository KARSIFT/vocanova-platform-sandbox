import type { SituationMeaning } from "@vocanova/api-client";

export interface ConversationPracticeChoice {
  meaningId: string;
  wordText: string;
  wordSlug: string;
  explanation: string;
}

export interface ConversationPracticeCase {
  id: string;
  title: string;
  context: string;
  prompt: string;
  choices: readonly ConversationPracticeChoice[];
  correctMeaningId: string;
}

const DAILY_CONVERSATION_ID = "d9f0dfc7-bc05-5595-bbf2-035c4a61ca9c";

// Editorial references to the existing canonical senses. Display text comes
// from the current situation response; these names guard against content drift.
const references = {
  invite: {
    meaningId: "ee53d6ba-4303-5394-b7f9-79937ce66d09",
    wordText: "invite",
    wordSlug: "invite",
  },
  join: {
    meaningId: "0c030a86-51fc-5d13-9246-09950c45a43c",
    wordText: "join",
    wordSlug: "join",
  },
  soundsGood: {
    meaningId: "52b8ebc0-ba65-579d-8749-6ff51b9ce6b5",
    wordText: "sounds good",
    wordSlug: "sounds-good",
  },
  keepInTouch: {
    meaningId: "ddde743f-15ab-5d13-bea6-9835a56c991d",
    wordText: "keep in touch",
    wordSlug: "keep-in-touch",
  },
  reschedule: {
    meaningId: "0cdfa962-ee1c-570d-8533-999088905805",
    wordText: "reschedule",
    wordSlug: "reschedule",
  },
  cancel: {
    meaningId: "f98bfb7b-c140-55a6-b76c-774ee096d1ca",
    wordText: "cancel",
    wordSlug: "cancel",
  },
} as const;

type ReferenceKey = keyof typeof references;
type PracticeMeaning = Pick<
  SituationMeaning,
  "meaningId" | "wordText" | "wordSlug"
>;

export function getConversationPractice(
  situationId: string,
  meanings: readonly PracticeMeaning[],
): ConversationPracticeCase[] {
  if (situationId !== DAILY_CONVERSATION_ID) return [];

  // An optional activity must not teach an absent sense or link to a different
  // word. Keep the normal situation list available if its references drift.
  for (const reference of Object.values(references)) {
    const matches = meanings.filter(
      (meaning) => meaning.meaningId === reference.meaningId,
    );
    if (
      matches.length !== 1 ||
      matches[0]!.wordSlug !== reference.wordSlug ||
      matches[0]!.wordText !== reference.wordText
    ) {
      return [];
    }
  }

  function choice(
    key: ReferenceKey,
    explanation: string,
  ): ConversationPracticeChoice {
    const reference = references[key];
    const meaning = meanings.find(
      (item) => item.meaningId === reference.meaningId,
    )!;
    return {
      meaningId: meaning.meaningId,
      wordText: meaning.wordText,
      wordSlug: meaning.wordSlug,
      explanation,
    };
  }

  return [
    {
      id: "invite-a-friend",
      title: "Ask a friend to dinner",
      context: "You are planning dinner at home. You want Sam to come.",
      prompt: "Complete your message: “I'd like to ___ you to dinner.”",
      choices: [
        choice("invite", "Invite is what you do when you ask Sam to come."),
        choice(
          "join",
          "Join means taking part with other people. Sam joins you if he takes part.",
        ),
      ],
      correctMeaningId: references.invite.meaningId,
    },
    {
      id: "respond-to-an-idea",
      title: "Respond to an idea",
      context: "Sam says, “How about lunch on Saturday?” You like the idea.",
      prompt: "Which reply fits?",
      choices: [
        choice(
          "keepInTouch",
          "Keep in touch means continuing to contact someone over time. It does not say that you like this lunch suggestion.",
        ),
        choice(
          "soundsGood",
          "Sounds good is a friendly way to say you like the idea. You may still need to agree on the time and place.",
        ),
      ],
      correctMeaningId: references.soundsGood.meaningId,
    },
    {
      id: "change-a-plan",
      title: "Move lunch to another day",
      context:
        "Tuesday is no longer possible. You still want lunch with Sam on Friday.",
      prompt: "Complete your message: “Could we ___ lunch for Friday?”",
      choices: [
        choice(
          "reschedule",
          "Reschedule means moving a planned event to a different time or day. You want to keep the lunch plan and move it to Friday.",
        ),
        choice(
          "cancel",
          "Cancel means stopping a plan from going ahead. It does not by itself say that the event has a new date.",
        ),
      ],
      correctMeaningId: references.reschedule.meaningId,
    },
  ];
}
