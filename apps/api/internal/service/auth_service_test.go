package service

import "testing"

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"RetailPulse AI":     "retailpulse-ai",
		"  ACME & Sons LLC ": "acme-sons-llc",
		"":                   "organization",
	}
	for input, expected := range tests {
		if actual := slugify(input); actual != expected {
			t.Fatalf("slugify(%q) = %q, want %q", input, actual, expected)
		}
	}
}
