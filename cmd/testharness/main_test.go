package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// testSchema is a minimal schema used across all tests.
var testSchema = map[string]typeSchema{
	"Thing": {
		Required:  []string{"NAME"},
		Defaults:  map[string]any{"COLOR": "red"},
		AutoTime:  "CHANGED_AT",
		Immutable: []string{"OWNER"},
	},
	"Ref": {
		Required:   []string{"THING_ID"},
		References: []reference{{Field: "THING_ID", Type: "Thing", Key: "ID"}},
	},
	"ReadOnly": {
		Required: []string{"NAME"},
		ReadOnly: true,
	},
	"Lnl_Badge": {
		Required:  []string{"CARDNUM", "PERSONID"},
		Immutable: []string{"PERSONID"},
	},
}

// setupServer creates a fresh test server for each test.
func setupServer(t *testing.T) (*httptest.Server, *store) {
	t.Helper()
	s := newStore(time.Hour)
	s.schema = testSchema
	mux := http.NewServeMux()
	mux.HandleFunc("/authentication", handleAuthentication(s))
	mux.HandleFunc("/instances", handleInstances(s))
	mux.HandleFunc("/cardholders", handleCardholders(s))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, s
}

// login authenticates and returns a session token.
func login(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/authentication",
		strings.NewReader(`{"user_name":"test","password":"x","directory_id":"1"}`))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Application-Id", "test")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer resp.Body.Close()
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	token := result["session_token"]
	if token == "" {
		t.Fatal("login: no session_token in response")
	}
	return token
}

// call sends a JSON request with the standard headers and decodes the response body.
func call(t *testing.T, srv *httptest.Server, token, method, path string, body any, out any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("call encode: %v", err)
		}
	}
	req, _ := http.NewRequest(method, srv.URL+path, &buf)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Application-Id", "test")
	if token != "" {
		req.Header.Set("Session-Token", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("call %s %s: %v", method, path, err)
	}
	if out != nil {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("call decode: %v", err)
		}
	}
	return resp
}

// propMap extracts property_value_map from a POST/PUT response body.
func propMap(t *testing.T, v map[string]any) map[string]any {
	t.Helper()
	pm, ok := v["property_value_map"].(map[string]any)
	if !ok {
		t.Fatalf("property_value_map missing or wrong type in %v", v)
	}
	return pm
}

// ---- Authentication -------------------------------------------------------

func TestAuthentication_Post(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)
	if token == "" {
		t.Error("expected non-empty session token")
	}
}

func TestAuthentication_Delete_InvalidatesToken(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/authentication", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Application-Id", "test")
	req.Header.Set("Session-Token", token)
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("logout: expected 200, got %d", resp.StatusCode)
	}

	resp2 := call(t, srv, token, http.MethodGet, "/instances?type_name=Thing", nil, nil)
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 after logout, got %d", resp2.StatusCode)
	}
}

func TestAuthentication_MissingHeaders(t *testing.T) {
	srv, _ := setupServer(t)

	t.Run("missing Accept", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/authentication",
			strings.NewReader(`{}`))
		req.Header.Set("Application-Id", "test")
		resp, _ := http.DefaultClient.Do(req)
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("missing Application-Id", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/authentication",
			strings.NewReader(`{}`))
		req.Header.Set("Accept", "application/json")
		resp, _ := http.DefaultClient.Do(req)
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})
}

