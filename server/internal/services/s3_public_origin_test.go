package services

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func TestS3PublicObjectOrigin(t *testing.T) {
	t.Run("private upload and public signed read", func(t *testing.T) {
		var mu sync.Mutex
		var uploaded string
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			requests++
			if !strings.Contains(r.Header.Get("Authorization"), "/us-east-1/s3/aws4_request") {
				t.Error("private request lost its configured signing region")
			}
			if r.Header.Get("X-Amz-Acl") != "" {
				t.Error("private request supplied an object ACL")
			}
			if r.Method == http.MethodPut {
				uploaded = r.URL.Path
				body, err := io.ReadAll(io.LimitReader(r.Body, 1024))
				if err != nil || string(body) != "private fixture" {
					t.Error("upload body changed", err)
				}
				w.Header().Set("ETag", `"fixture"`)
				return
			}
			if r.Method != http.MethodGet || r.URL.Path != "/xem-files" {
				t.Errorf("unexpected private request %s %s", r.Method, r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Name>xem-files</Name></ListBucketResult>`))
		}))
		defer server.Close()
		const publicOrigin = "https://xem.example.test:8443"
		t.Setenv("S3_ENDPOINT_URL", server.URL)
		t.Setenv("S3_PUBLIC_ENDPOINT_URL", publicOrigin+"/")
		t.Setenv("STORAGE_PROVIDER", "s3")
		t.Setenv("S3_DISABLE_ACL", "true")
		t.Setenv("S3_CREATE_BUCKET", "false")
		svc, err := NewS3Service("xem-files", "", "us-east-1", "hakopod", "fixture-secret")
		if err != nil {
			t.Fatal(err)
		}
		objectURL, err := svc.UploadFile(context.Background(), []byte("private fixture"), "fixture.txt", types.ObjectCannedACLAuthenticatedRead, "text/plain")
		if err != nil {
			t.Fatal(err)
		}
		mu.Lock()
		objectPath, requestCount := uploaded, requests
		mu.Unlock()
		if objectPath == "" || objectURL != publicOrigin+objectPath || requestCount != 2 {
			t.Fatal("upload did not use the private client and public object origin")
		}
		key := strings.TrimPrefix(objectPath, "/xem-files/")
		signed, err := svc.GetSignedURL(context.Background(), key, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := url.Parse(signed)
		if err != nil || parsed.Scheme != "https" || parsed.Host != "xem.example.test:8443" || parsed.Path != objectPath {
			t.Fatal("signed URL lost the public origin or bucket path")
		}
		query := parsed.Query()
		if query.Get("X-Amz-Signature") == "" || query.Get("X-Amz-SignedHeaders") != "host" || query.Get("X-Amz-Expires") != "60" || !strings.Contains(query.Get("X-Amz-Credential"), "/us-east-1/s3/aws4_request") {
			t.Fatal("signed URL lost its host binding, expiry or signing region")
		}
		mu.Lock()
		requestCount = requests
		mu.Unlock()
		if requestCount != 2 {
			t.Fatal("presigning unexpectedly sent a storage request")
		}
	})

	t.Run("invalid public origin", func(t *testing.T) {
		for _, invalid := range []string{"http://host", "ftp://host", "https://user:pass@host", "https://host/path", "https://host?key=value", "https://host#fragment", "/relative", "://"} {
			if _, err := resolveS3PublicEndpoint(invalid, "http://storage:9000", "us-east-1"); err == nil {
				t.Errorf("accepted public endpoint %q", invalid)
			}
		}
		if _, err := resolveS3PublicEndpoint("https://xem.example.test", "", "us-east-1"); err == nil {
			t.Error("accepted public origin without an explicit private endpoint")
		}
		if endpoint, err := resolveS3PublicEndpoint("", "", ""); err != nil || endpoint != "" {
			t.Error("optional public origin changed existing storage behavior")
		}
	})
}
