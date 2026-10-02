package lessons

import (
	"crypto/sha256"
	"fmt"
	"sort"
)

// Catalog versions change whenever an editorial prompt or answer changes.
// Existing sessions retain their snapshot, including the canonical teaching text.
type Definition struct {
	Key            string      `json:"key"`
	Version        string      `json:"version"`
	Title          string      `json:"title"`
	SituationSlug  string      `json:"situationSlug"`
	SituationTitle string      `json:"situationTitle"`
	Description    string      `json:"description"`
	Words          []Reference `json:"words"`
}
type Reference struct {
	MeaningID   string `json:"meaningId"`
	WordText    string `json:"wordText"`
	Context     string `json:"context"`
	Prompt      string `json:"prompt"`
	Explanation string `json:"explanation"`
}

func lesson(slug, title, situation string, words ...Reference) Definition {
	return lessonIn(slug, slug, title, situation, words...)
}

func lessonIn(key, slug, title, situation string, words ...Reference) Definition {
	return Definition{Key: key, Version: "1", Title: title, SituationSlug: slug, SituationTitle: situation, Description: "Learn three useful words, remember their meanings, and use them in a real-life situation.", Words: words}
}

// Explicit starter-course order; keys remain stable when presentation changes.
var catalog = []Definition{
	lessonIn("conversation-start", "daily-conversation", "Start a friendly conversation", "Daily Conversation",
		Reference{"ac93d068-7c3e-55c9-9d85-3fd3518592dc", "greeting", "A new neighbour comes to your door. You say hello and welcome her to the building.", "What have you offered her?", "A greeting is what you say to welcome someone or say hello."},
		Reference{"3265c02c-5751-5465-b94b-e767aa871d8d", "small talk", "You are waiting for a bus with a neighbour. You chat about the weather, with no serious subject to discuss.", "What kind of conversation is this?", "Small talk is light conversation about everyday things."},
		Reference{"0d267f75-f227-5478-962d-b4b944e0e93e", "casual", "You are chatting with friends in a park. Nobody uses formal language, and everyone feels relaxed.", "Which word describes this conversation?", "Casual describes a relaxed, informal style."},
	),
	lessonIn("conversation-find-time", "daily-conversation", "Find a time that works", "Daily Conversation",
		Reference{"c99075bf-94cb-5f51-90d5-f244f66ea942", "weekend plans", "It is Friday. Your friend asks what you are going to do on Saturday and Sunday.", "What is your friend asking about?", "Weekend plans are activities you intend to do on Saturday or Sunday."},
		Reference{"b2d5132c-b969-587f-975d-4009c454968f", "available", "A friend asks to meet at seven. You have no work or other plans at that time.", "Complete your reply: “Yes, I am ___ at seven.”", "Available means not busy and able to do something at a particular time."},
		Reference{"52b8ebc0-ba65-579d-8749-6ff51b9ce6b5", "sounds good", "A friend suggests a picnic. You like that idea and want to agree.", "Complete your reply: “A picnic ___!”", "Sounds good is a friendly way to say you like a suggestion."},
	),
	lessonIn("conversation-suggest-activity", "daily-conversation", "Suggest something to do", "Daily Conversation",
		Reference{"d602623a-1bf7-5656-b867-484d2be3f5a9", "suggest", "Your friends have not decided what to do. You offer an idea for them to consider.", "Complete your message: “I ___ going for a walk.”", "Suggest means offering an idea; the group has not agreed on it yet."},
		Reference{"0c030a86-51fc-5d13-9246-09950c45a43c", "join", "Your friends are already having lunch. They ask you to take part and eat with them.", "Complete their question: “Would you like to ___ us?”", "Join means taking part in an activity with other people."},
		Reference{"40ed986a-7261-54eb-a3c8-4213c3bdc4fd", "arrange", "Everyone has agreed to a picnic. Now you need to organize its time and place.", "Complete your message: “Let us ___ the picnic for Sunday afternoon.”", "Arrange means planning and organizing the details of an event."},
	),
	lesson("daily-conversation", "Make a plan with a friend", "Daily Conversation",
		Reference{"ee53d6ba-4303-5394-b7f9-79937ce66d09", "invite", "You are planning dinner at home. You want Sam to come, but you have not asked him yet.", "Complete your message: “I'd like to ___ you to dinner.”", "Invite means asking someone to come to an event or do something with you."},
		Reference{"6348aad5-937a-57d0-aebc-15887a771938", "confirm", "You and Sam agreed to meet at six. You send a message to say that six is definitely still the correct time.", "Complete your message: “I'm writing to ___ our meeting at six.”", "Confirm means stating that a planned time or arrangement is definite."},
		Reference{"0cdfa962-ee1c-570d-8533-999088905805", "reschedule", "You cannot meet on Tuesday anymore. You still want to meet Sam, and Friday works for both of you.", "Complete your message: “Could we ___ our meeting for Friday?”", "Reschedule means moving a planned event to another time or day."}),
	lessonIn("conversation-plan-changes", "daily-conversation", "Keep a plan on track", "Daily Conversation",
		Reference{"f98bfb7b-c140-55a6-b76c-774ee096d1ca", "cancel", "A storm is coming. You decide that the outdoor party will not happen, and you do not choose a new date.", "Complete your message: “We need to ___ the party.”", "Cancel means stopping a planned event from going ahead."},
		Reference{"b535cf81-3777-5553-959d-1ccf9d22a209", "on time", "Your friends agreed to meet at six. You arrive exactly at six, so nobody has to wait for you.", "Complete the sentence: “I arrived ___.”", "On time means at the planned time, without being late."},
		Reference{"c5454676-5349-52d2-ad14-814d516fed56", "meet up", "You and a friend want to see each other after work and spend an hour together.", "Complete your message: “Shall we ___ outside the station?”", "Meet up means meeting someone to spend time together."},
	),
	lessonIn("conversation-stay-connected", "daily-conversation", "Stay connected", "Daily Conversation",
		Reference{"3d64c3c9-ede0-5ffd-b1ef-278f6b70e486", "catch up", "You have not seen a friend for six months. You want to hear what has happened in her life.", "Complete your message: “Let us ___ over coffee.”", "Catch up means talking and sharing news after time apart."},
		Reference{"ddde743f-15ab-5d13-bea6-9835a56c991d", "keep in touch", "Your course ends tomorrow. You and a classmate want to continue sending messages over the next year.", "Complete your message: “Let us ___ after the course ends.”", "Keep in touch means continuing to communicate over time."},
		Reference{"c954cdb6-c0bc-5d61-9045-18db360f4b84", "farewell", "A friend is moving to another country for several years. You say a special goodbye at the station.", "What is this goodbye called?", "A farewell is a goodbye, especially before a long absence."},
	),
	lesson("restaurant", "Order and pay", "Restaurant",
		Reference{"329c4ec9-6562-568e-9e71-9134da2f0d8d", "menu", "You sit down at a restaurant. You want to see the dishes and prices before choosing your food.", "What do you ask to see?", "A menu lists the food and drinks you can order."},
		Reference{"2534b539-a80f-5001-b9a4-27a5e0679539", "appetizer", "You order a small bowl of soup to eat before your main meal.", "What kind of dish is the soup in this meal?", "An appetizer is a small dish served before the main meal."},
		Reference{"5d21c02b-ef51-5545-bd8a-01f5814d3a67", "bill", "You have finished eating. You want the statement showing the total amount you need to pay.", "What do you ask for?", "The bill shows the amount you need to pay for your meal."}),
	lessonIn("restaurant-other-plans", "restaurant", "Plan a meal out or at home", "Restaurant",
		Reference{"52c8ed28-544d-5a80-a911-6ddd03953606", "reservation", "You want a restaurant to keep a table for four people at seven. You phone before you arrive.", "What do you ask the restaurant to make?", "A reservation here is a booking for a restaurant table."},
		Reference{"1e9e55f8-d329-5c52-bc30-7ed77a31a167", "take-out", "You order cooked food from a restaurant, but you carry it home instead of eating there.", "What kind of food order is this?", "Take-out is prepared food bought to eat somewhere else."},
		Reference{"bd5eacb6-be32-5e7b-a8d7-d56db1908347", "tip", "After paying for your meal, you choose to leave extra money to thank the person who served you.", "What is this extra money called?", "A tip is extra money for service. Whether to leave one depends on the situation."},
	),
	lessonIn("shopping-clothes", "shopping", "Choose clothes that fit", "Shopping",
		Reference{"7dee60f4-25c4-5d55-9111-dc5174dcd5b5", "size", "The shop has the same shirt marked small, medium and large. You need a medium.", "Which detail are you choosing?", "Size tells you how big an item of clothing is."},
		Reference{"481f699b-4788-5620-b8a5-51347d90e6e5", "try on", "You like a jacket but do not know how it feels or looks on you. You ask to put it on before buying.", "Complete your question: “May I ___ this jacket?”", "Try on means putting on clothes or shoes before deciding whether to buy them."},
		Reference{"6c80ff44-eb45-5332-b48d-228872ead757", "fit", "The shoes are neither too tight nor too loose. They are the right size for your feet.", "Complete the sentence: “These shoes ___ me well.”", "Fit means being the right size and shape for someone."},
	),
	lessonIn("shopping-returns", "shopping", "Return a purchase", "Shopping",
		Reference{"74a55fe5-7736-58f0-98b8-23eef998067d", "receipt", "You have paid for a lamp. The shop gives you a record of the purchase and the amount paid.", "What has the shop given you?", "A receipt records a payment that has already been made."},
		Reference{"538cc7f5-983c-59cc-8664-173cefca65ad", "refund", "A lamp you bought is broken. You return it and ask for your money back, rather than another lamp.", "What are you asking for?", "A refund is money returned after a payment."},
		Reference{"c916143f-197b-5cd7-a4c9-7b21178f04aa", "exchange", "A shirt is too small. You want to return it and receive a larger shirt instead of your money back.", "Complete your question: “Can I ___ this shirt for a larger one?”", "Exchange means returning an item and receiving another in its place."},
	),
	lessonIn("home-rental", "home-and-renting", "Deal with a home problem", "Home and Renting",
		Reference{"e507b507-5432-51a0-bcc3-be12294143db", "rent", "You live in a flat owned by someone else. You pay money each month to continue living there.", "What is this regular payment called?", "Rent is the money paid to use a home owned by someone else."},
		Reference{"75be2229-a161-5e08-856f-143098221439", "landlord", "A person owns the flat you live in and lets you use it in return for a monthly payment.", "What is this person called?", "A landlord rents out a home or other property."},
		Reference{"3c382279-cf27-5d6b-ba75-a0389f1d5680", "repair", "The door lock is broken. Someone fixes it so the door can close safely again.", "What is this work called?", "A repair is work done to fix something damaged or broken."},
	),
	lessonIn("transport-station", "public-transport", "Find the right service", "Public Transport",
		Reference{"bf1bee5c-b8f0-5ebc-9252-53c4ac9ef7de", "bus stop", "You want to catch a bus. You wait beside a roadside sign where passengers get on and off.", "Where are you waiting?", "A bus stop is a place where a bus picks up or drops off passengers."},
		Reference{"6d08080d-4855-5652-ae35-cf0c2bcb9222", "timetable", "You need to know when buses leave on Sunday. You look at a list showing each service and its departure time.", "What are you reading?", "A timetable lists the times when transport services arrive or leave."},
		Reference{"9aa8876d-25c6-54d1-9b65-0816e604e7a5", "platform", "Your train is arriving. You wait in the numbered area beside the railway track.", "Where are you waiting?", "A platform is the area next to the track where passengers wait for a train."},
	),
	lessonIn("transport-tickets", "public-transport", "Buy a ticket and change trains", "Public Transport",
		Reference{"faf13c57-a106-5406-a2bc-e743eb00ec4e", "single ticket", "You are travelling to another town today. You do not want this purchase to include a journey back.", "Which ticket do you ask for?", "A single ticket covers travel in one direction only."},
		Reference{"35b47079-26eb-5441-bf1d-05060d125a7e", "return ticket", "You are travelling to another town and coming home that evening. You want one ticket covering both journeys.", "Which ticket do you ask for?", "A return ticket covers the outward journey and the journey back."},
		Reference{"91888c86-1fc6-5633-a6dd-a46017f8fdf7", "change", "This train does not go all the way to your destination. At the next station, you must leave it and take another train.", "Complete the instruction: “You need to ___ trains at the next station.”", "Change here means moving from one transport service to another."},
	),
	lesson("airport", "Find your flight", "Airport",
		Reference{"8a6ec801-3292-5fc2-bfb3-242ac405d891", "boarding pass", "You are ready to get on the plane. A staff member asks to scan the document with your flight and seat details.", "What do you show?", "A boarding pass is the document scanned before you get on your flight."},
		Reference{"6a61d337-80e7-5d50-bc61-8b45b6a054a1", "gate", "The screen says your flight leaves from B12. You need to walk to the place where passengers board.", "Which place are you looking for?", "The gate is the place where you wait and board your plane."},
		Reference{"bee2408e-5991-5805-923c-6fba33ab7e7a", "luggage", "You have two suitcases and a travel bag. An airport worker asks what you are bringing on the trip.", "Which word describes these bags together?", "Luggage means the bags and suitcases you take on a trip."}),
	lessonIn("airport-checks", "airport", "Move through the airport", "Airport",
		Reference{"2a610b3d-4031-5eee-a0be-d3def2cc0b53", "security check", "Before entering the departure area, you put your bag through a scanner and follow the screening staff instructions.", "Which process are you going through?", "A security check screens passengers and bags before they enter the departure area."},
		Reference{"d62f1f3f-8382-5999-9e9c-30858f6b2c5b", "customs", "After an international flight, an officer asks about goods you are bringing into the country.", "Which border check is this?", "Customs checks goods brought across a border."},
		Reference{"8f12f8e0-d0a3-5010-941b-8f7d526121cc", "layover", "Your journey uses two flights. You spend three hours at an airport between the first flight and the next one.", "What is this stop between flights called?", "A layover is a stop between connecting flights."},
	),
	lesson("hotel-check-in", "Arrive at your hotel", "Hotel Check-in",
		Reference{"edb20c6e-1023-5f10-af4b-0dc059a781b2", "reservation", "You booked a hotel room online last week. You arrive and say that the booking is under your name.", "Which word means your booking?", "A reservation is a booking for a room at the hotel."},
		Reference{"3008733d-9cab-547c-9712-4014a26ca697", "front desk", "You are in the hotel lobby. You need to find the counter where staff help guests check in.", "Where do you go?", "The front desk is the hotel counter where staff help guests."},
		Reference{"d990ce6e-4756-52e0-96e8-9926442963f2", "key card", "A member of staff gives you a plastic electronic card. You hold it against the lock to open your room.", "What did the staff member give you?", "A key card is the electronic card used to open your room."}),
	lessonIn("hotel-services", "hotel-check-in", "Use hotel services", "Hotel Check-in",
		Reference{"dfe9c593-cf2e-505b-96ca-4afef46422d8", "amenities", "A hotel has a swimming pool, a gym and free internet for its guests.", "What general word describes these useful features?", "Amenities are useful services and features offered by a hotel."},
		Reference{"de015c5a-0fad-507d-944e-572f6a10da66", "wake-up call", "You must get up early for a flight. You ask hotel staff to telephone your room at six in the morning.", "What service are you requesting?", "A wake-up call is a phone call arranged to wake you at a particular time."},
		Reference{"846f3dde-4b90-5eb8-929e-ecd28e8315f3", "check-out", "Your hotel stay has ended. You return the room card and settle the bill before leaving.", "What is this process called?", "Check-out is the process of leaving a hotel at the end of your stay."},
	),
	lessonIn("delivery-parcel", "deliveries", "Receive a parcel", "Deliveries",
		Reference{"0b593408-6679-5fb8-a21c-b73a525c76e5", "parcel", "Your friend sends you a book wrapped in paper and packed for posting.", "What is this wrapped object called?", "A parcel is an object or group of objects wrapped for sending."},
		Reference{"3bb8e710-2cfb-5f6f-a281-c0a167d46fb2", "address", "A delivery driver needs your building number, street and town to find your home.", "What information does the driver need?", "An address gives the details of a location."},
		Reference{"4609031d-af20-550a-906e-3fd3cde7588e", "delivery", "You buy a table online. The shop brings it to your home the next day.", "What is the service of bringing it called?", "Delivery means bringing goods to a person or place."},
	),
	lessonIn("money-payment", "everyday-payments", "Pay for everyday things", "Everyday Payments",
		Reference{"b20c1447-0e63-5712-9ae7-dd4aca9aeb85", "cash", "At a shop, you pay with coins and paper notes instead of using an electronic method.", "What are you paying with?", "Cash is money in the form of coins and notes."},
		Reference{"d78438f3-30bd-58bf-9b96-1489c4f57d3b", "bank card", "At a shop, you tap a plastic card linked to your account on the machine to pay.", "What are you using?", "A bank card lets you pay using a bank or credit account."},
		Reference{"c6a6c8f7-1ce0-59df-b963-e07066796992", "payment", "An online shop confirms that it has received the money you sent for your order.", "Complete the message: “Your ___ has been received.”", "A payment is an amount of money paid, or the act of paying it."},
	),
	lessonIn("services-form", "everyday-services", "Complete a simple form", "Everyday Services",
		Reference{"b548ade3-8682-521e-bf26-848526519854", "form", "An office gives you a document with empty spaces for your name, address and other details.", "What do you need to complete?", "A form is a document with spaces for information."},
		Reference{"c7bbc8ae-3f64-58e9-acf0-3069e317c471", "signature", "You have finished filling in a document. At the bottom, you write your name in the usual way you sign things.", "What have you added?", "A signature is your name written in your usual signing style."},
		Reference{"64360906-46e2-588e-96d7-c78213875c9b", "queue", "Several people arrived at a service counter before you. You stand behind them and wait your turn.", "What have you joined?", "A queue is a line of people waiting for something."},
	),
	lessonIn("health-services", "health-appointments", "Ask for health services", "Health Appointments",
		Reference{"c76cb86d-be5a-5375-8d88-72bda7b98e7d", "appointment", "You call a clinic and agree to see a doctor at ten on Tuesday.", "What have you arranged?", "An appointment is an arranged meeting at a particular time."},
		Reference{"096a46ea-93b8-59c2-9423-bc07f3332297", "symptom", "A doctor asks what health changes you have noticed. You describe a cough that started yesterday.", "What is the cough in this conversation?", "A symptom is a change that may be a sign of a health problem; it does not by itself identify an illness."},
		Reference{"c3a26504-d404-5151-bbac-5e22511aae5f", "pharmacy", "You want to find a place where trained staff supply medicines. You ask which shop nearby provides that service.", "What place are you looking for?", "A pharmacy is a place or service where people can get medicines."},
	),
	lessonIn("phone-calls", "phone-calls", "Handle a phone call", "Phone Calls",
		Reference{"5966f508-b502-5669-a7d6-c619ce92639f", "signal", "Your phone cannot connect to the mobile network inside a building. It connects again when you go outside.", "What was missing inside the building?", "A signal here is the mobile connection needed for the phone to communicate."},
		Reference{"2fd7a730-c9e9-5a7b-b094-dea868bc7cd3", "voicemail", "You call a friend, but she does not answer. You record a spoken message for her to hear later.", "What have you left?", "A voicemail is a recorded phone message."},
		Reference{"e1b649b8-b6d8-5caf-8671-5b9e57ebd71c", "call back", "A friend phones while you are busy. You decide to phone him after your meeting.", "Complete your reply: “I will ___ after my meeting.”", "Call back means phoning someone after they have phoned you."},
	),
	lessonIn("digital-email", "email-and-online-tasks", "Exchange an email", "Email and Online Tasks",
		Reference{"ee851d91-6476-5942-9454-8f1b2db92682", "attachment", "An email says that a document has been sent with it as a separate file.", "What is this file called?", "An attachment is a file included with a message."},
		Reference{"ae343b98-d972-5565-8bac-87f04fe96d06", "link", "An email contains an item you can click to open a booking page in your browser.", "What is this clickable item called?", "A link takes you to another online page or resource."},
		Reference{"2d1adc4c-058b-5d12-ae17-86afc4195588", "reply", "A friend sends you an email with a question. You want to answer that message.", "Complete your plan: “I will ___ to the email tonight.”", "Reply means answering a message or question."},
	),
	lessonIn("work-day", "everyday-work", "Talk about your working day", "Everyday Work",
		Reference{"f97ad45b-01ee-51f9-b77a-ddab1881b374", "shift", "Your working hours today run from seven in the morning until three in the afternoon.", "What is this scheduled work period called?", "A shift is a scheduled period of work."},
		Reference{"ea64a58a-2521-50fa-8af9-a5cf9e013763", "break", "You stop working for fifteen minutes to rest, then return to the same task.", "What is this short rest called?", "A break is a short pause in work."},
		Reference{"8ecac8d4-b7c0-51bb-95be-14a73d8637e3", "colleague", "Someone works on the same team as you. You often discuss work tasks together.", "What is this person in relation to you?", "A colleague is someone you work with."},
	),
	lesson("job-interview", "Prepare for an interview", "Job Interview",
		Reference{"e5817d4f-eaf9-577e-881f-b100fe1e24f8", "resume", "The interviewer asks for a document listing your previous jobs, education, and skills.", "Which document do you send?", "A resume gives a summary of your experience and skills."},
		Reference{"9ab971f4-9b7b-5551-b274-62ad41d9273f", "references", "The company wants to contact people who know your work. You give the names of two former managers.", "What are these people called in a job application?", "References are people who can speak about your work."},
		Reference{"0b9878b3-64a7-5175-92f7-ce0435cfbed4", "salary expectations", "The interviewer asks how much money you hope to earn in the new job.", "What is the interviewer asking about?", "Salary expectations are the pay range you hope to receive."}),
	lessonIn("interview-application", "job-interview", "Send a job application", "Job Interview",
		Reference{"7ee5fc1b-8108-5fb4-a254-f75af84695f3", "cover letter", "You are applying for a job. Alongside your work-history document, you send a letter explaining your interest in this particular role.", "What is this letter called?", "A cover letter introduces your job application to an employer."},
		Reference{"c7796a82-cc8d-56de-a97f-773ea694e337", "qualifications", "An employer asks about the skills and experience that make you suitable for a job.", "What is the employer asking about?", "Qualifications here are the relevant skills and experience you bring to a job."},
		Reference{"4c3724d8-5abf-5159-8171-76ccdcdae7c0", "follow-up", "Your interview ended yesterday. Today you send a short message thanking the interviewer and asking about the next step.", "What kind of message is this?", "A follow-up here is a message sent after a job interview."},
	),
	lesson("work-meeting", "Leave a meeting with a plan", "Work Meeting",
		Reference{"8a7f86d3-ed7a-512c-9c19-d86b5d46ffa8", "agenda", "Before a meeting, your manager sends a list of the topics everyone will discuss.", "What is this list called?", "An agenda lists the topics planned for a meeting."},
		Reference{"71c851c2-c028-5d8f-9e0a-5be524410a82", "action items", "The meeting ends. You must send a report, and your colleague must call a customer. These are the tasks the team agreed to do.", "What are these tasks called?", "Action items are the tasks assigned during or after a meeting."},
		Reference{"b571a540-89bb-533b-be81-9cedf443a64a", "deadline", "Your manager says the report must be finished no later than Friday at noon.", "What does Friday at noon represent?", "A deadline is the latest time something must be finished."}),
	lessonIn("meeting-collaboration", "work-meeting", "Share ideas and progress", "Work Meeting",
		Reference{"1eee9c5f-7079-54d4-bcd8-183739df837a", "brainstorm", "Your team has not chosen a solution yet. Everyone shares possible ideas before judging which will work.", "Complete the suggestion: “Let us ___ some ideas.”", "Brainstorm means sharing ideas freely to help solve a problem."},
		Reference{"1e91dacc-7700-5f65-9310-312ac4dff706", "stakeholder", "A project affects a customer team. One member of that team has an interest in its result and is invited to give feedback.", "What is this person in relation to the project?", "A stakeholder has an interest in a project, for example because its result affects them."},
		Reference{"e62a2ba4-c6b7-5186-98f6-20d69f4dc196", "update", "Your manager knows about the project but wants to hear what has changed since last week.", "Complete the request: “Could you give us an ___ on the project?”", "An update gives new information about progress or the current situation."},
	),
	lesson("university-class", "Get ready for your course", "University Class",
		Reference{"fd6c39c0-f097-5a5b-b947-c93096e635d0", "lecture", "You sit with other students while an instructor gives a lesson about history.", "What are you attending?", "A lecture is a lesson given by an instructor to a class."},
		Reference{"ae72824e-1341-5081-84ea-0e0fdbf09c50", "assignment", "Your teacher asks you to write a short essay and bring it to the next class.", "What is this piece of work called?", "An assignment is work a teacher asks students to complete."},
		Reference{"58529fdd-9425-5aef-a8c7-bc1086be7080", "syllabus", "At the start of a course, your teacher gives you a document describing its topics, schedule, and requirements.", "What is this course document called?", "A syllabus is a document describing a course."}),
	lessonIn("study-collaboration", "university-class", "Work with other students", "University Class",
		Reference{"5b7f44dd-77ba-5d20-9ee4-01d6967d203a", "group project", "Three students must work together to research a topic and produce one piece of work.", "What is this shared assignment called?", "A group project is an assignment completed with other students."},
		Reference{"ce20a0cc-6a7e-5b5e-bbfe-5ec49f1c58d5", "office hours", "Your instructor sets aside Tuesday afternoons for students to visit and ask course questions.", "What are these scheduled help times called?", "Office hours are times when students can meet an instructor."},
		Reference{"c1e69c4e-2915-5c99-96ab-d946e67653fd", "presentation", "Your group stands in front of the class and gives a talk about its research, using a few slides.", "What is your group giving?", "A presentation is a talk that explains a topic to an audience."},
	),
}

