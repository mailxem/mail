package services

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func TestS3ObjectACL(t *testing.T) {
	for _, test := range []struct {
		name       string
		provider   string
		disableACL string
		wantACL    string
	}{
		{name: "existing S3 behavior", provider: "s3", wantACL: "authenticated-read"},
		{name: "existing R2 behavior", provider: "r2", wantACL: "public-read"},
		{name: "private S3 bucket", provider: "s3", disableACL: "true"},
		{name: "private R2 bucket", provider: "r2", disableACL: "true"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var mu sync.Mutex
			var uploaded string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				if r.Method == http.MethodPut {
					if acl := r.Header.Get("X-Amz-Acl"); acl != test.wantACL {
						t.Errorf("upload ACL = %q; want %q", acl, test.wantACL)
					}
					uploaded = r.URL.Path
					body, err := io.ReadAll(io.LimitReader(r.Body, 1024))
					if err != nil || string(body) != "private fixture" {
						t.Error("upload body changed", err)
					}
					w.Header().Set("ETag", `"fixture"`)
					return
				}
				if r.Method != http.MethodGet || r.URL.Path != "/xem" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/xml")
				_, _ = w.Write([]byte(`<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Name>xem</Name></ListBucketResult>`))
			}))
			defer server.Close()
			t.Setenv("S3_ENDPOINT_URL", server.URL)
			t.Setenv("S3_PUBLIC_ENDPOINT_URL", "")
			t.Setenv("S3_CREATE_BUCKET", "false")
			t.Setenv("S3_DISABLE_ACL", test.disableACL)
			t.Setenv("STORAGE_PROVIDER", test.provider)
			svc, err := NewS3Service("xem", "", "us-east-1", "fixture", "fixture-secret")
			if err != nil {
				t.Fatal(err)
			}
			objectURL, err := svc.UploadFile(context.Background(), []byte("private fixture"), "fixture.txt", types.ObjectCannedACLAuthenticatedRead, "text/plain")
			if err != nil {
				t.Fatal(err)
			}
			mu.Lock()
			objectPath := uploaded
			mu.Unlock()
			if objectPath == "" || objectURL != server.URL+objectPath {
				t.Fatal("upload did not preserve the configured bucket origin")
			}
			key := strings.TrimPrefix(objectPath, "/xem/")
			signed, err := svc.GetSignedURL(context.Background(), key, time.Minute)
			if err != nil || !strings.HasPrefix(signed, server.URL+objectPath+"?") || !strings.Contains(signed, "X-Amz-Signature=") {
				t.Fatal("file did not retain its presigned read URL", err)
			}
		})
	}
}
