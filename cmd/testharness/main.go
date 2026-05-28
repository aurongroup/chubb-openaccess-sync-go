package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type authRequest struct {
	UserName    string `json:"user_name"`
	Password    string `json:"password"`
	DirectoryID string `json:"directory_id"`
}

type instanceItem struct {
	PropertyMap map[string]any `json:"property_value_map"`
}

type instanceResponse struct {
	PageNumber int            `json:"page_number"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
	TotalItems int            `json:"total_items"`
	Count      int            `json:"count"`
	ItemList   []instanceItem `json:"item_list"`
	TypeName   string         `json:"type_name"`
}

type instanceBody struct {
	TypeName    string         `json:"type_name"`
	PropertyMap map[string]any `json:"property_value_map"`
}

type store struct {
	mu        sync.Mutex
	sessions  map[string]time.Time
	instances map[string][]map[string]any
	timeout   time.Duration
}

func newStore(timeout time.Duration) *store {
	return &store{
		sessions:  make(map[string]time.Time),
		instances: make(map[string][]map[string]any),
		timeout:   timeout,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func checkAcceptAndAppID(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Accept") == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing Accept header"})
		return false
	}
	if r.Header.Get("Application-Id") == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing Application-Id header"})
		return false
	}
	return true
}

func checkSessionToken(w http.ResponseWriter, r *http.Request, s *store) bool {
	token := r.Header.Get("Session-Token")
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing Session-Token header"})
		return false
	}
	s.mu.Lock()
	expiry, ok := s.sessions[token]
	s.mu.Unlock()
	if !ok || time.Now().After(expiry) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired session token"})
		return false
	}
	return true
}

func generateToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func handleAuthentication(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAcceptAndAppID(w, r) {
			return
		}

		switch r.Method {
		case http.MethodPost:
			var req authRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
				return
			}
			token, err := generateToken()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
				return
			}
			expiry := time.Now().Add(s.timeout)
			s.mu.Lock()
			s.sessions[token] = expiry
			s.mu.Unlock()
			log.Printf("auth: created session for user=%s directory=%s expires=%s", req.UserName, req.DirectoryID, expiry.Format(time.RFC3339))
			writeJSON(w, http.StatusOK, map[string]string{
				"password_expiration_time": "2026-08-08T14:43:36+00:00",
				"session_token":            token,
				"token_expiration_time":    expiry.UTC().Format(time.RFC3339),
				"version":                  "1.0",
			})

		case http.MethodDelete:
			token := r.Header.Get("Session-Token")
			if token != "" {
				s.mu.Lock()
				delete(s.sessions, token)
				s.mu.Unlock()
				log.Printf("auth: deleted session token=%s", token)
			}
			writeJSON(w, http.StatusOK, map[string]any{})

		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	}
}

func queryInt(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return def
	}
	return v
}

func paginate(items []map[string]any, page, size int) ([]map[string]any, int) {
	total := len(items)
	totalPages := (total + size - 1) / size
	if totalPages == 0 {
		totalPages = 1
	}
	start := (page - 1) * size
	if start >= total {
		return nil, totalPages
	}
	end := start + size
	if end > total {
		end = total
	}
	return items[start:end], totalPages
}

func handleInstances(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAcceptAndAppID(w, r) {
			return
		}
		if !checkSessionToken(w, r, s) {
			return
		}

		switch r.Method {
		case http.MethodGet:
			typeName := r.URL.Query().Get("type_name")
			page := queryInt(r, "page_number", 1)
			size := queryInt(r, "page_size", 10)

			s.mu.Lock()
			all := s.instances[typeName]
			s.mu.Unlock()

			slice, totalPages := paginate(all, page, size)
			items := make([]instanceItem, len(slice))
			for i, m := range slice {
				items[i] = instanceItem{PropertyMap: m}
			}
			writeJSON(w, http.StatusOK, instanceResponse{
				PageNumber: page,
				PageSize:   size,
				TotalPages: totalPages,
				TotalItems: len(all),
				Count:      len(items),
				ItemList:   items,
				TypeName:   typeName,
			})

		case http.MethodPost:
			var body instanceBody
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
				return
			}
			s.mu.Lock()
			s.instances[body.TypeName] = append(s.instances[body.TypeName], body.PropertyMap)
			id := len(s.instances[body.TypeName])
			s.mu.Unlock()
			log.Printf("instances: created %s id=%d", body.TypeName, id)
			writeJSON(w, http.StatusOK, map[string]any{
				"property_value_map": map[string]any{"ID": id},
				"type_name":          body.TypeName,
				"version":            "1.0",
			})

		case http.MethodPut:
			var body instanceBody
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
				return
			}
			targetID := body.PropertyMap["ID"]
			s.mu.Lock()
			list := s.instances[body.TypeName]
			found := false
			for i, m := range list {
				if m["ID"] == targetID {
					list[i] = body.PropertyMap
					found = true
					break
				}
			}
			s.mu.Unlock()
			if !found {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "instance not found"})
				return
			}
			log.Printf("instances: updated %s id=%v", body.TypeName, targetID)
			writeJSON(w, http.StatusOK, map[string]any{})

		case http.MethodDelete:
			var body instanceBody
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
				return
			}
			targetID := body.PropertyMap["ID"]
			s.mu.Lock()
			list := s.instances[body.TypeName]
			filtered := list[:0]
			for _, m := range list {
				if m["ID"] != targetID {
					filtered = append(filtered, m)
				}
			}
			s.instances[body.TypeName] = filtered
			s.mu.Unlock()
			log.Printf("instances: deleted %s id=%v", body.TypeName, targetID)
			writeJSON(w, http.StatusOK, map[string]any{})

		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	}
}

func handleCardholders(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAcceptAndAppID(w, r) {
			return
		}
		if !checkSessionToken(w, r, s) {
			return
		}

		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		page := queryInt(r, "page_number", 1)
		size := queryInt(r, "page_size", 10)

		s.mu.Lock()
		all := s.instances["Lnl_Cardholder"]
		s.mu.Unlock()

		slice, totalPages := paginate(all, page, size)
		items := make([]instanceItem, len(slice))
		for i, m := range slice {
			items[i] = instanceItem{PropertyMap: m}
		}
		writeJSON(w, http.StatusOK, instanceResponse{
			PageNumber: page,
			PageSize:   size,
			TotalPages: totalPages,
			TotalItems: len(all),
			Count:      len(items),
			ItemList:   items,
			TypeName:   "Lnl_Cardholder",
		})
	}
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	sessionTimeout := flag.Duration("session-timeout", 5*time.Minute, "session token lifetime")
	flag.Parse()

	s := newStore(*sessionTimeout)

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"product_name":    "OnGuard 8.1",
			"product_version": "8.1.639.0",
			"version":         "1.0",
		})
	})
	http.HandleFunc("/authentication", handleAuthentication(s))
	http.HandleFunc("/instances", handleInstances(s))
	http.HandleFunc("/cardholders", handleCardholders(s))

	log.Printf("test harness listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
