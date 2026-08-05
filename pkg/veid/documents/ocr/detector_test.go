package ocr

import (
	"reflect"
	"testing"
)

func TestDetectLanguagesReturnsStableSortedOutput(t *testing.T) {
	want := []Language{LanguageArabic, LanguageCJK, LanguageCyrillic, LanguageLatin}
	for range 100 {
		if got := DetectLanguages("English Русский العربية 日本語"); !reflect.DeepEqual(got, want) {
			t.Fatalf("language detection output is not deterministic: got %v, want %v", got, want)
		}
	}
}
