package m3u_test

import (
	"strings"
	"testing"

	"github.com/user/navidrome-playlists/internal/m3u"
)

func TestParseExtM3U(t *testing.T) {
	input := "#EXTM3U\n#EXTINF:253,Daft Punk - Around the World\n/music/daftpunk/around.flac\n"
	tracks, err := m3u.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Path != "/music/daftpunk/around.flac" {
		t.Errorf("wrong path: %s", tracks[0].Path)
	}
	if tracks[0].Artist != "Daft Punk" {
		t.Errorf("wrong artist: %s", tracks[0].Artist)
	}
	if tracks[0].Title != "Around the World" {
		t.Errorf("wrong title: %s", tracks[0].Title)
	}
	if tracks[0].Duration != 253 {
		t.Errorf("wrong duration: %d", tracks[0].Duration)
	}
}

func TestParseBareM3U(t *testing.T) {
	input := "/music/track1.flac\n/music/track2.mp3\n"
	tracks, err := m3u.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(tracks))
	}
}

func TestParseWindowsLineEndings(t *testing.T) {
	input := "#EXTM3U\r\n/music/track.flac\r\n"
	tracks, err := m3u.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
}
