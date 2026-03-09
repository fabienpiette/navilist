package handlers

import (
	"testing"

	"github.com/user/navilist/pkg/navidrome"
)

func song(artist, album string) navidrome.Song {
	return navidrome.Song{Artist: artist, Album: album}
}

func TestSuggestName_PurePlaylist(t *testing.T) {
	tracks := []navidrome.Song{
		song("Pink Floyd", "The Wall"),
		song("Pink Floyd", "The Wall"),
		song("Pink Floyd", "The Wall"),
		song("Pink Floyd", "The Wall"),
		song("Pink Floyd", "The Wall"),
	}
	got := suggestName(tracks)
	want := "Pink Floyd - The Wall"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSuggestName_JustAboveThreshold(t *testing.T) {
	// 4/5 = 80% — exactly at threshold, should suggest
	tracks := []navidrome.Song{
		song("Pink Floyd", "The Wall"),
		song("Pink Floyd", "The Wall"),
		song("Pink Floyd", "The Wall"),
		song("Pink Floyd", "The Wall"),
		song("Other", "Album"),
	}
	got := suggestName(tracks)
	want := "Pink Floyd - The Wall"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSuggestName_JustBelowThreshold(t *testing.T) {
	// 3/5 = 60% — below threshold, no suggestion
	tracks := []navidrome.Song{
		song("Pink Floyd", "The Wall"),
		song("Pink Floyd", "The Wall"),
		song("Pink Floyd", "The Wall"),
		song("Other", "Album"),
		song("Other2", "Album2"),
	}
	got := suggestName(tracks)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestSuggestName_Empty(t *testing.T) {
	got := suggestName(nil)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestSuggestName_MixedArtistSameAlbum(t *testing.T) {
	// Same album name but different artists = different keys, no suggestion
	tracks := []navidrome.Song{
		song("Artist A", "Greatest Hits"),
		song("Artist A", "Greatest Hits"),
		song("Artist B", "Greatest Hits"),
		song("Artist B", "Greatest Hits"),
		song("Artist C", "Greatest Hits"),
	}
	got := suggestName(tracks)
	if got != "" {
		t.Errorf("expected empty for mixed artists, got %q", got)
	}
}