func TestAuthentication_UnsupportedMethod(t *testing.T) {
	srv, _ := setupServer(t)
	resp := call(t, srv, "", http.MethodGet, "/authentication", nil, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

// ---- Instances: GET -------------------------------------------------------

func TestInstancesGet_All(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Thing"] = []map[string]any{
		{"ID": float64(1), "NAME": "alpha", "COLOR": "red"},
		{"ID": float64(2), "NAME": "beta", "COLOR": "blue"},
		{"ID": float64(3), "NAME": "gamma", "COLOR": "red"},
	}
	s.rebuildIndex("Thing")
	s.mu.Unlock()

	var result instanceResponse
	resp := call(t, srv, token, http.MethodGet, "/instances?type_name=Thing&page_size=10", nil, &result)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if result.TotalItems != 3 || result.Count != 3 {
		t.Errorf("expected 3 items, got TotalItems=%d Count=%d", result.TotalItems, result.Count)
	}
}

func TestInstancesGet_FilterEquals(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Thing"] = []map[string]any{
		{"ID": float64(1), "NAME": "alpha", "COLOR": "red"},
		{"ID": float64(2), "NAME": "beta", "COLOR": "blue"},
		{"ID": float64(3), "NAME": "gamma", "COLOR": "red"},
	}
	s.rebuildIndex("Thing")
	s.mu.Unlock()

	var result instanceResponse
	call(t, srv, token, http.MethodGet, "/instances?type_name=Thing&filter=COLOR=red", nil, &result).Body.Close()
	if result.TotalItems != 2 {
		t.Errorf("expected 2 red items, got %d", result.TotalItems)
	}
}

func TestInstancesGet_FilterNotEquals(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Thing"] = []map[string]any{
		{"ID": float64(1), "NAME": "alpha", "COLOR": "red"},
		{"ID": float64(2), "NAME": "beta", "COLOR": "blue"},
	}
	s.rebuildIndex("Thing")
	s.mu.Unlock()

	var result instanceResponse
	call(t, srv, token, http.MethodGet, "/instances?type_name=Thing&filter=COLOR!=red", nil, &result).Body.Close()
	if result.TotalItems != 1 {
		t.Errorf("expected 1 non-red item, got %d", result.TotalItems)
	}
}

func TestInstancesGet_FilterEqualsNull(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Thing"] = []map[string]any{
		{"ID": float64(1), "NAME": "alpha", "COLOR": "red"},
		{"ID": float64(2), "NAME": "beta"},
	}
	s.rebuildIndex("Thing")
	s.mu.Unlock()

	var result instanceResponse
	call(t, srv, token, http.MethodGet, "/instances?type_name=Thing&filter=COLOR=null", nil, &result).Body.Close()
	if result.TotalItems != 1 {
		t.Errorf("expected 1 item with null COLOR, got %d", result.TotalItems)
	}
}

func TestInstancesGet_FilterNotEqualsNull(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Thing"] = []map[string]any{
		{"ID": float64(1), "NAME": "alpha", "COLOR": "red"},
		{"ID": float64(2), "NAME": "beta"},
	}
	s.rebuildIndex("Thing")
	s.mu.Unlock()

	var result instanceResponse
	call(t, srv, token, http.MethodGet, "/instances?type_name=Thing&filter=COLOR!=null", nil, &result).Body.Close()
	if result.TotalItems != 1 {
		t.Errorf("expected 1 item with non-null COLOR, got %d", result.TotalItems)
	}
}

func TestInstancesGet_Pagination(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Thing"] = []map[string]any{
		{"ID": float64(1), "NAME": "a"},
		{"ID": float64(2), "NAME": "b"},
		{"ID": float64(3), "NAME": "c"},
	}
	s.rebuildIndex("Thing")
	s.mu.Unlock()

	var p1 instanceResponse
	call(t, srv, token, http.MethodGet, "/instances?type_name=Thing&page_size=2&page_number=1", nil, &p1).Body.Close()
	if p1.Count != 2 || p1.TotalPages != 2 {
		t.Errorf("page 1: expected count=2 totalPages=2, got count=%d totalPages=%d", p1.Count, p1.TotalPages)
	}

	var p2 instanceResponse
	call(t, srv, token, http.MethodGet, "/instances?type_name=Thing&page_size=2&page_number=2", nil, &p2).Body.Close()
	if p2.Count != 1 {
		t.Errorf("page 2: expected count=1, got %d", p2.Count)
	}
}

func TestInstancesGet_InvalidFilter(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)
	resp := call(t, srv, token, http.MethodGet, "/instances?type_name=Thing&filter=NOOPERATOR", nil, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid filter, got %d", resp.StatusCode)
	}
}

