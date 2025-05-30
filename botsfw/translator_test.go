package botsfw

import (
	"testing"

	"github.com/strongo/i18n"
)

// Ensure we're implementing the correct interface
var _ i18n.Translator = (*mockTranslator)(nil)

// mockTranslator is a mock implementation of i18n.Translator for testing
type mockTranslator struct {
	translations map[string]map[string]string
}

func (m mockTranslator) Translate(key, locale string, args ...interface{}) string {
	if translations, ok := m.translations[locale]; ok {
		if translation, ok := translations[key]; ok {
			return translation
		}
	}
	return key
}

func (m mockTranslator) TranslateNoWarning(key, locale string, args ...interface{}) string {
	return m.Translate(key, locale, args...)
}

func (m mockTranslator) TranslateWithMap(key, locale string, args map[string]string) string {
	if translations, ok := m.translations[locale]; ok {
		if translation, ok := translations[key]; ok {
			return translation
		}
	}
	return key
}

func TestTranslator_Locale(t *testing.T) {
	// Create a translator with a fixed locale
	tr := translator{
		Translator:  mockTranslator{},
		localeCode5: func() string { return "en-US" },
	}

	// Test that the Locale method returns the correct locale
	locale := tr.Locale()
	if locale.Code5 != "en-US" {
		t.Errorf("Expected locale code 'en-US', got '%s'", locale.Code5)
	}
}

func TestTranslator_Translate(t *testing.T) {
	// Create a mock translator with some translations
	mockTr := mockTranslator{
		translations: map[string]map[string]string{
			"en-US": {
				"hello":   "Hello",
				"welcome": "Welcome, %s",
			},
			"es-ES": {
				"hello":   "Hola",
				"welcome": "Bienvenido, %s",
			},
		},
	}

	tests := []struct {
		name       string
		localeCode string
		key        string
		args       []interface{}
		want       string
	}{
		{
			name:       "English hello",
			localeCode: "en-US",
			key:        "hello",
			args:       nil,
			want:       "Hello",
		},
		{
			name:       "Spanish hello",
			localeCode: "es-ES",
			key:        "hello",
			args:       nil,
			want:       "Hola",
		},
		{
			name:       "Unknown key",
			localeCode: "en-US",
			key:        "unknown",
			args:       nil,
			want:       "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := translator{
				Translator:  mockTr,
				localeCode5: func() string { return tt.localeCode },
			}

			got := tr.Translate(tt.key, tt.args...)
			if got != tt.want {
				t.Errorf("Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranslator_TranslateNoWarning(t *testing.T) {
	// Create a mock translator with some translations
	mockTr := mockTranslator{
		translations: map[string]map[string]string{
			"en-US": {
				"hello": "Hello",
			},
		},
	}

	tr := translator{
		Translator:  mockTr,
		localeCode5: func() string { return "en-US" },
	}

	got := tr.TranslateNoWarning("hello")
	if got != "Hello" {
		t.Errorf("TranslateNoWarning() = %v, want %v", got, "Hello")
	}
}

func TestTranslator_TranslateWithMap(t *testing.T) {
	// Create a mock translator with some translations
	mockTr := mockTranslator{
		translations: map[string]map[string]string{
			"en-US": {
				"greeting": "Hello, {name}",
			},
		},
	}

	tr := translator{
		Translator:  mockTr,
		localeCode5: func() string { return "en-US" },
	}

	args := map[string]string{"name": "John"}
	got := tr.TranslateWithMap("greeting", args)
	if got != "Hello, {name}" {
		t.Errorf("TranslateWithMap() = %v, want %v", got, "Hello, {name}")
	}
}
