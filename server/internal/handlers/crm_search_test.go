package handlers

import "testing"

func TestContactSearchPattern(t *testing.T) {
	for in, want := range map[string]string{
		"@Example.com":      "%@example.com%",
		"*@example.com":     "%%@example.com%",
		"jane*@example.com": "%jane%@example.com%",
		"50%_off!":          "%50!%!_off!!%",
	} {
		if got := contactSearchPattern(in); got != want {
			t.Errorf("contactSearchPattern(%q) = %q, want %q", in, got, want)
		}
	}
}