func TestInstancesGet_MissingSessionToken(t *testing.T) {
	srv, _ := setupServer(t)
	resp := call(t, srv, "", http.MethodGet, "/instances?type_name=Thing", nil, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// ---- Instances: POST ------------------------------------------------------

func TestInstancesPost_AssignsID(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)

	var result map[string]any
	resp := call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"NAME": "foo"}}, &result)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	pm := propMap(t, result)
	if pm["ID"] != float64(1) {
		t.Errorf("expected ID=1, got %v", pm["ID"])
	}
}

func TestInstancesPost_AutoIncrements(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)

	var r1, r2 map[string]any
	call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"NAME": "a"}}, &r1).Body.Close()
	call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"NAME": "b"}}, &r2).Body.Close()

	id1 := propMap(t, r1)["ID"].(float64)
	id2 := propMap(t, r2)["ID"].(float64)
	if id2 != id1+1 {
		t.Errorf("expected sequential IDs, got %v and %v", id1, id2)
	}
}

func TestInstancesPost_IncrementsPastLoadedMax(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Thing"] = []map[string]any{{"ID": float64(50), "NAME": "existing"}}
	s.rebuildIndex("Thing")
	s.initNextID("Thing")
	s.mu.Unlock()

	var result map[string]any
	call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"NAME": "new"}}, &result).Body.Close()
	pm := propMap(t, result)
	if pm["ID"] != float64(51) {
		t.Errorf("expected ID=51 (past loaded max of 50), got %v", pm["ID"])
	}
}

func TestInstancesPost_AppliesDefaults(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)

	var result map[string]any
	call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"NAME": "foo"}}, &result).Body.Close()
	pm := propMap(t, result)
	if pm["COLOR"] != "red" {
		t.Errorf("expected COLOR=red from defaults, got %v", pm["COLOR"])
	}
}

func TestInstancesPost_SubmittedOverridesDefault(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)

	var result map[string]any
	call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"NAME": "foo", "COLOR": "green"}}, &result).Body.Close()
	pm := propMap(t, result)
	if pm["COLOR"] != "green" {
		t.Errorf("expected COLOR=green, got %v", pm["COLOR"])
	}
}

func TestInstancesPost_SetsAutoTime(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)

	var result map[string]any
	call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"NAME": "foo"}}, &result).Body.Close()
	pm := propMap(t, result)
	if v, ok := pm["CHANGED_AT"].(string); !ok || v == "" {
		t.Errorf("expected CHANGED_AT to be set, got %v", pm["CHANGED_AT"])
	}
}

func TestInstancesPost_MissingRequired(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)

	var result map[string]string
	resp := call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"COLOR": "blue"}}, &result)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if !strings.Contains(result["error"], "NAME") {
		t.Errorf("expected error to mention NAME, got %q", result["error"])
	}
}

func TestInstancesPost_FKViolation(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)

	var result map[string]string
	resp := call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Ref", PropertyMap: map[string]any{"THING_ID": float64(9999)}}, &result)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if !strings.Contains(result["error"], "THING_ID") {
		t.Errorf("expected error to mention THING_ID, got %q", result["error"])
	}
}

func TestInstancesPost_FKValid(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)

	var thingResult map[string]any
	call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"NAME": "target"}}, &thingResult).Body.Close()
	thingID := propMap(t, thingResult)["ID"]

	resp := call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Ref", PropertyMap: map[string]any{"THING_ID": thingID}}, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 with valid FK, got %d", resp.StatusCode)
	}
}

func TestInstancesPost_LnlBadgeUsesBadgeKey(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)

	var result map[string]any
	resp := call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Lnl_Badge", PropertyMap: map[string]any{"CARDNUM": "12345", "PERSONID": float64(1)}}, &result)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	pm := propMap(t, result)
	if pm["BADGEKEY"] == nil {
		t.Error("expected BADGEKEY in response")
	}
	if _, hasID := pm["ID"]; hasID {
		t.Error("Lnl_Badge response should not contain an ID field")
	}
}

