package handlers

import (
	"kori/internal/models"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Link scanners and security gateways that name themselves in the user agent.
// Gateways that pose as a browser (Safe Links, Proofpoint URL Defense) are caught by timing instead.
var scannerAgents = []string{
	"barracuda", "proofpoint", "mimecast", "messagelabs", "symantec", "trendmicro", "trend micro",
	"fortiguard", "sophos", "ironport", "zscaler", "forcepoint", "bitdefender", "safelinks",
	"bot/", "bot;", "crawler", "spider", "headlesschrome", "phantomjs",
	"python-requests", "python-urllib", "go-http-client", "curl/", "wget/", "okhttp", "java/", "apache-httpclient", "libwww-perl",
}

const (
	// A person cannot read an email and click in it this soon after it was sent.
	clickAfterDelivery = 2 * time.Second
	// Scanners follow every link at once; nobody clicks two different links within a second.
	linkBurstWindow = time.Second
)

// machineReason classifies a tracking request from the request alone.
// It returns an empty kind for traffic that looks like a person.
func machineReason(method, agent string, event models.EmailTrackingEvent) (kind, reason string) {
	ua := strings.ToLower(strings.TrimSpace(agent))
	if method == http.MethodHead {
		return models.TrackingMachineBot, "HEAD request"
	}
	if ua == "" {
		return models.TrackingMachineBot, "no user agent"
	}
	for _, s := range scannerAgents {
		if strings.Contains(ua, s) {
			return models.TrackingMachineBot, "link scanner (" + strings.Trim(s, "/;") + ")"
		}
	}
	if event == models.EmailTrackingEventOpen {
		if strings.Contains(ua, "googleimageproxy") {
			return models.TrackingMachineProxy, "Google image proxy"
		}
		// Apple Mail Privacy Protection fetches every image with this bare agent.
		if ua == "mozilla/5.0" {
			return models.TrackingMachineProxy, "Apple Mail Privacy Protection"
		}
	}
	return "", ""
}

// flagClickBurst runs after a click is saved. When a different link in the same
// email was clicked within linkBurstWindow, every human click in that window is
// a scanner's. Checking after the insert catches scanners that fire in parallel.
func flagClickBurst(db *gorm.DB, t *models.EmailTracking) error {
	since := t.Timestamp.Add(-linkBurstWindow)
	var others int64
	if err := db.Model(&models.EmailTracking{}).Where("email_id = ? AND event = ? AND url <> ? AND timestamp >= ?", t.EmailID, models.EmailTrackingEventClick, t.URL, since).Count(&others).Error; err != nil || others == 0 {
		return err
	}
	const reason = "several links clicked at once"
	if t.Machine == "" {
		t.Machine, t.MachineReason = models.TrackingMachineBot, reason
	}
	return db.Model(&models.EmailTracking{}).Where("email_id = ? AND event = ? AND timestamp >= ? AND machine = ''", t.EmailID, models.EmailTrackingEventClick, since).
		Updates(map[string]interface{}{"machine": models.TrackingMachineBot, "machine_reason": reason}).Error
}
