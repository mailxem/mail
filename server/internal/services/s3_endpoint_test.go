package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExplicitS3Endpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/xem" {
			t.Errorf("expected path-style bucket, got %s", r.URL.Path)
		}
		if !strings.Contains(r.Header.Get("Authorization"), "/us-east-1/s3/aws4_request") {
			t.Error("incorrect signing region")
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Name>xem</Name></ListBucketResult>`))
	}))
	defer server.Close()
	t.Setenv("S3_ENDPOINT_URL", server.URL)
	t.Setenv("S3_PUBLIC_ENDPOINT_URL", "")
	t.Setenv("S3_CREATE_BUCKET", "false")
	svc, err := NewS3Service("xem", "", "us-east-1", "test", "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	signed, err := svc.GetSignedURL(context.Background(), "file.png", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(signed, server.URL+"/xem/file.png?") {
		t.Fatalf("unexpected signed URL: %s", signed)
	}
}

func TestResolveS3Endpoint(t *testing.T) {
	endpoint, region, pathStyle, err := resolveS3Endpoint("example.com", "apac", "")
	if err != nil || endpoint != "https://apac.example.com" || region != "apac" || pathStyle {
		t.Fatal("legacy endpoint changed")
	}
	for _, invalid := range []string{"ftp://host", "http://user:pass@host", "http://host/path", "http://host?key=value", "http://host#fragment", "://"} {
		if _, _, _, err := resolveS3Endpoint("", "us-east-1", invalid); err == nil {
			t.Errorf("accepted %s", invalid)
		}
	}
	if _, _, _, err := resolveS3Endpoint("", "", "http://storage:9000"); err == nil {
		t.Error("accepted missing signing region")
	}
}
