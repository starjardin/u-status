package checker_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/u-status/internal/checker"
)

func TestCheckURL_Up(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	result := checker.CheckURL(srv.URL)
	if !result.IsUp {
		t.Fatalf("expected IsUp=true, got false (error: %v)", result.Error)
	}
	if result.StatusCode == nil || *result.StatusCode != 200 {
		t.Fatalf("expected status 200, got %v", result.StatusCode)
	}
	if result.ResponseTimeMs == nil || *result.ResponseTimeMs < 0 {
		t.Fatal("expected non-nil response time")
	}
}

func TestCheckURL_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	result := checker.CheckURL(srv.URL)
	if result.IsUp {
		t.Fatal("expected IsUp=false for 500")
	}
}

func TestCheckURL_Redirect(t *testing.T) {
	// 3xx should be treated as UP
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/other", http.StatusMovedPermanently)
	}))
	defer srv.Close()

	result := checker.CheckURL(srv.URL)
	if !result.IsUp {
		t.Fatalf("expected 3xx to be UP, got down (error: %v)", result.Error)
	}
}

func TestCheckURL_ConnectionRefused(t *testing.T) {
	// Port with no listener
	result := checker.CheckURL("http://127.0.0.1:19999")
	if result.IsUp {
		t.Fatal("expected IsUp=false for refused connection")
	}
	if result.Error == nil || *result.Error == "" {
		t.Fatal("expected error message")
	}
}
