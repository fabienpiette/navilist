package navidrome

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) ListPlaylists(ctx context.Context) ([]Playlist, error) {
	resp, err := c.Do(ctx, http.MethodGet, "/api/playlist?_end=500&_start=0", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list playlists: status %d", resp.StatusCode)
	}
	var r []Playlist
	return r, json.NewDecoder(resp.Body).Decode(&r)
}

func (c *Client) GetPlaylist(ctx context.Context, id string) (Playlist, error) {
	resp, err := c.Do(ctx, http.MethodGet, "/api/playlist/"+id, nil)
	if err != nil {
		return Playlist{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Playlist{}, fmt.Errorf("get playlist: status %d", resp.StatusCode)
	}
	var r Playlist
	return r, json.NewDecoder(resp.Body).Decode(&r)
}

func (c *Client) GetPlaylistTracks(ctx context.Context, id string) ([]Song, error) {
	resp, err := c.Do(ctx, http.MethodGet, "/api/playlist/"+id+"/tracks?_start=0&_end=500", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get tracks: status %d", resp.StatusCode)
	}
	var r []Song
	return r, json.NewDecoder(resp.Body).Decode(&r)
}

func (c *Client) CreatePlaylist(ctx context.Context, req CreatePlaylistRequest) (Playlist, error) {
	resp, err := c.Do(ctx, http.MethodPost, "/api/playlist", req)
	if err != nil {
		return Playlist{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Playlist{}, fmt.Errorf("create playlist: status %d", resp.StatusCode)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return Playlist{}, fmt.Errorf("create playlist decode: %w", err)
	}
	return c.GetPlaylist(ctx, created.ID)
}

func (c *Client) UpdatePlaylist(ctx context.Context, id string, req UpdatePlaylistRequest) error {
	resp, err := c.Do(ctx, http.MethodPut, "/api/playlist/"+id, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update playlist: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) DeletePlaylist(ctx context.Context, id string) error {
	resp, err := c.Do(ctx, http.MethodDelete, "/api/playlist/"+id, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete playlist: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) AddTracks(ctx context.Context, playlistID string, songIDs []string) error {
	resp, err := c.Do(ctx, http.MethodPost, "/api/playlist/"+playlistID+"/tracks", AddTracksRequest{IDs: songIDs})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("add tracks: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) RemoveTracks(ctx context.Context, playlistID string, trackItemIDs []string) error {
	q := url.Values{}
	for _, id := range trackItemIDs {
		q.Add("id", id)
	}
	resp, err := c.Do(ctx, http.MethodDelete, "/api/playlist/"+playlistID+"/tracks?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remove tracks: status %d", resp.StatusCode)
	}
	return nil
}
