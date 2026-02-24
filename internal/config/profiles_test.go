package config

import "testing"

func TestSuggestProfileNames(t *testing.T) {
	t.Parallel()

	cfg := &Config{Profiles: map[string]Profile{
		"api-ci":    {},
		"domain-ci": {},
		"smoke":     {},
	}}

	suggestions := cfg.SuggestProfileNames("api-c", 3)
	if len(suggestions) == 0 {
		t.Fatal("expected suggestions")
	}
	if suggestions[0] != "api-ci" {
		t.Fatalf("expected first suggestion to be api-ci, got: %+v", suggestions)
	}
}
