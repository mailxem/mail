// Package sitedomains connects readers' own domains to newsletter websites
// through Cloudflare for SaaS custom hostnames on the xem.email zone.
package sitedomains

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// ErrDisabled means this server has no Cloudflare credentials for custom domains.
var ErrDisabled = errors.New("custom domains aren't set up on this server")

// Target is the hostname readers point their CNAME at.
func Target() string {
	if t := os.Getenv("SITE_DOMAIN_TARGET"); t != "" {
		return t
	}
	return "sites.xem.email"
}

// Record is a DNS record the domain owner must add.
type Record struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Hostname is a custom hostname's state at Cloudflare.
type Hostname struct {
	ID string `json:"id"`
	// Status is "active" once traffic and the certificate both work.
	Status  string   `json:"status"`
	Records []Record `json:"records"`
}

// Client talks to Cloudflare's custom hostnames API.
type Client struct {
	Token, Zone, BaseURL string
	HTTP                 *http.Client
}

// FromEnv reads CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID; nil when unset.
func FromEnv() *Client {
	token, zone := os.Getenv("CLOUDFLARE_API_TOKEN"), os.Getenv("CLOUDFLARE_ZONE_ID")
	if token == "" || zone == "" {
		return nil
	}
	return &Client{Token: token, Zone: zone, BaseURL: "https://api.cloudflare.com/client/v4", HTTP: &http.Client{Timeout: 20 * time.Second}}
}

var hostPattern = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)

// Normalize lower-cases a domain and rejects anything that isn't a plain
// hostname, or that belongs to Xem itself.
func Normalize(raw string) (string, error) {
	host := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(raw)), ".")
	host = strings.TrimPrefix(strings.TrimPrefix(host, "https://"), "http://")
	if len(host) > 253 || !hostPattern.MatchString(host) {
		return "", errors.New("Enter a domain like news.example.com")
	}
	if host == "xem.email" || strings.HasSuffix(host, ".xem.email") {
		return "", errors.New("Use a domain you own")
	}
	return host, nil
}

type apiHostname struct {
	ID                    string `json:"id"`
	Status                string `json:"status"`
	OwnershipVerification struct {
		Type, Name, Value string
	} `json:"ownership_verification"`
	SSL struct {
		Status            string `json:"status"`
		ValidationRecords []struct {
			TxtName  string `json:"txt_name"`
			TxtValue string `json:"txt_value"`
		} `json:"validation_records"`
	} `json:"ssl"`
}

func (a apiHostname) hostname(host string) Hostname {
	h := Hostname{ID: a.ID, Status: "pending", Records: []Record{{Type: "CNAME", Name: host, Value: Target()}}}
	if a.Status == "active" && a.SSL.Status == "active" {
		h.Status = "active"
	}
	if v := a.OwnershipVerification; v.Name != "" && a.Status != "active" {
		h.Records = append(h.Records, Record{Type: strings.ToUpper(v.Type), Name: v.Name, Value: v.Value})
	}
	for _, r := range a.SSL.ValidationRecords {
		if r.TxtName != "" && a.SSL.Status != "active" {
			h.Records = append(h.Records, Record{Type: "TXT", Name: r.TxtName, Value: r.TxtValue})
		}
	}
	return h
}

func (c *Client) call(ctx context.Context, method, path string, body any, out *apiHostname) error {
	if c == nil {
		return ErrDisabled
	}
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+"/zones/"+c.Zone+"/custom_hostnames"+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var env struct {
		Success bool
		Errors  []struct{ Message string }
		Result  json.RawMessage
	}
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		return fmt.Errorf("cloudflare %s: %w", res.Status, err)
	}
	if !env.Success {
		if len(env.Errors) > 0 {
			return fmt.Errorf("cloudflare: %s", env.Errors[0].Message)
		}
		return fmt.Errorf("cloudflare %s", res.Status)
	}
	if out != nil && len(env.Result) > 0 {
		return json.Unmarshal(env.Result, out)
	}
	return nil
}

// Create starts serving host with a Cloudflare-issued certificate, validated
// over HTTP once the owner's CNAME points at Target.
func (c *Client) Create(ctx context.Context, host string) (Hostname, error) {
	var a apiHostname
	err := c.call(ctx, http.MethodPost, "", map[string]any{
		"hostname": host,
		"ssl":      map[string]any{"method": "http", "type": "dv", "settings": map[string]any{"min_tls_version": "1.2"}},
	}, &a)
	return a.hostname(host), err
}

// Get refreshes a custom hostname's status.
func (c *Client) Get(ctx context.Context, id, host string) (Hostname, error) {
	var a apiHostname
	err := c.call(ctx, http.MethodGet, "/"+id, nil, &a)
	return a.hostname(host), err
}

// Delete stops serving a custom hostname.
func (c *Client) Delete(ctx context.Context, id string) error {
	return c.call(ctx, http.MethodDelete, "/"+id, nil, nil)
}
