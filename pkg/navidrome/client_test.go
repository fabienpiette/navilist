package navidrome_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/navilist/pkg/navidrome"
)

func TestClientAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/login" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
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
	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"token": "tok"})
	})
	mux.HandleFunc("/", handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, navidrome.New(srv.URL, "admin", "admin")
}

func TestListPlaylists(t *testing.T) {
	_, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/playlist" {
			return
		}
		w.Header().Set("X-Total-Count", "1")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "1", "name": "Test", "songCount": 5},
		})
	})
	playlists, err := c.ListPlaylists(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(playlists) != 1 || playlists[0].Name != "Test" {
		t.Fatalf("unexpected: %+v", playlists)
	}
}

func TestDeletePlaylist(t *testing.T) {
	deleted := false
	_, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/api/playlist/abc" {
			deleted = true
			json.NewEncoder(w).Encode(map[string]any{})
		}
	})
	if err := c.DeletePlaylist(context.Background(), "abc"); err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Fatal("DELETE not called")
	}
}

func TestSearchSongs(t *testing.T) {
	_, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/song" {
			return
		}
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "s1", "title": "Around the World", "artist": "Daft Punk"},
		})
	})
	songs, err := c.SearchSongs(context.Background(), "Around", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(songs) != 1 {
		t.Fatalf("expected 1 song, got %d", len(songs))
	}
}

func TestSearchArtists(t *testing.T) {
	_, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/artist" {
			return
		}
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "a1", "name": "Daft Punk"},
			{"id": "a2", "name": "Daft Punk Tribute"},
		})
	})
	names, err := c.SearchArtists(context.Background(), "Daft", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "Daft Punk" {
		t.Fatalf("unexpected: %+v", names)
	}
}

func TestSearchAlbums(t *testing.T) {
	_, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/album" {
			return
		}
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "al1", "name": "Discovery"},
		})
	})
	names, err := c.SearchAlbums(context.Background(), "Disc", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "Discovery" {
		t.Fatalf("unexpected: %+v", names)
	}
}

func TestSearchGenres(t *testing.T) {
	_, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/genre" {
			return
		}
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "g1", "name": "Rock"},
			{"id": "g2", "name": "Post-Rock"},
			{"id": "g3", "name": "Jazz"},
		})
	})
	names, err := c.SearchGenres(context.Background(), "rock", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 matches for 'rock', got %d: %v", len(names), names)
	}
}
