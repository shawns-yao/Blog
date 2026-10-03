package music

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	domainmusic "github.com/shawns-yao/shawn-blog/server/internal/domain/music"
)

type libraryTestRepository struct {
	domainmusic.Repository
	entries []domainmusic.Entry
	user    int64
	removed string
}

func (r *libraryTestRepository) Favorites(_ context.Context, user int64, _, _ int) ([]domainmusic.Entry, error) {
	r.user = user
	return r.entries, nil
}
func (r *libraryTestRepository) SetFavorite(_ context.Context, user int64, entry domainmusic.Entry, add bool) error {
	r.user = user
	if !add {
		r.removed = entry.SongID
	}
	return nil
}

func TestPlaylistNameValidation(t *testing.T) {
	for _, name := range []string{"", "   ", strings.Repeat("曲", 101)} {
		if _, err := playlistName(name); err == nil {
			t.Fatal("accepted invalid name")
		}
	}
	if name, err := playlistName("  我的歌单  "); err != nil || name != "我的歌单" {
		t.Fatal("valid name was not normalized")
	}
}
func TestRemoveFavoriteDoesNotRequireUpstream(t *testing.T) {
	repo := &libraryTestRepository{}
	s := NewLibraryService(repo, &Service{})
	if err := s.SetFavorite(context.Background(), 7, "missing-song", false); err != nil {
		t.Fatal(err)
	}
	if repo.user != 7 || repo.removed != "missing-song" {
		t.Fatal("wrong owner or target")
	}
}
func TestHydrateDistinguishesMissingFromUpstreamFailure(t *testing.T) {
	for _, test := range []struct {
		code        int
		wantMissing bool
	}{{70, true}, {40, false}} {
		t.Run(strconv.Itoa(test.code), func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(map[string]any{"subsonic-response": map[string]any{"status": "failed", "error": map[string]any{"code": test.code}}})
			}))
			defer upstream.Close()
			base, _ := url.Parse(upstream.URL)
			music := &Service{configured: true, baseURL: base, http: upstream.Client()}
			music.cfg.RequestTimeout = time.Second
			repo := &libraryTestRepository{entries: []domainmusic.Entry{{SongID: "missing", Snapshot: []byte(`{"id":"missing","title":"Saved","public":true,"coverArt":"secret"}`)}}}
			service := NewLibraryService(repo, music)
			page, err := service.Favorites(context.Background(), 1, 0, 30)
			if test.wantMissing {
				if err != nil || len(page.Songs) != 1 {
					t.Fatalf("missing record not preserved: %v", err)
				}
				if !page.Songs[0].Unavailable || page.Songs[0].Public || page.Songs[0].CoverArt != "" {
					t.Fatal("snapshot granted playback")
				}
			} else if !errors.Is(err, ErrUpstream) {
				t.Fatalf("upstream failure disguised as missing: %v", err)
			}
		})
	}
}
func TestLibraryRejectsUnboundedQueries(t *testing.T) {
	s := NewLibraryService(nil, nil)
	if _, err := s.Favorites(context.Background(), 1, 0, 100); err == nil {
		t.Fatal("unbounded page accepted")
	}
	if _, err := s.FavoriteIDs(context.Background(), 1, []string{"../private"}); err == nil {
		t.Fatal("invalid song id accepted")
	}
	if err := s.ReorderPlaylist(context.Background(), 1, 1, []string{"a", "a"}); err == nil {
		t.Fatal("duplicate order accepted")
	}
}
