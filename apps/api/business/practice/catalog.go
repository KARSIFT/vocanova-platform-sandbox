package practice

// This original practice catalog pins the 90 reviewed starter-course meanings.
// Additions/answer changes require a content/grading version bump; active sessions
// retain their original private snapshot. Lesson progress is never changed.
type reference struct{ LessonKey, MeaningID, WordText string }

var catalog = []reference{
	{"conversation-start", "ac93d068-7c3e-55c9-9d85-3fd3518592dc", "greeting"},
	{"conversation-start", "3265c02c-5751-5465-b94b-e767aa871d8d", "small talk"},
	{"conversation-start", "0d267f75-f227-5478-962d-b4b944e0e93e", "casual"},
	{"conversation-find-time", "c99075bf-94cb-5f51-90d5-f244f66ea942", "weekend plans"},
	{"conversation-find-time", "b2d5132c-b969-587f-975d-4009c454968f", "available"},
	{"conversation-find-time", "52b8ebc0-ba65-579d-8749-6ff51b9ce6b5", "sounds good"},
	{"conversation-suggest-activity", "d602623a-1bf7-5656-b867-484d2be3f5a9", "suggest"},
	{"conversation-suggest-activity", "0c030a86-51fc-5d13-9246-09950c45a43c", "join"},
	{"conversation-suggest-activity", "40ed986a-7261-54eb-a3c8-4213c3bdc4fd", "arrange"},
	{"daily-conversation", "ee53d6ba-4303-5394-b7f9-79937ce66d09", "invite"},
	{"daily-conversation", "6348aad5-937a-57d0-aebc-15887a771938", "confirm"},
	{"daily-conversation", "0cdfa962-ee1c-570d-8533-999088905805", "reschedule"},
	{"conversation-plan-changes", "f98bfb7b-c140-55a6-b76c-774ee096d1ca", "cancel"},
	{"conversation-plan-changes", "b535cf81-3777-5553-959d-1ccf9d22a209", "on time"},
	{"conversation-plan-changes", "c5454676-5349-52d2-ad14-814d516fed56", "meet up"},
	{"conversation-stay-connected", "3d64c3c9-ede0-5ffd-b1ef-278f6b70e486", "catch up"},
	{"conversation-stay-connected", "ddde743f-15ab-5d13-bea6-9835a56c991d", "keep in touch"},
	{"conversation-stay-connected", "c954cdb6-c0bc-5d61-9045-18db360f4b84", "farewell"},
	{"restaurant", "329c4ec9-6562-568e-9e71-9134da2f0d8d", "menu"},
	{"restaurant", "2534b539-a80f-5001-b9a4-27a5e0679539", "appetizer"},
	{"restaurant", "5d21c02b-ef51-5545-bd8a-01f5814d3a67", "bill"},
	{"restaurant-other-plans", "52c8ed28-544d-5a80-a911-6ddd03953606", "reservation"},
	{"restaurant-other-plans", "1e9e55f8-d329-5c52-bc30-7ed77a31a167", "take-out"},
	{"restaurant-other-plans", "bd5eacb6-be32-5e7b-a8d7-d56db1908347", "tip"},
	{"shopping-clothes", "7dee60f4-25c4-5d55-9111-dc5174dcd5b5", "size"},
	{"shopping-clothes", "481f699b-4788-5620-b8a5-51347d90e6e5", "try on"},
	{"shopping-clothes", "6c80ff44-eb45-5332-b48d-228872ead757", "fit"},
	{"shopping-returns", "74a55fe5-7736-58f0-98b8-23eef998067d", "receipt"},
	{"shopping-returns", "538cc7f5-983c-59cc-8664-173cefca65ad", "refund"},
	{"shopping-returns", "c916143f-197b-5cd7-a4c9-7b21178f04aa", "exchange"},
	{"home-rental", "e507b507-5432-51a0-bcc3-be12294143db", "rent"},
	{"home-rental", "75be2229-a161-5e08-856f-143098221439", "landlord"},
	{"home-rental", "3c382279-cf27-5d6b-ba75-a0389f1d5680", "repair"},
	{"transport-station", "bf1bee5c-b8f0-5ebc-9252-53c4ac9ef7de", "bus stop"},
	{"transport-station", "6d08080d-4855-5652-ae35-cf0c2bcb9222", "timetable"},
	{"transport-station", "9aa8876d-25c6-54d1-9b65-0816e604e7a5", "platform"},
	{"transport-tickets", "faf13c57-a106-5406-a2bc-e743eb00ec4e", "single ticket"},
	{"transport-tickets", "35b47079-26eb-5441-bf1d-05060d125a7e", "return ticket"},
	{"transport-tickets", "91888c86-1fc6-5633-a6dd-a46017f8fdf7", "change"},
	{"airport", "8a6ec801-3292-5fc2-bfb3-242ac405d891", "boarding pass"},
	{"airport", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", "gate"},
	{"airport", "bee2408e-5991-5805-923c-6fba33ab7e7a", "luggage"},
	{"airport-checks", "2a610b3d-4031-5eee-a0be-d3def2cc0b53", "security check"},
	{"airport-checks", "d62f1f3f-8382-5999-9e9c-30858f6b2c5b", "customs"},
	{"airport-checks", "8f12f8e0-d0a3-5010-941b-8f7d526121cc", "layover"},
	{"hotel-check-in", "edb20c6e-1023-5f10-af4b-0dc059a781b2", "reservation"},
	{"hotel-check-in", "3008733d-9cab-547c-9712-4014a26ca697", "front desk"},
	{"hotel-check-in", "d990ce6e-4756-52e0-96e8-9926442963f2", "key card"},
	{"hotel-services", "dfe9c593-cf2e-505b-96ca-4afef46422d8", "amenities"},
	{"hotel-services", "de015c5a-0fad-507d-944e-572f6a10da66", "wake-up call"},
	{"hotel-services", "846f3dde-4b90-5eb8-929e-ecd28e8315f3", "check-out"},
	{"delivery-parcel", "0b593408-6679-5fb8-a21c-b73a525c76e5", "parcel"},
	{"delivery-parcel", "3bb8e710-2cfb-5f6f-a281-c0a167d46fb2", "address"},
	{"delivery-parcel", "4609031d-af20-550a-906e-3fd3cde7588e", "delivery"},
	{"money-payment", "b20c1447-0e63-5712-9ae7-dd4aca9aeb85", "cash"},
	{"money-payment", "d78438f3-30bd-58bf-9b96-1489c4f57d3b", "bank card"},
	{"money-payment", "c6a6c8f7-1ce0-59df-b963-e07066796992", "payment"},
	{"services-form", "b548ade3-8682-521e-bf26-848526519854", "form"},
	{"services-form", "c7bbc8ae-3f64-58e9-acf0-3069e317c471", "signature"},
	{"services-form", "64360906-46e2-588e-96d7-c78213875c9b", "queue"},
	{"health-services", "c76cb86d-be5a-5375-8d88-72bda7b98e7d", "appointment"},
	{"health-services", "096a46ea-93b8-59c2-9423-bc07f3332297", "symptom"},
	{"health-services", "c3a26504-d404-5151-bbac-5e22511aae5f", "pharmacy"},
	{"phone-calls", "5966f508-b502-5669-a7d6-c619ce92639f", "signal"},
	{"phone-calls", "2fd7a730-c9e9-5a7b-b094-dea868bc7cd3", "voicemail"},
	{"phone-calls", "e1b649b8-b6d8-5caf-8671-5b9e57ebd71c", "call back"},
	{"digital-email", "ee851d91-6476-5942-9454-8f1b2db92682", "attachment"},
	{"digital-email", "ae343b98-d972-5565-8bac-87f04fe96d06", "link"},
	{"digital-email", "2d1adc4c-058b-5d12-ae17-86afc4195588", "reply"},
	{"work-day", "f97ad45b-01ee-51f9-b77a-ddab1881b374", "shift"},
	{"work-day", "ea64a58a-2521-50fa-8af9-a5cf9e013763", "break"},
	{"work-day", "8ecac8d4-b7c0-51bb-95be-14a73d8637e3", "colleague"},
	{"job-interview", "e5817d4f-eaf9-577e-881f-b100fe1e24f8", "resume"},
	{"job-interview", "9ab971f4-9b7b-5551-b274-62ad41d9273f", "references"},
	{"job-interview", "0b9878b3-64a7-5175-92f7-ce0435cfbed4", "salary expectations"},
	{"interview-application", "7ee5fc1b-8108-5fb4-a254-f75af84695f3", "cover letter"},
	{"interview-application", "c7796a82-cc8d-56de-a97f-773ea694e337", "qualifications"},
	{"interview-application", "4c3724d8-5abf-5159-8171-76ccdcdae7c0", "follow-up"},
	{"work-meeting", "8a7f86d3-ed7a-512c-9c19-d86b5d46ffa8", "agenda"},
	{"work-meeting", "71c851c2-c028-5d8f-9e0a-5be524410a82", "action items"},
	{"work-meeting", "b571a540-89bb-533b-be81-9cedf443a64a", "deadline"},
	{"meeting-collaboration", "1eee9c5f-7079-54d4-bcd8-183739df837a", "brainstorm"},
	{"meeting-collaboration", "1e91dacc-7700-5f65-9310-312ac4dff706", "stakeholder"},
	{"meeting-collaboration", "e62a2ba4-c6b7-5186-98f6-20d69f4dc196", "update"},
	{"university-class", "fd6c39c0-f097-5a5b-b947-c93096e635d0", "lecture"},
	{"university-class", "ae72824e-1341-5081-84ea-0e0fdbf09c50", "assignment"},
	{"university-class", "58529fdd-9425-5aef-a8c7-bc1086be7080", "syllabus"},
	{"study-collaboration", "5b7f44dd-77ba-5d20-9ee4-01d6967d203a", "group project"},
	{"study-collaboration", "ce20a0cc-6a7e-5b5e-bbfe-5ec49f1c58d5", "office hours"},
	{"study-collaboration", "c1e69c4e-2915-5c99-96ab-d946e67653fd", "presentation"},
}

func validLesson(k string) bool {
	if k == "" {
		return true
	}
	for _, r := range catalog {
		if r.LessonKey == k {
			return true
		}
	}
	return false
}
func referenceFor(id string) (reference, bool) {
	for _, r := range catalog {
		if r.MeaningID == id {
			return r, true
		}
	}
	return reference{}, false
}

func acceptedAnswers(word string) []string {
	switch word {
	case "appetizer":
		return []string{"appetizer", "appetiser"}
	case "key card":
		return []string{"key card", "keycard", "key-card"}
	case "resume":
		return []string{"resume", "résumé", "resumé", "re\u0301sume\u0301", "resume\u0301"}
	case "reschedule":
		return []string{"reschedule", "re-schedule"}
	case "take-out":
		return []string{"take-out", "takeout"}
	case "check-out":
		return []string{"check-out", "checkout"}
	case "follow-up":
		return []string{"follow-up", "followup"}
	case "voicemail":
		return []string{"voicemail", "voice mail"}
	default:
		return []string{word}
	}
}
func recallDefinition(w Word) string {
	switch w.WordText {
	case "boarding pass":
		return "At the airport, a staff member scans this document, which shows your flight and seat, before you get on the plane."
	case "bill":
		return "At a restaurant, the statement showing the total you need to pay for your meal."
	case "references":
		return "In a job application, people an employer can contact to ask about your previous work."
	case "lecture":
		return "At university, a class where an instructor talks about a subject while students listen and take notes."
	case "syllabus":
		return "A course document listing its topics, schedule, and requirements."
	default:
		return w.Definition
	}
}

// SupportsMeaning reports whether a canonical meaning has reviewed practice content.
func SupportsMeaning(id, text, definition string) bool {
	ref, ok := referenceFor(id)
	return ok && ref.WordText == text && definition != ""
}
