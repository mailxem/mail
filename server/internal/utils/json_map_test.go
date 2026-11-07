package utils

import "testing"

func TestJSONToMapTreatsEmptyDataAsNoValues(t *testing.T) {
	for _, in := range []string{"", "null", "{}"} {
		m, err := JSONToMap([]byte(in))
		if err != nil || len(m) != 0 {
			t.Fatalf("JSONToMap(%q) = %v, %v", in, m, err)
		}
	}
	m, err := JSONToMap([]byte(`{"first_name":"Ada"}`))
	if err != nil || m["first_name"] != "Ada" {
		t.Fatalf("got %v, %v", m, err)
	}
}
