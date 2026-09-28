package utils

import (
	"encoding/base64"
	"net/http"
	"strings"
)

// 🖼️ TransparentGIF returns a 1x1 transparent GIF
func TransparentGIF() []byte {
	// This is a base64 encoded 1x1 transparent GIF
	const transparentPixel = "R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
	decoded, _ := base64.StdEncoding.DecodeString(transparentPixel)
	return decoded
}

// 🌐 GetIPAddress gets the real IP address from request
func GetIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// 🌍 GeoData represents geolocation information
type GeoData struct {
	Country string
	City    string
	Region  string
}

// 🌍 GetGeolocationData gets location data from IP address
// You would implement this using your preferred geolocation service
// For example: MaxMind GeoIP2, IP-API, etc.
func GetGeolocationData(ipAddress string) (*GeoData, error) {
	// TODO: Implement actual geolocation lookup
	// For now return placeholder data
	return &GeoData{
		Country: "Unknown",
		City:    "Unknown",
		Region:  "Unknown",
	}, nil
}
