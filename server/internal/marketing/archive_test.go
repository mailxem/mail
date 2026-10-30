package marketing

import "testing"

func TestNormalizePublicSlug(t *testing.T) {
	for raw, want := range map[string]string{" Weekly-Notes ": "weekly-notes", "abc": "abc"} {
		got, err := NormalizePublicSlug(raw)
		if err != nil || got == nil || *got != want {
			t.Fatalf("%q: got %v, %v", raw, got, err)
		}
	}
	if got, err := NormalizePublicSlug(""); got != nil || err != nil {
		t.Fatalf("empty should be private, got %v %v", got, err)
	}
	for _, raw := range []string{"ab", "-abc", "abc-", "a b c", "ünï", "../etc"} {
		if _, err := NormalizePublicSlug(raw); err == nil {
			t.Fatalf("%q should be refused", raw)
		}
	}
}

func TestPublicEditionHidesMergeFields(t *testing.T) {
	got := PublicEditionHTML(`<p>Hi {{first_name}}, from {{ inviter }} at {{company}}.</p>`)
	if got != `<p>Hi, from at.</p>` {
		t.Fatalf("got %q", got)
	}
	if s := PublicEditionSubject("Three bets for Q4, {{first_name}}"); s != "Three bets for Q4," {
		t.Fatalf("got %q", s)
	}
}
