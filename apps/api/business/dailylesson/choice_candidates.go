package dailylesson

// Shared by session construction and the exhaustive seeded-corpus availability
// regression. Keep the DISTINCT ordering and cap identical in both paths.
const candidateDefinitionsSQL = `SELECT DISTINCT ON(lower(trim(wm.short_definition))) wm.short_definition,cw.text FROM word_meanings wm JOIN canonical_words cw ON cw.id=wm.word_id WHERE wm.status='active' AND cw.status='active' AND cw.id<>$1 AND lower(trim(wm.short_definition))<>lower(trim($2)) AND strpos(lower(wm.short_definition),lower($3))=0
 AND EXISTS(SELECT 1 FROM journey_words jw JOIN journey_situations js ON js.id=jw.journey_situation_id WHERE jw.meaning_id=wm.id AND js.status='active')
 AND NOT EXISTS(SELECT 1 FROM knowledge_catalog_entries target JOIN knowledge_catalog_entries candidate ON candidate.source_sense_key=target.source_sense_key WHERE target.meaning_id=$4 AND candidate.meaning_id=wm.id)
 ORDER BY lower(trim(wm.short_definition)),wm.id LIMIT 60`
