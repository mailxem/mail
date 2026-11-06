package workflowconfig

import "testing"

func TestValidateWaitUntil(t *testing.T) {
	for raw, ok := range map[string]bool{
		`{"event":"opened","timeout":"72h"}`:                           true,
		`{"event":"clicked","url":"https://x.test/a","timeout":"1h"}`:  true,
		`{"event":"tag_added","tag":"VIP","timeout":"24h"}`:            true,
		`{"event":"field_changed","field":"company","timeout":"24h"}`:  true,
		`{"event":"app_event","name":"trial_started","timeout":"24h"}`: true,
		`{"event":"opened"}`:                                       false,
		`{"event":"opened","timeout":"30s"}`:                       false,
		`{"event":"opened","timeout":"9000h"}`:                     false,
		`{"event":"tag_added","timeout":"1h"}`:                     false,
		`{"event":"field_changed","field":"email","timeout":"1h"}`: false,
		`{"event":"app_event","timeout":"1h"}`:                     false,
		`{"event":"replied","timeout":"1h"}`:                       false,
	} {
		if err := Validate("WAIT_UNTIL", []byte(raw)); (err == nil) != ok {
			t.Errorf("%s: got %v", raw, err)
		}
	}
}
