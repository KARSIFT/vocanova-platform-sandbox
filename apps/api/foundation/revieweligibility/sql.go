// Package revieweligibility keeps every review queue/count on one server-owned
// rule. SQL fragments are built only from fixed identifiers/placeholders in code.
package revieweligibility

import "fmt"

// ReviewedToday uses the server-created audit, never client answered_at. Local
// midnight boundaries are converted separately so 23/25-hour DST days work.
func ReviewedToday(user, word, now string) string {
	zone := "COALESCE((SELECT timezone FROM user_settings WHERE user_id=" + user + "),'UTC')"
	local := "(" + now + "::timestamptz AT TIME ZONE " + zone + ")::date"
	// Keep the set independent of the outer word. PostgreSQL can hash today's
	// canonical identities once instead of rescanning every audit per due row.
	return fmt.Sprintf(`(%s IN(SELECT day_meaning.word_id FROM review_attempts day_attempt JOIN word_meanings day_meaning ON day_meaning.id=day_attempt.meaning_id WHERE day_attempt.user_id=%s AND day_attempt.created_at >= (%s::timestamp AT TIME ZONE %s) AND day_attempt.created_at < ((%s::timestamp + INTERVAL '1 day') AT TIME ZONE %s)))`, word, user, local, zone, local, zone)
}

// Due selects one deterministic due meaning of each canonical word. Other uses
// remain separate saved meanings and can be selected on later days. Every
// caller binds its authenticated user as $1; this keeps the audit set uncorrelated.
// IMPORTANT: $1 MUST be the authenticated requester, never another parameter.
func Due(alias, now string, activeOnly ...bool) string {
	currentContent, earlierContent := "", ""
	if len(activeOnly) > 0 && activeOnly[0] {
		currentContent = " AND EXISTS(SELECT 1 FROM word_meanings eligible_meaning JOIN canonical_words eligible_word ON eligible_word.id=eligible_meaning.word_id WHERE eligible_meaning.id=" + alias + ".meaning_id AND eligible_meaning.status='active' AND eligible_word.status='active')"
		earlierContent = " AND earlier_meaning.status='active'"
	}
	word := "(SELECT word_id FROM word_meanings WHERE id=" + alias + ".meaning_id)"
	return fmt.Sprintf(`(%[1]s.deleted_at IS NULL AND %[1]s.status IN('new','learning','reviewing') AND (%[1]s.next_review_at IS NULL OR %[1]s.next_review_at<=%[2]s)%[5]s AND NOT EXISTS(SELECT 1 FROM user_word_knowledge k WHERE k.user_id=%[1]s.user_id AND k.meaning_id=%[1]s.meaning_id AND k.self_reported_known) AND NOT %[3]s AND NOT EXISTS(SELECT 1 FROM user_words earlier JOIN word_meanings earlier_meaning ON earlier_meaning.id=earlier.meaning_id WHERE earlier.user_id=%[1]s.user_id AND earlier_meaning.word_id=%[4]s%[6]s AND earlier.deleted_at IS NULL AND earlier.status IN('new','learning','reviewing') AND (earlier.next_review_at IS NULL OR earlier.next_review_at<=%[2]s) AND NOT EXISTS(SELECT 1 FROM user_word_knowledge k WHERE k.user_id=earlier.user_id AND k.meaning_id=earlier.meaning_id AND k.self_reported_known) AND (COALESCE(earlier.next_review_at,'-infinity'::timestamptz),earlier.id)<(COALESCE(%[1]s.next_review_at,'-infinity'::timestamptz),%[1]s.id)))`, alias, now, ReviewedToday("$1", word, now), word, currentContent, earlierContent)
}
