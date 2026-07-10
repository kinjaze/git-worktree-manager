package config

import "testing"

func TestNormalizeLanguage(t *testing.T) {
	if NormalizeLanguage("") != "en" {
		t.Fatal("empty language should default to en")
	}
	if NormalizeLanguage("zh") != "zh" {
		t.Fatal("zh should be preserved")
	}
	if NormalizeLanguage("fr") != "en" {
		t.Fatal("unsupported language should fall back to en")
	}
}