func TestInstancesPost_ReturnedObjectStoredInIndex(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)

	call(t, srv, token, http.MethodPost, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"NAME": "findme", "COLOR": "purple"}}, nil).Body.Close()

	var result instanceResponse
	call(t, srv, token, http.MethodGet, "/instances?type_name=Thing&filter=COLOR=purple", nil, &result).Body.Close()
	if result.TotalItems != 1 {
		t.Errorf("expected 1 result for filter on created instance, got %d", result.TotalItems)
	}
	_ = s
}

// ---- Instances: PUT -------------------------------------------------------

func TestInstancesPut_UpdatesRecord(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Thing"] = []map[string]any{{"ID": float64(1), "NAME": "original", "OWNER": "alice"}}
	s.rebuildIndex("Thing")
	s.nextID["Thing"] = 1
	s.mu.Unlock()

	resp := call(t, srv, token, http.MethodPut, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"ID": float64(1), "NAME": "updated", "OWNER": "alice"}}, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	s.mu.Lock()
	name := s.instances["Thing"][0]["NAME"]
	s.mu.Unlock()
	if name != "updated" {
		t.Errorf("expected NAME=updated, got %v", name)
	}
}

func TestInstancesPut_SetsAutoTime(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Thing"] = []map[string]any{{"ID": float64(1), "NAME": "x", "OWNER": "alice"}}
	s.rebuildIndex("Thing")
	s.mu.Unlock()

	call(t, srv, token, http.MethodPut, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"ID": float64(1), "NAME": "x", "OWNER": "alice"}}, nil).Body.Close()

	s.mu.Lock()
	changedAt := s.instances["Thing"][0]["CHANGED_AT"]
	s.mu.Unlock()
	if v, ok := changedAt.(string); !ok || v == "" {
		t.Error("expected CHANGED_AT to be set on PUT")
	}
}

func TestInstancesPut_NotFound(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)

	resp := call(t, srv, token, http.MethodPut, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"ID": float64(9999), "NAME": "ghost"}}, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestInstancesPut_ImmutableFieldChanged(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Thing"] = []map[string]any{{"ID": float64(1), "NAME": "x", "OWNER": "alice"}}
	s.rebuildIndex("Thing")
	s.mu.Unlock()

	var result map[string]string
	resp := call(t, srv, token, http.MethodPut, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"ID": float64(1), "NAME": "x", "OWNER": "bob"}}, &result)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if !strings.Contains(result["error"], "OWNER") {
		t.Errorf("expected error to mention OWNER, got %q", result["error"])
	}
}

func TestInstancesPut_ImmutableFieldUnchanged(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Thing"] = []map[string]any{{"ID": float64(1), "NAME": "x", "OWNER": "alice"}}
	s.rebuildIndex("Thing")
	s.mu.Unlock()

	resp := call(t, srv, token, http.MethodPut, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"ID": float64(1), "NAME": "y", "OWNER": "alice"}}, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 when immutable field is unchanged, got %d", resp.StatusCode)
	}
}

func TestInstancesPut_ReadOnly(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["ReadOnly"] = []map[string]any{{"ID": float64(1), "NAME": "fixed"}}
	s.rebuildIndex("ReadOnly")
	s.mu.Unlock()

	resp := call(t, srv, token, http.MethodPut, "/instances",
		instanceBody{TypeName: "ReadOnly", PropertyMap: map[string]any{"ID": float64(1), "NAME": "changed"}}, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for readonly type, got %d", resp.StatusCode)
	}
}

// ---- Instances: DELETE ----------------------------------------------------

func TestInstancesDelete_ByID(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Thing"] = []map[string]any{
		{"ID": float64(1), "NAME": "to-delete"},
		{"ID": float64(2), "NAME": "keep"},
	}
	s.rebuildIndex("Thing")
	s.mu.Unlock()

	resp := call(t, srv, token, http.MethodDelete, "/instances",
		instanceBody{TypeName: "Thing", PropertyMap: map[string]any{"ID": float64(1)}}, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	s.mu.Lock()
	remaining := len(s.instances["Thing"])
	s.mu.Unlock()
	if remaining != 1 {
		t.Errorf("expected 1 item remaining, got %d", remaining)
	}
}

