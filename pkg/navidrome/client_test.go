package navidrome_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/navidrome-playlists/pkg/navidrome"
)

func TestClientAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/login" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]string{"token": "test-token"},
		})
	}))
	defer srv.Close()

	c := navidrome.New(srv.URL, "admin", "admin")
	if err := c.Authenticate(); err != nil {
		t.Fatalf("Authenticate() error: %v", err)
	}
}

func newMockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *navidrome.Client) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]string{"token": "tok"}})
	})
	mux.HandleFunc("/", handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, navidrome.New(srv.URL, "admin", "admin")
}
