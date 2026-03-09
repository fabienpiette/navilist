package navidrome

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) SearchSongs(ctx context.Context, query string, limit int) ([]Song, error) {
	q := url.Values{
		"title":  []string{query},
		"_start": []string{"0"},
		"_end":   []string{strconv.Itoa(limit)},
	}
	resp, err := c.Do(ctx, http.MethodGet, "/api/song?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search songs: status %d", resp.StatusCode)
	}
	var r []Song
	return r, json.NewDecoder(resp.Body).Decode(&r)
}

func (c *Client) SearchArtists(ctx context.Context, query string, limit int) ([]string, error) {
	q := url.Values{
		"name":   []string{query},
		"_start": []string{"0"},
		"_end":   []string{strconv.Itoa(limit)},
	}
	resp, err := c.Do(ctx, http.MethodGet, "/api/artist?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search artists: status %d", resp.StatusCode)
	}
	var r []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	names := make([]string, len(r))
	for i, a := range r {
		names[i] = a.Name
	}
	return names, nil
}

func (c *Client) SearchAlbums(ctx context.Context, query string, limit int) ([]string, error) {
	q := url.Values{
		"name":   []string{query},
		"_start": []string{"0"},
		"_end":   []string{strconv.Itoa(limit)},
	}
	resp, err := c.Do(ctx, http.MethodGet, "/api/album?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search albums: status %d", resp.StatusCode)
	}
	var r []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	names := make([]string, len(r))
	for i, a := range r {
		names[i] = a.Name
	}
	return names, nil
}

// SearchGenres fetches all genres and returns those containing query (case-insensitive).
// Genres are a small set, so we fetch up to 300 and filter in Go.
func (c *Client) SearchGenres(ctx context.Context, query string, limit int) ([]string, error) {
	q := url.Values{
		"_sort":  []string{"name"},
		"_start": []string{"0"},
		"_end":   []string{"300"},
	}
	resp, err := c.Do(ctx, http.MethodGet, "/api/genre?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search genres: status %d", resp.StatusCode)
	}
	var r []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	lq := strings.ToLower(query)
	var names []string
	for _, g := range r {
		if strings.Contains(strings.ToLower(g.Name), lq) {
			names = append(names, g.Name)
			if len(names) >= limit {
				break
			}
		}
	}
	return names, nil
}

func (c *Client) GetSongByPath(ctx context.Context, path string) (Song, bool, error) {
	q := url.Values{"path": []string{path}, "_start": []string{"0"}, "_end": []string{"1"}}
	resp, err := c.Do(ctx, http.MethodGet, "/api/song?"+q.Encode(), nil)
	if err != nil {
		return Song{}, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Song{}, false, fmt.Errorf("get song by path: status %d", resp.StatusCode)
	}
	var r []Song
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Song{}, false, err
	}
	if len(r) == 0 {
		return Song{}, false, nil
	}
	return r[0], true, nil
}
