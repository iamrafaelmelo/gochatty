package config

import "testing"

func TestParseAllowedOriginsReturnsEmptyWhenUnset(t *testing.T) {
	origins := ParseAllowedOrigins("")

	if len(origins) != 0 {
		t.Fatalf("expected no default origins, got %d", len(origins))
	}
}
