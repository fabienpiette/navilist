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
	if tracks[0].Path != "/music/track.flac" {
		t.Errorf("wrong path: %s", tracks[0].Path)
	}
}

func TestParseMalformedDuration(t *testing.T) {
	input := "#EXTM3U\n#EXTINF:notanumber,Title Only\n/music/track.flac\n"
	tracks, err := m3u.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Duration != -1 {
		t.Errorf("expected duration -1, got %d", tracks[0].Duration)
	}
	if tracks[0].Title != "Title Only" {
		t.Errorf("expected title-only, got %q", tracks[0].Title)
	}
	if tracks[0].Artist != "" {
		t.Errorf("expected no artist, got %q", tracks[0].Artist)
	}
}

func TestParseEmpty(t *testing.T) {
	tracks, err := m3u.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 0 {
		t.Fatalf("expected 0 tracks, got %d", len(tracks))
	}
}

func TestWrite(t *testing.T) {
	var buf strings.Builder
	tracks := []m3u.WriteTrack{
		{Path: "/music/song.flac", Title: "Song", Artist: "Artist", Duration: 120},
	}
	if err := m3u.Write(&buf, tracks); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "#EXTM3U") {
		t.Error("missing #EXTM3U header")
	}
	if !strings.Contains(out, "#EXTINF:120,Artist - Song") {
		t.Errorf("missing EXTINF line, got: %s", out)
	}
	if !strings.Contains(out, "/music/song.flac") {
		t.Error("missing path")
	}
}
