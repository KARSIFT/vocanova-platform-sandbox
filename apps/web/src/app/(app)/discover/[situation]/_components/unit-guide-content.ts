export interface GuidePhrase {
  text: string;
  purpose: string;
  word: string;
}
export interface UnitGuide {
  goal: string;
  tip: string;
  phrases: readonly GuidePhrase[];
}

// Original Vocanova teaching text. The word anchors are checked against the
// canonical situation catalog; these are practical goals, not level estimates.
export const unitGuides: Readonly<Record<string, UnitGuide>> = {
  airport: {
    goal: "Find your flight, ask for directions, and explain what you are carrying.",
    tip: "Check the flight number as well as the destination. Ask again if a gate changes.",
    phrases: [
      {
        text: "Could you check my boarding pass, please?",
        purpose: "Ask staff to confirm your travel document.",
        word: "boarding pass",
      },
      {
        text: "Which gate should I go to?",
        purpose: "Find the place where you board.",
        word: "gate",
      },
      {
        text: "Is this bag allowed as hand luggage?",
        purpose: "Ask about the bag you want to take on board.",
        word: "luggage",
      },
    ],
  },
  restaurant: {
    goal: "Ask about the food, order politely, and pay for your meal.",
    tip: "Use 'Could I have…?' for a polite request. Tell the server clearly if you need information about ingredients.",
    phrases: [
      {
        text: "Could I have the menu, please?",
        purpose: "Ask to see the food and drink choices.",
        word: "menu",
      },
      {
        text: "We have a reservation for two people.",
        purpose: "Explain that you booked a table.",
        word: "reservation",
      },
      {
        text: "Could we have the bill, please?",
        purpose: "Ask for the amount to pay.",
        word: "bill",
      },
    ],
  },
  "hotel-check-in": {
    goal: "Confirm your room booking and ask about hotel services.",
    tip: "Say the name on your booking. Confirm breakfast and check-out times before you leave the front desk.",
    phrases: [
      {
        text: "I have a reservation for two nights.",
        purpose: "Explain the length of your stay.",
        word: "reservation",
      },
      {
        text: "My key card does not open the door.",
        purpose: "Explain a problem with room access.",
        word: "key card",
      },
      {
        text: "What time is check-out?",
        purpose: "Find out when you need to leave the room.",
        word: "check-out",
      },
    ],
  },
  "job-interview": {
    goal: "Describe your experience and ask clear questions about a job.",
    tip: "Give one concrete example of your experience. If a question is unclear, ask the interviewer to repeat it.",
    phrases: [
      {
        text: "I have included my recent experience in my resume.",
        purpose: "Point to your work-history document.",
        word: "resume",
      },
      {
        text: "Which qualifications are needed for this role?",
        purpose: "Ask about required skills or certificates.",
        word: "qualifications",
      },
      {
        text: "I can provide references from my previous manager.",
        purpose: "Offer someone who can describe your work.",
        word: "references",
      },
    ],
  },
  "daily-conversation": {
    goal: "Start a friendly conversation and arrange a time to meet.",
    tip: "Offer a specific time and place. 'Sounds good' accepts an idea; 'I'm not available' explains a timing problem.",
    phrases: [
      {
        text: "Are you available on Saturday afternoon?",
        purpose: "Check when a friend is free.",
        word: "available",
      },
      {
        text: "Would you like to join us for a walk?",
        purpose: "Invite someone to take part.",
        word: "join",
      },
      {
        text: "Can we reschedule our coffee for Friday?",
        purpose: "Suggest a different time.",
        word: "reschedule",
      },
    ],
  },
  "work-meeting": {
    goal: "Follow a meeting and agree on the next actions.",
    tip: "Before the meeting ends, confirm who will do each task and when it is due.",
    phrases: [
      {
        text: "Could we look at the agenda first?",
        purpose: "Ask to review the planned discussion topics.",
        word: "agenda",
      },
      {
        text: "What is the deadline for this task?",
        purpose: "Confirm the latest time to finish.",
        word: "deadline",
      },
      {
        text: "I'll send an update tomorrow morning.",
        purpose: "Promise fresh information about progress.",
        word: "update",
      },
    ],
  },
  "university-class": {
    goal: "Ask about class tasks, deadlines, and support from a teacher.",
    tip: "Name the assignment when asking a question. Check both the due date and how to submit it.",
    phrases: [
      {
        text: "When is this assignment due?",
        purpose: "Ask when the task must be handed in.",
        word: "assignment",
      },
      {
        text: "Could we review the syllabus together?",
        purpose: "Ask about the course plan.",
        word: "syllabus",
      },
      {
        text: "May I visit during your office hours?",
        purpose: "Ask for help at the teacher's scheduled time.",
        word: "office hours",
      },
    ],
  },
  shopping: {
    goal: "Find a good fit and ask about returning or exchanging an item.",
    tip: "Explain the problem with the item. Keep the receipt and ask about the shop's return policy.",
    phrases: [
      {
        text: "May I try on this jacket?",
        purpose: "Ask to put it on and check the fit.",
        word: "try on",
      },
      {
        text: "Do you have this in a larger size?",
        purpose: "Ask for a different clothing measurement.",
        word: "size",
      },
      {
        text: "Can I exchange this for a smaller one?",
        purpose: "Ask to replace the item with another.",
        word: "exchange",
      },
    ],
  },
  "home-and-renting": {
    goal: "Ask about renting a home and explain a repair problem.",
    tip: "Describe where the problem is and when you noticed it. Confirm what is included before agreeing to rent.",
    phrases: [
      {
        text: "Does the rent include the water bill?",
        purpose: "Ask what the regular housing payment covers.",
        word: "rent",
      },
      {
        text: "How can I contact the landlord?",
        purpose: "Find the person responsible for renting out the home.",
        word: "landlord",
      },
      {
        text: "The kitchen tap needs a repair.",
        purpose: "Explain something that needs fixing.",
        word: "repair",
      },
    ],
  },
  "public-transport": {
    goal: "Choose a ticket and find the right place to board.",
    tip: "Check the direction of travel as well as the departure time. A return ticket covers both journeys.",
    phrases: [
      {
        text: "Where is the nearest bus stop?",
        purpose: "Find a place where the bus picks up passengers.",
        word: "bus stop",
      },
      {
        text: "Which platform does the train leave from?",
        purpose: "Find the correct boarding place.",
        word: "platform",
      },
      {
        text: "I'd like a return ticket to the city centre.",
        purpose: "Ask for travel there and back.",
        word: "return ticket",
      },
    ],
  },
  deliveries: {
    goal: "Confirm where a parcel should go and ask about delivery.",
    tip: "Check the full address, including the flat number. Use the tracking number if you have one.",
    phrases: [
      {
        text: "I'm waiting for a parcel.",
        purpose: "Explain that a package should arrive.",
        word: "parcel",
      },
      {
        text: "Could you confirm the delivery address?",
        purpose: "Check where the package will be sent.",
        word: "address",
      },
      {
        text: "What time is the delivery expected?",
        purpose: "Ask when the item should arrive.",
        word: "delivery",
      },
    ],
  },
  "everyday-payments": {
    goal: "Ask how to pay and confirm whether a payment worked.",
    tip: "Confirm the amount before paying. If a payment fails, ask staff to check its status before trying again.",
    phrases: [
      {
        text: "Can I pay in cash?",
        purpose: "Ask whether notes and coins are accepted.",
        word: "cash",
      },
      {
        text: "Do you accept this bank card?",
        purpose: "Check that your card can be used.",
        word: "bank card",
      },
      {
        text: "Could you check whether the payment went through?",
        purpose: "Ask staff to confirm the payment status.",
        word: "payment",
      },
    ],
  },
  "everyday-services": {
    goal: "Complete a simple service request and ask for help with a form.",
    tip: "Read the instructions before signing. Ask which details are required if a box is unclear.",
    phrases: [
      {
        text: "Could you help me complete this form?",
        purpose: "Ask for help with a written application.",
        word: "form",
      },
      {
        text: "Where does my signature go?",
        purpose: "Ask where to sign your name.",
        word: "signature",
      },
      {
        text: "Is this the queue for the service desk?",
        purpose: "Check that you are waiting in the right line.",
        word: "queue",
      },
    ],
  },
  "health-appointments": {
    goal: "Arrange a visit and clearly describe a symptom.",
    tip: "For language practice, describe what you feel and when it began. These phrases do not assess a health condition.",
    phrases: [
      {
        text: "I'd like to make an appointment.",
        purpose: "Ask to arrange a time to see a professional.",
        word: "appointment",
      },
      {
        text: "My main symptom is a sore throat.",
        purpose: "Describe the problem you are experiencing.",
        word: "symptom",
      },
      {
        text: "Where is the nearest pharmacy?",
        purpose: "Ask where medicines are supplied.",
        word: "pharmacy",
      },
    ],
  },
  "phone-calls": {
    goal: "Explain a poor connection and arrange another call.",
    tip: "If you cannot hear, ask the speaker to repeat the last part. Confirm a number slowly.",
    phrases: [
      {
        text: "The signal is weak. Could you repeat that?",
        purpose: "Explain why you did not hear clearly.",
        word: "signal",
      },
      {
        text: "Please leave a voicemail if I don't answer.",
        purpose: "Ask the caller to record a message.",
        word: "voicemail",
      },
      {
        text: "Can I call back in ten minutes?",
        purpose: "Ask to phone again shortly.",
        word: "call back",
      },
    ],
  },
  "email-and-online-tasks": {
    goal: "Send a file, share a link, and request a reply.",
    tip: "Name the file or task in your message. Check that the attachment is included before sending.",
    phrases: [
      {
        text: "The form is in the attachment.",
        purpose: "Explain where to find the file sent with the email.",
        word: "attachment",
      },
      {
        text: "Could you send me the link?",
        purpose: "Ask for an address to open an online page.",
        word: "link",
      },
      {
        text: "Please reply when you have checked the document.",
        purpose: "Ask for a response after a task.",
        word: "reply",
      },
    ],
  },
  "everyday-work": {
    goal: "Ask about a workday and coordinate with a colleague.",
    tip: "Confirm times clearly. When asking for help, say which task you need help with.",
    phrases: [
      {
        text: "What time does my shift start?",
        purpose: "Ask when your working period begins.",
        word: "shift",
      },
      {
        text: "Can I take my break after this task?",
        purpose: "Ask about a short rest from work.",
        word: "break",
      },
      {
        text: "My colleague will help with the delivery.",
        purpose: "Explain who will support the task.",
        word: "colleague",
      },
    ],
  },
};

const storySituations: Readonly<Record<string, string>> = {
  restaurant: "restaurant",
  transport: "public-transport",
  work: "everyday-work",
  shopping: "shopping",
  friends: "daily-conversation",
  hotel: "hotel-check-in",
};
export function storySituationSlug(situation: string): string | undefined {
  return storySituations[situation];
}
