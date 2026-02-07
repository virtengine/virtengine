package ocr

import (
	"unicode"
)

func DetectLanguages(text string) []Language {
	seen := map[Language]struct{}{}

	for _, char := range text {
		switch {
		case unicode.Is(unicode.Cyrillic, char):
			seen[LanguageCyrillic] = struct{}{}
		case unicode.Is(unicode.Arabic, char):
			seen[LanguageArabic] = struct{}{}
		case unicode.Is(unicode.Han, char) || unicode.Is(unicode.Hiragana, char) || unicode.Is(unicode.Katakana, char) || unicode.Is(unicode.Hangul, char):
			seen[LanguageCJK] = struct{}{}
		case unicode.Is(unicode.Devanagari, char):
			seen[LanguageDevanagari] = struct{}{}
		case unicode.Is(unicode.Thai, char):
			seen[LanguageThai] = struct{}{}
		case unicode.IsLetter(char):
			seen[LanguageLatin] = struct{}{}
		}
	}

	if len(seen) == 0 {
		return []Language{LanguageLatin}
	}

	languages := make([]Language, 0, len(seen))
	for lang := range seen {
		languages = append(languages, lang)
	}
	return languages
}
