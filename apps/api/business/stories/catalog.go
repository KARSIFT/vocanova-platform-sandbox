package stories

import (
	"encoding/json"
	"fmt"
)

// All dialogue and questions are authored for Vocanova. Every question has one
// answer directly supported by preceding lines; distractors are distinct facts or
// incompatible phrases. Editorial changes require a new content version.
func vocab(text, slug, id, definition string) StoryVocabulary {
	return StoryVocabulary{text, slug, id, definition}
}
func line(speaker, text string) privateStep {
	return privateStep{Public: StoryStep{Kind: "line", Line: &StoryLine{Speaker: speaker, Text: text}}}
}
func question(kind, prompt string, correct int, explanation string, options ...string) privateStep {
	step := privateStep{Public: StoryStep{Kind: kind, Prompt: prompt, Choices: []StoryChoice{}}, Explanation: explanation}
	for i, text := range options {
		id := fmt.Sprintf("choice-%d", i+1)
		step.Public.Choices = append(step.Public.Choices, StoryChoice{id, text})
		if i == correct {
			step.CorrectChoice = id
		}
	}
	return step
}
func story(key, title, description, situation string, v []StoryVocabulary, steps ...privateStep) Snapshot {
	item := StoryCatalogItem{Key: key, Title: title, Description: description, Situation: situation, Level: "A2–B1", ContentVersion: ContentVersion, Vocabulary: v}
	for i := range steps {
		steps[i].Public.ID = fmt.Sprintf("%s-%d", key, i+1)
		if steps[i].Public.Line != nil {
			steps[i].Public.Line.ID = steps[i].Public.ID
			item.LineCount++
		} else {
			item.QuestionCount++
		}
	}
	return Snapshot{Story: item, ContentVersion: ContentVersion, GradingVersion: GradingVersion, Steps: steps}
}

