export const NATIVE_LANGUAGES = [
  { code: "fa", label: "Persian", native: "فارسی", aliases: ["Farsi"] },
  { code: "en", label: "English", native: "English" },
  { code: "ar", label: "Arabic", native: "العربية" },
  { code: "es", label: "Spanish", native: "Español" },
  { code: "tr", label: "Turkish", native: "Türkçe" },
  { code: "ur", label: "Urdu", native: "اردو" },
  { code: "fr", label: "French", native: "Français" },
  { code: "de", label: "German", native: "Deutsch" },
  { code: "az", label: "Azerbaijani", native: "Azərbaycanca" },
  { code: "bn", label: "Bengali", native: "বাংলা" },
  { code: "bg", label: "Bulgarian", native: "Български" },
  { code: "zh", label: "Chinese", native: "中文" },
  { code: "cs", label: "Czech", native: "Čeština" },
  { code: "da", label: "Danish", native: "Dansk" },
  { code: "nl", label: "Dutch", native: "Nederlands" },
  { code: "fi", label: "Finnish", native: "Suomi" },
  { code: "el", label: "Greek", native: "Ελληνικά" },
  { code: "he", label: "Hebrew", native: "עברית" },
  { code: "hi", label: "Hindi", native: "हिन्दी" },
  { code: "hu", label: "Hungarian", native: "Magyar" },
  { code: "id", label: "Indonesian", native: "Bahasa Indonesia" },
  { code: "it", label: "Italian", native: "Italiano" },
  { code: "ja", label: "Japanese", native: "日本語" },
  { code: "ko", label: "Korean", native: "한국어" },
  { code: "ku", label: "Kurdish", native: "Kurdî" },
  { code: "ms", label: "Malay", native: "Bahasa Melayu" },
  { code: "ne", label: "Nepali", native: "नेपाली" },
  { code: "no", label: "Norwegian", native: "Norsk" },
  { code: "ps", label: "Pashto", native: "پښتو" },
  { code: "pl", label: "Polish", native: "Polski" },
  { code: "pt", label: "Portuguese", native: "Português" },
  { code: "pa", label: "Punjabi", native: "ਪੰਜਾਬੀ" },
  { code: "ro", label: "Romanian", native: "Română" },
  { code: "ru", label: "Russian", native: "Русский" },
  { code: "sv", label: "Swedish", native: "Svenska" },
  { code: "ta", label: "Tamil", native: "தமிழ்" },
  { code: "te", label: "Telugu", native: "తెలుగు" },
  { code: "th", label: "Thai", native: "ไทย" },
  { code: "uk", label: "Ukrainian", native: "Українська" },
  { code: "vi", label: "Vietnamese", native: "Tiếng Việt" },
] as const;

function searchKey(value: string): string {
  return value.trim().normalize("NFKC").toLowerCase();
}

export function findNativeLanguage(value: string) {
  const key = searchKey(value);
  return NATIVE_LANGUAGES.find((language) =>
    [
      language.code,
      language.label,
      language.native,
      ...("aliases" in language ? language.aliases : []),
    ].some((name) => searchKey(name) === key),
  );
}

export function nativeLanguageLabel(value: string): string {
  return findNativeLanguage(value)?.label ?? value.trim();
}

export function nativeLanguageValue(value: string): string {
  return findNativeLanguage(value)?.code ?? value.trim();
}

export function matchingNativeLanguages(value: string) {
  const key = searchKey(value);
  if (!key) return NATIVE_LANGUAGES.slice(0, 8);
  return NATIVE_LANGUAGES.filter((language) =>
    [
      language.code,
      language.label,
      language.native,
      ...("aliases" in language ? language.aliases : []),
    ].some((name) => searchKey(name).includes(key)),
  ).slice(0, 8);
}
