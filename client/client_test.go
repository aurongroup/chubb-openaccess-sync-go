package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient starts an httptest.Server, registers cleanup, and returns a
// Client pointed at it. The Client is constructed directly to avoid coupling
// tests to Ping and Authenticate.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &Client{
		baseURL:  srv.URL,
		appID:    "test-app",
		pageSize: 10,
		http:     srv.Client(),
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func instancePage(totalPages int, items ...map[string]any) InstanceResponse {
	list := make([]InstanceItem, len(items))
	for i, m := range items {
		list[i] = InstanceItem{PropertyMap: m}
	}
	return InstanceResponse{TotalPages: totalPages, ItemList: list}
}

// ---- Ping ----

func TestPing_shouldSucceedOn200(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	if err := c.Ping(); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestPing_shouldReturnErrorWhenServerUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	c := &Client{
		baseURL: srv.URL,
		appID:   "test-app",
		http:    srv.Client(),
	}
	srv.Close()
	if err := c.Ping(); err == nil {
		t.Error("expected error for unreachable server")
	}
}

// ---- Authenticate ----

func TestAuthenticate_shouldStoreSessionToken(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"session_token": "tok-123"})
	})
	if err := c.Authenticate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.sessionToken != "tok-123" {
		t.Errorf("expected session token %q, got %q", "tok-123", c.sessionToken)
	}
}

func TestAuthenticate_shouldReturnErrorOnNon200(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	if err := c.Authenticate(); err == nil {
		t.Error("expected error for 401 response")
	}
}

func TestAuthenticate_shouldReturnErrorWhenTokenMissing(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{})
	})
	if err := c.Authenticate(); err == nil {
		t.Error("expected error when session_token absent from response")
	}
}

// ---- Close ----

func TestClose_shouldSendDeleteAndClearToken(t *testing.T) {
	called := false
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			called = true
		}
		w.WriteHeader(http.StatusOK)
	})
	c.sessionToken = "tok-to-clear"

	if err := c.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected DELETE request to be sent")
	}
	if c.sessionToken != "" {
		t.Errorf("expected session token to be cleared, got %q", c.sessionToken)
	}
}

func TestClose_shouldBeNoOpWhenTokenEmpty(t *testing.T) {
	c := &Client{sessionToken: ""}
	if err := c.Close(); err != nil {
		t.Errorf("expected nil for empty token, got %v", err)
	}
}

// ---- GetInstances ----

func TestGetInstances_shouldReturnItemsFromSinglePage(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, instancePage(1,
			map[string]any{"ID": float64(1), "Name": "Alpha"},
			map[string]any{"ID": float64(2), "Name": "Beta"},
		))
	})

	items, err := c.GetInstances("Lnl_AccessLevel", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestGetInstances_shouldAccumulateAcrossMultiplePages(t *testing.T) {
	page := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		page++
		switch page {
		case 1:
			writeJSON(w, instancePage(2, map[string]any{"ID": float64(1)}))
		default:
			writeJSON(w, instancePage(2, map[string]any{"ID": float64(2)}))
		}
	})

	items, err := c.GetInstances("Lnl_AccessLevel", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items across 2 pages, got %d", len(items))
	}
}

func TestGetInstances_shouldReturnErrorOnNon200(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	if _, err := c.GetInstances("Lnl_AccessLevel", ""); err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestGetInstances_shouldReturnErrorOnMalformedJSON(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	})
	if _, err := c.GetInstances("Lnl_AccessLevel", ""); err == nil {
		t.Error("expected error for malformed JSON response")
	}
}

// ---- DeleteInstance ----

func TestDeleteInstance_shouldReturnNilOn200(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	if err := c.DeleteInstance("Lnl_Badge", map[string]any{"ID": 1}); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestDeleteInstance_shouldReturnErrorOnNon200(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	if err := c.DeleteInstance("Lnl_Badge", map[string]any{"ID": 1}); err == nil {
		t.Error("expected error for 400 response")
	}
}