func TestInstancesDelete_LnlBadgeByBadgeKey(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Lnl_Badge"] = []map[string]any{
		{"BADGEKEY": float64(10), "CARDNUM": "AAA"},
		{"BADGEKEY": float64(20), "CARDNUM": "BBB"},
	}
	s.rebuildIndex("Lnl_Badge")
	s.mu.Unlock()

	resp := call(t, srv, token, http.MethodDelete, "/instances",
		instanceBody{TypeName: "Lnl_Badge", PropertyMap: map[string]any{"BADGEKEY": float64(10)}}, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	s.mu.Lock()
	remaining := s.instances["Lnl_Badge"]
	s.mu.Unlock()
	if len(remaining) != 1 || remaining[0]["BADGEKEY"] != float64(20) {
		t.Errorf("expected only BADGEKEY=20 remaining, got %v", remaining)
	}
}

// ---- Cardholders ----------------------------------------------------------

func TestCardholders_Get(t *testing.T) {
	srv, s := setupServer(t)
	token := login(t, srv)
	s.mu.Lock()
	s.instances["Lnl_Cardholder"] = []map[string]any{
		{"ID": float64(1), "LASTNAME": "Smith"},
		{"ID": float64(2), "LASTNAME": "Jones"},
	}
	s.rebuildIndex("Lnl_Cardholder")
	s.mu.Unlock()

	var result instanceResponse
	resp := call(t, srv, token, http.MethodGet, "/cardholders", nil, &result)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if result.TotalItems != 2 {
		t.Errorf("expected 2 cardholders, got %d", result.TotalItems)
	}
	if result.TypeName != "Lnl_Cardholder" {
		t.Errorf("expected type_name=Lnl_Cardholder, got %q", result.TypeName)
	}
}

func TestCardholders_NonGetReturns405(t *testing.T) {
	srv, _ := setupServer(t)
	token := login(t, srv)
	resp := call(t, srv, token, http.MethodPost, "/cardholders", nil, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

// ---- Store unit tests -----------------------------------------------------

func TestInitNextID_UsesMax(t *testing.T) {
	s := newStore(time.Hour)
	s.instances["Thing"] = []map[string]any{
		{"ID": float64(5)},
		{"ID": float64(3)},
		{"ID": float64(10)},
	}
	s.rebuildIndex("Thing")
	s.initNextID("Thing")
	if s.nextID["Thing"] != 10 {
		t.Errorf("expected nextID=10, got %d", s.nextID["Thing"])
	}
}

func TestInitNextID_LnlBadgeUsesBadgeKey(t *testing.T) {
	s := newStore(time.Hour)
	s.instances["Lnl_Badge"] = []map[string]any{
		{"BADGEKEY": float64(7)},
		{"BADGEKEY": float64(42)},
	}
	s.rebuildIndex("Lnl_Badge")
	s.initNextID("Lnl_Badge")
	if s.nextID["Lnl_Badge"] != 42 {
		t.Errorf("expected nextID=42, got %d", s.nextID["Lnl_Badge"])
	}
}

func TestParseFilter(t *testing.T) {
	tests := []struct {
		input   string
		field   string
		op      string
		value   string
		wantErr bool
	}{
		{"NAME = foo", "NAME", "=", "foo", false},
		{"NAME != foo", "NAME", "!=", "foo", false},
		{`NAME = "foo bar"`, "NAME", "=", "foo bar", false},
		{"NAME = null", "NAME", "=", "null", false},
		{"NAME != null", "NAME", "!=", "null", false},
		{"NOOPERATOR", "", "", "", true},
		{"= foo", "", "", "", true},
	}
	for _, tc := range tests {
		field, op, value, err := parseFilter(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseFilter(%q): expected error", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseFilter(%q): unexpected error: %v", tc.input, err)
			continue
		}
		if field != tc.field || op != tc.op || value != tc.value {
			t.Errorf("parseFilter(%q): got (%q,%q,%q) want (%q,%q,%q)",
				tc.input, field, op, value, tc.field, tc.op, tc.value)
		}
	}
}