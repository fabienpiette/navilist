package m3u

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// Track is a single entry parsed from an M3U file.
type Track struct {
	Path     string
	Title    string
	Artist   string
	Duration int // seconds; -1 if unknown
}

// Parse reads an M3U or EXTM3U playlist from r.
func Parse(r io.Reader) ([]Track, error) {
	var tracks []Track
	var pending *Track

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 4096), 1024*1024) // allow up to 1 MiB per line
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		line = strings.TrimSpace(line)

		if line == "" || line == "#EXTM3U" {
			continue
		}

		if rest, ok := strings.CutPrefix(line, "#EXTINF:"); ok {
			parts := strings.SplitN(rest, ",", 2)
			dur := -1
			if d, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
				dur = d
			}
			t := &Track{Duration: dur}
			if len(parts) == 2 {
				info := strings.TrimSpace(parts[1])
				if artist, title, found := strings.Cut(info, " - "); found {
					t.Artist = strings.TrimSpace(artist)
					t.Title = strings.TrimSpace(title)
				} else {
					t.Title = info
				}
			}
			pending = t
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		if pending != nil {
			pending.Path = line
			tracks = append(tracks, *pending)
			pending = nil
		} else {
			tracks = append(tracks, Track{Path: line, Duration: -1})
		}
	}
	return tracks, scanner.Err()
}

// WriteTrack is input for the Write function.
type WriteTrack struct {
	Path     string
	Title    string
	Artist   string
	Duration int
}

// Write renders tracks as an EXTM3U playlist to w.
func Write(w io.Writer, tracks []WriteTrack) error {
	bw := bufio.NewWriter(w)
	bw.WriteString("#EXTM3U\n")
	for _, t := range tracks {
		dur := strconv.Itoa(t.Duration)
		var info string
		if t.Artist != "" {
			info = t.Artist + " - " + t.Title
		} else {
			info = t.Title
		}
		bw.WriteString("#EXTINF:" + dur + "," + info + "\n")
		bw.WriteString(t.Path + "\n")
	}
	return bw.Flush()
}
