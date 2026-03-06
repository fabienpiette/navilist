package navidrome

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	var r ndResponse[[]Song]
	return r.Data, json.NewDecoder(resp.Body).Decode(&r)
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
	var r ndResponse[[]Song]
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Song{}, false, err
	}
	if len(r.Data) == 0 {
		return Song{}, false, nil
	}
	return r.Data[0], true, nil
}