var catalog = []Snapshot{
	story("a-quiet-lunch", "A quiet lunch", "Order a light lunch before a meeting.", "restaurant",
		[]StoryVocabulary{
			vocab("menu", "menu", "329c4ec9-6562-568e-9e71-9134da2f0d8d", "A list of food and drinks you can order."),
			vocab("appetizer", "appetizer", "2534b539-a80f-5001-b9a4-27a5e0679539", "A small dish before the main meal."),
			vocab("bill", "bill", "5d21c02b-ef51-5545-bd8a-01f5814d3a67", "The amount you need to pay for your meal."),
		},
		line("Narrator", "Nora has a meeting at two. She stops at a small restaurant for lunch."),
		line("Server", "Here is the menu. Would you like a table by the window?"),
		line("Nora", "Yes, please. I only have forty minutes."),
		line("Server", "The soup is a small appetizer. The vegetable sandwich is a full meal."),
		question("comprehension", "Why does Nora need a quick lunch?", 1, "Nora has a meeting at two and says she only has forty minutes.", "She is meeting the server at the station.", "She has a meeting at two.", "The restaurant is closing now."),
		line("Nora", "I'll have the vegetable sandwich and a glass of water, please."),
		line("Server", "Of course. Your sandwich will be ready in ten minutes."),
		line("Nora", "That gives me enough time. Could I have the bill with my food?"),
		line("Narrator", "Nora pays after eating and arrives at her meeting on time."),
		question("phrase_completion", "Complete Nora's request: Could I have the ___ with my food?", 2, "Nora asks for the bill so she can pay promptly after her lunch.", "window", "meeting", "bill"),
	),
	story("the-right-platform", "The right platform", "Ask for help when a train timetable changes.", "transport",
		[]StoryVocabulary{
			vocab("timetable", "timetable", "6d08080d-4855-5652-ae35-cf0c2bcb9222", "A list of the times when buses or trains leave."),
			vocab("platform", "platform", "9aa8876d-25c6-54d1-9b65-0816e604e7a5", "The place beside the track where you board a train."),
			vocab("return ticket", "return-ticket", "35b47079-26eb-5441-bf1d-05060d125a7e", "A ticket for a journey there and back."),
		},
		line("Narrator", "Sam wants to visit his cousin in Lakeside and come home that evening."),
		line("Sam", "A return ticket to Lakeside, please."),
		line("Clerk", "Here you are. The next train leaves at eleven twenty."),
		line("Sam", "The old timetable says platform two. Is that still right?"),
		line("Clerk", "No, today's train leaves from platform four. Platform two is closed."),
		question("comprehension", "Where should Sam wait for today's train?", 0, "The clerk says today's train leaves from platform four. The old timetable is no longer correct.", "Platform four.", "Platform two.", "Outside the station."),
		line("Sam", "Thank you. Is platform four far from here?"),
		line("Clerk", "Go past the cafe and turn left. You have fifteen minutes."),
		line("Narrator", "Sam finds platform four and checks the sign before getting on."),
		question("phrase_completion", "Sam plans to come home the same day. He asks for a ___ ticket.", 1, "A return ticket covers the journey to Lakeside and the journey back.", "single", "return", "closed"),
	),
	story("a-clear-deadline", "A clear deadline", "Check a work task and agree when to send it.", "work",
		[]StoryVocabulary{
			vocab("colleague", "colleague", "8ecac8d4-b7c0-51bb-95be-14a73d8637e3", "A person you work with."),
			vocab("deadline", "deadline", "b571a540-89bb-533b-be81-9cedf443a64a", "The latest time or date when work must be finished."),
			vocab("attachment", "attachment", "ee851d91-6476-5942-9454-8f1b2db92682", "A file sent with an email."),
		},
		line("Narrator", "Mina is preparing a price list. She asks her colleague Alex for help."),
		line("Mina", "I have the new prices, but two product names are missing."),
		line("Alex", "I can check those names after our morning meeting."),
		line("Mina", "Great. The deadline is four this afternoon."),
		question("comprehension", "What does Mina need Alex to check?", 2, "Mina says two product names are missing. Alex offers to check those names.", "The time of tomorrow's meeting.", "The prices of every product.", "Two missing product names."),
		line("Alex", "I'll send the names before lunch. When will you email the list?"),
		line("Mina", "At three, so I have an hour to fix any problems."),
		line("Alex", "Please send the list as an attachment. The manager wants a file."),
		line("Narrator", "Mina adds the names and sends the file at three."),
		question("phrase_completion", "Complete Alex's request: Please send the list as an ___.", 0, "An attachment is a file sent with an email, which is what the manager needs.", "attachment", "appointment", "appetizer"),
	),
	story("a-better-fit", "A better fit", "Exchange a jacket and check the new size.", "shopping",
		[]StoryVocabulary{
			vocab("receipt", "receipt", "74a55fe5-7736-58f0-98b8-23eef998067d", "A paper or digital record showing what you paid."),
			vocab("exchange", "exchange", "c916143f-197b-5cd7-a4c9-7b21178f04aa", "To replace a bought item with another item."),
			vocab("try on", "try-on", "481f699b-4788-5620-b8a5-51347d90e6e5", "To put on clothes to see how they look or fit."),
		},
		line("Narrator", "Leo bought a blue jacket yesterday. At home, he noticed the sleeves were too short."),
		line("Leo", "I'd like to exchange this jacket for a larger size, please."),
		line("Assistant", "Do you have your receipt?"),
		line("Leo", "Yes, here it is. I don't need a refund; I like the jacket."),
		question("comprehension", "What does Leo want to do?", 1, "Leo asks for a larger size and says he does not need a refund.", "Get his money back.", "Get the same jacket in a larger size.", "Buy a jacket in a different colour."),
		line("Assistant", "We have the same blue jacket in a large. You can try it on over there."),
		line("Leo", "This one feels better. The sleeves reach my wrists."),
		line("Assistant", "Good. Both sizes cost the same, so there is nothing extra to pay."),
		line("Narrator", "Leo takes the larger jacket and keeps the new receipt."),
		question("phrase_completion", "Before choosing a size, Leo can ___ the jacket.", 2, "To try on a jacket means to put it on and check the fit.", "call back", "check out", "try on"),
	),
	story("plans-for-saturday", "Plans for Saturday", "Make plans with a friend when the weather changes.", "friends",
		[]StoryVocabulary{
			vocab("available", "available", "b2d5132c-b969-587f-975d-4009c454968f", "Free to do something at a particular time."),
			vocab("reschedule", "reschedule", "0cdfa962-ee1c-570d-8533-999088905805", "To arrange an activity for a different time."),
			vocab("meet up", "meet-up", "c5454676-5349-52d2-ad14-814d516fed56", "To meet someone, usually for a social activity."),
		},
		line("Narrator", "Ella and Ravi planned a picnic for Saturday morning."),
		line("Ella", "The weather report says it will rain all morning."),
		line("Ravi", "Let's reschedule the picnic for Sunday. Are you available then?"),
		line("Ella", "I'm working on Sunday, but I'm free on Saturday afternoon."),
		question("comprehension", "When is Ella free?", 0, "Ella says she is free on Saturday afternoon and working on Sunday.", "Saturday afternoon.", "Sunday morning.", "Every day this weekend."),
		line("Ravi", "We could meet up at the museum on Saturday afternoon instead."),
		line("Ella", "Sounds good. The museum is indoors. Shall we meet at two?"),
		line("Ravi", "Yes, at the main entrance. We can have a picnic another weekend."),
		line("Narrator", "They keep their Saturday plans and choose an activity away from the rain."),
		question("phrase_completion", "Ravi suggests: We could ___ at the museum.", 1, "Meet up means to meet someone for a social activity, like their museum visit.", "pay back", "meet up", "try on"),
	),
	story("a-room-for-tonight", "A room for tonight", "Check in and ask about breakfast and check-out.", "hotel",
		[]StoryVocabulary{
			vocab("reservation", "reservation", "edb20c6e-1023-5f10-af4b-0dc059a781b2", "An arrangement to keep a room for you."),
			vocab("key card", "key-card", "d990ce6e-4756-52e0-96e8-9926442963f2", "A plastic card used to open a hotel room."),
			vocab("check-out", "check-out", "846f3dde-4b90-5eb8-929e-ecd28e8315f3", "The process of leaving a hotel and returning the room key."),
		},
		line("Narrator", "Aisha arrives at a hotel after a long bus journey."),
		line("Aisha", "I have a reservation for one night under Aisha Malik."),
		line("Receptionist", "Welcome. Your room is on the second floor. Here is your key card."),
		line("Aisha", "Thank you. What time is breakfast?"),
		line("Receptionist", "Breakfast is from seven to nine. Check-out is at eleven."),
		question("comprehension", "What is the latest check-out time?", 2, "The receptionist says check-out is at eleven. Seven to nine are breakfast times.", "Seven in the morning.", "Nine in the morning.", "Eleven in the morning."),
		line("Aisha", "My bus leaves at one. Can you keep my bag after check-out?"),
		line("Receptionist", "Yes, bring it to the front desk. We will keep it until you leave."),
		line("Narrator", "Aisha can enjoy breakfast and explore the town before her bus journey."),
		question("phrase_completion", "Complete the receptionist's instruction: Here is your ___ to open the room.", 0, "The receptionist gives Aisha a key card, which opens her hotel room.", "key card", "return ticket", "receipt"),
	),
}

func build(key string) (Snapshot, error) {
	for _, snap := range catalog {
		if snap.Story.Key == key {
			// Copy all nested content so callers cannot mutate the authoring catalog.
			raw, _ := json.Marshal(snap)
			var copy Snapshot
			if err := json.Unmarshal(raw, &copy); err != nil {
				return Snapshot{}, err
			}
			return copy, nil
		}
	}
	return Snapshot{}, ErrNotFound
}
