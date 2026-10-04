package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestS3BucketBootstrap(t *testing.T) {
	for _, test := range []struct {
		name       string
		create     string
		firstError string
		putError   string
		lastError  string
		wantLists  int
		wantPuts   int
		wantError  bool
	}{
		{name: "existing private bucket", create: "true", wantLists: 1},
		{name: "external bucket creation disabled", firstError: "NoSuchBucket", wantLists: 1, wantError: true},
		{name: "bundled bucket created", create: "true", firstError: "NoSuchBucket", wantLists: 2, wantPuts: 1},
		{name: "access denied never creates", create: "true", firstError: "AccessDenied", wantLists: 1, wantError: true},
		{name: "creation denied", create: "true", firstError: "NoSuchBucket", putError: "AccessDenied", wantLists: 1, wantPuts: 1, wantError: true},
		{name: "owned create race rechecks access", create: "true", firstError: "NoSuchBucket", putError: "BucketAlreadyOwnedByYou", wantLists: 2, wantPuts: 1},
		{name: "unowned bucket remains error", create: "true", firstError: "NoSuchBucket", putError: "BucketAlreadyExists", wantLists: 1, wantPuts: 1, wantError: true},
		{name: "created bucket must be accessible", create: "true", firstError: "NoSuchBucket", lastError: "AccessDenied", wantLists: 2, wantPuts: 1, wantError: true},
		{name: "flag must be explicit", create: "yes", firstError: "NoSuchBucket", wantLists: 1, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var mu sync.Mutex
			lists, puts := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				w.Header().Set("Content-Type", "application/xml")
				if r.URL.Path != "/xem-files" || r.Header.Get("Authorization") == "" || r.Header.Get("X-Amz-Acl") != "" {
					t.Error("unexpected bucket request or authorization")
				}
				var code string
				switch r.Method {
				case http.MethodGet:
					lists++
					if r.URL.Query().Get("max-keys") != "1" {
						t.Error("bucket verification did not bound list size")
					}
					if lists == 1 {
						code = test.firstError
					} else {
						code = test.lastError
					}
				case http.MethodPut:
					puts++
					code = test.putError
				default:
					t.Errorf("unexpected bucket method %s", r.Method)
				}
				if code != "" {
					status := http.StatusConflict
					if code == "NoSuchBucket" {
						status = http.StatusNotFound
					} else if code == "AccessDenied" {
						status = http.StatusForbidden
					}
					w.WriteHeader(status)
					_, _ = fmt.Fprintf(w, "<Error><Code>%s</Code><Message>fixture</Message></Error>", code)
					return
				}
				if r.Method == http.MethodGet {
					_, _ = w.Write([]byte(`<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Name>xem-files</Name></ListBucketResult>`))
				}
			}))
			defer server.Close()
			t.Setenv("S3_ENDPOINT_URL", server.URL)
			t.Setenv("S3_PUBLIC_ENDPOINT_URL", "")
			t.Setenv("S3_CREATE_BUCKET", test.create)
			_, err := NewS3Service("xem-files", "", "us-east-1", "hakopod", "fixture-secret")
			if (err != nil) != test.wantError {
				t.Fatalf("storage initialization error=%v; want error=%v", err, test.wantError)
			}
			mu.Lock()
			listCount, putCount := lists, puts
			mu.Unlock()
			if listCount != test.wantLists || putCount != test.wantPuts {
				t.Errorf("got %d list and %d create requests; want %d and %d", listCount, putCount, test.wantLists, test.wantPuts)
			}
		})
	}
}
