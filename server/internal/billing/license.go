package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// ErrLicenseFeature is shown when an unlicensed self-hosted install sets up a
// feature that needs a paid plan on Xem Cloud.
const ErrLicenseFeature = "Newsletters, Cloudflare mailboxes and the Xem inbox need a Xem license on self-hosted installs: buy one for $9 at xem.email/pricing and set XEM_LICENSE_KEY"

// LicenseURL is Dodo Payments' public license check; it needs no API key.
var LicenseURL = "https://live.dodopayments.com/licenses/validate"

var (
	licenseRequired atomic.Bool // set by WatchLicense: a self-hosted server
	licensed        atomic.Bool
)

// unlicensed reports whether this is a self-hosted install without a valid
// license, which gets the Free plan's paid-feature limits and email badge.
func unlicensed() bool { return licenseRequired.Load() && !licensed.Load() }

// Licensed reports whether a self-hosted install has a valid license.
func Licensed() bool { return licensed.Load() }

// WatchLicense checks a self-hosted install's license key now and daily until
// ctx ends. A failed check keeps the last answer, so a Dodo outage never takes
// a licensed install's features away.
// ponytail: an install that has never reached Dodo stays unlicensed; add a
// signed offline key if air-gapped installs ask for one.
func WatchLicense(ctx context.Context, key string, logf func(string, ...any)) {
	licenseRequired.Store(true)
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	for {
		valid, err := validateLicense(ctx, key)
		switch {
		case err != nil:
			logf("Xem license check failed, keeping the last result: %v", err)
		case valid != licensed.Load():
			licensed.Store(valid)
			logf("Xem license valid: %v", valid)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(24 * time.Hour):
		}
	}
}

func validateLicense(ctx context.Context, key string) (bool, error) {
	body, _ := json.Marshal(map[string]string{"license_key": key})
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, LicenseURL, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	// Dodo answers 404 or 422 for keys it does not know.
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusUnprocessableEntity {
		return false, nil
	}
	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("license check returned %s", res.Status)
	}
	var out struct {
		Valid bool `json:"valid"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return false, err
	}
	return out.Valid, nil
}
