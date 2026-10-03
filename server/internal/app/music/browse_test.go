package music

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/shawns-yao/shawn-blog/server/internal/config"
)

func TestBrowseHidesPrivateCoverButKeepsSong(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"subsonic-response": map[string]any{
			"status": "ok", "searchResult3": map[string]any{"song": []Song{
				{ID: "private", Title: "Private", CoverArt: "secret"},
				{ID: "public", Title: "Public", CoverArt: "shared"},
			}},
		}})
	}))
	defer upstream.Close()
	s := NewService(config.MusicConfig{})
	s.configured = true
	s.baseURL, _ = url.Parse(upstream.URL)
	s.cfg.RequestTimeout = time.Second
	s.publications = map[string]publication{"public": {Song: Song{ID: "public", Public: true}}}
	catalog, err := s.Browse(context.Background(), "", 0, 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Songs) != 2 {
		t.Fatalf("want both songs, got %d", len(catalog.Songs))
	}
	if catalog.Songs[0].CoverArt != "" || catalog.Songs[0].Public {
		t.Fatal("private metadata exposed a cover or playback grant")
	}
	if !catalog.Songs[1].Public || catalog.Songs[1].CoverArt != "shared" {
		t.Fatal("public song lost its availability")
	}
}
