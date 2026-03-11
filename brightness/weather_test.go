package brightness

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchCloudCover(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"current":{"cloud_cover":75}}`)
	}))
	defer srv.Close()

	cover, err := FetchCloudCover(srv.Client(), srv.URL, 37.77, -122.43)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cover != 75 {
		t.Errorf("got %d, want 75", cover)
	}
}

func TestFetchCloudCoverError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	_, err := FetchCloudCover(srv.Client(), srv.URL, 37.77, -122.43)
	if err == nil {
		t.Fatal("expected error for 503 response")
	}
}
