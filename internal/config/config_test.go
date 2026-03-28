package config

import "testing"

func TestParseAllowedOriginsUsesDefaults(t *testing.T) {
	origins := ParseAllowedOrigins("")

	if len(origins) != 2 {
		t.Fatalf("expected 2 default origins, got %d", len(origins))
	}

	if origins[0] != "http://localhost:8080" {
		t.Fatalf("unexpected first origin: %q", origins[0])
	}
}
