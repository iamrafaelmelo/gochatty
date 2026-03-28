package config

import (
	"testing"
)

func TestParseAllowedOriginsReturnsEmptyWhenUnset(t *testing.T) {
	origins := ParseAllowedOrigins("")

	if len(origins) != 0 {
		t.Fatalf("expected no default origins, got %d", len(origins))
	}
}

func TestParseAllowedOriginsTrimsAndSkipsEmptyValues(t *testing.T) {
	origins := ParseAllowedOrigins(" http://localhost:8080, ,http://127.0.0.1:8080 ,, ")

	expected := []string{"http://localhost:8080", "http://127.0.0.1:8080"}
	if len(origins) != len(expected) {
		t.Fatalf("expected %d origins, got %d (%#v)", len(expected), len(origins), origins)
	}

	for i, origin := range expected {
		if origins[i] != origin {
			t.Fatalf("expected origin %d to be %q, got %q", i, origin, origins[i])
		}
	}
}

func TestStringFromEnvUsesFallbackForBlankValues(t *testing.T) {
	t.Setenv("APP_PORT", "   ")

	value := stringFromEnv("APP_PORT", DefaultPort)

	if value != DefaultPort {
		t.Fatalf("expected fallback %q, got %q", DefaultPort, value)
	}
}