func definition(key string) (Definition, bool) {
	for _, d := range catalog {
		if d.Key == key {
			return d, true
		}
	}
	return Definition{}, false
}

func buildSnapshot(d Definition, words []Word, seed string) (Snapshot, error) {
	if len(words) != 3 || len(d.Words) != 3 {
		return Snapshot{}, ErrContentUnavailable
	}
	snap := Snapshot{Definition: d, Words: words, Steps: make([]privateStep, 0, 9)}
	for i, w := range words {
		if w.MeaningID != d.Words[i].MeaningID || w.WordText != d.Words[i].WordText || w.Definition == "" || w.Example == "" {
			return Snapshot{}, ErrContentUnavailable
		}
	}
	for _, kind := range []string{"teach", "recall", "context"} {
		for i, w := range words {
			id := fmt.Sprintf("%s-%d", kind, i+1)
			step := privateStep{Step: Step{ID: id, Kind: kind, Word: w, Choices: []Choice{}}, Explanations: map[string]string{}}
			if kind == "teach" {
				step.Step.Prompt = "Meet a useful word"
			} else {
				step.CorrectChoiceID = w.MeaningID
				if kind == "recall" {
					step.Step.Prompt = "Which meaning matches “" + w.WordText + "”?"
				} else {
					step.Step.Prompt = d.Words[i].Prompt
					step.Step.Context = d.Words[i].Context
				}
				for _, option := range words {
					text := option.Definition
					if kind == "context" {
						text = option.WordText
					}
					step.Step.Choices = append(step.Step.Choices, Choice{ID: option.MeaningID, Text: text})
					explanation := w.Definition
					if kind == "context" {
						explanation = d.Words[i].Explanation
					}
					if option.MeaningID != w.MeaningID {
						explanation = "“" + option.WordText + "”: " + option.Definition + " Here, “" + w.WordText + "” fits. " + explanation
					}
					step.Explanations[option.MeaningID] = explanation
				}
				// Stable within a persisted session, different for each learner's session.
				sort.Slice(step.Step.Choices, func(a, b int) bool {
					ha := sha256.Sum256([]byte(seed + id + step.Step.Choices[a].ID))
					hb := sha256.Sum256([]byte(seed + id + step.Step.Choices[b].ID))
					return string(ha[:]) < string(hb[:])
				})
			}
			snap.Steps = append(snap.Steps, step)
		}
	}
	return snap, nil
}
