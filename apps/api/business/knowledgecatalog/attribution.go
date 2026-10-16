package knowledgecatalog

import _ "embed"

//go:embed LICENSE.wordnet
var wordNetNotice string

type Attribution struct {
	Title          string `json:"title"`
	SourceURL      string `json:"sourceUrl"`
	LicenseURL     string `json:"licenseUrl"`
	WordNetLicense string `json:"wordnetLicense"`
}

// SourceAttribution travels with copies of source definitions, including daily
// lessons. It is public data and does not require a database or external call.
func SourceAttribution() Attribution {
	return Attribution{
		Title:     "Word order: Hermit Dave / FrequencyWords, OpenSubtitles 2018 (CC BY-SA 4.0). Definitions: WordNet 3.0 and Vocanova editorial core. Filtered and adapted by Vocanova.",
		SourceURL: "https://github.com/hermitdave/FrequencyWords", LicenseURL: "https://creativecommons.org/licenses/by-sa/4.0/", WordNetLicense: wordNetNotice,
	}
}
