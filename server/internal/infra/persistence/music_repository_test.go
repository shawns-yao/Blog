package persistence

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	domainmusic "github.com/shawns-yao/shawn-blog/server/internal/domain/music"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// This test requires an explicitly supplied disposable database. All DDL is rolled back.
func TestMusicRepositoryOwnershipAndOrdering(t *testing.T) {
	dsn := os.Getenv("MUSIC_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MUSIC_TEST_DATABASE_URL to a disposable PostgreSQL database")
	}
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), &gorm.Config{})
	if err != nil {
		t.Fatal("cannot connect to music test database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	schema := "music_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	for _, statement := range []string{
		"CREATE SCHEMA " + schema,
		"SET LOCAL search_path TO " + schema,
		"CREATE TABLE app_user (id BIGINT PRIMARY KEY, deleted_at TIMESTAMPTZ)",
		"INSERT INTO app_user(id) VALUES (1), (2)",
	} {
		if err := tx.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	migration, err := os.ReadFile("../../../migrations/0079_add_music_library.sql")
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Exec(strings.Split(string(migration), "-- +goose Down")[0]).Error; err != nil {
		t.Fatal(err)
	}
	repo := NewMusicRepository(tx)
	ctx := context.Background()
	entry := func(id string) domainmusic.Entry {
		return domainmusic.Entry{SongID: id, Snapshot: []byte(`{"title":"Test"}`)}
	}
	for i := 0; i < 2; i++ {
		if err := repo.SetFavorite(ctx, 1, entry("a"), true); err != nil {
			t.Fatal(err)
		}
	}
	favorites, err := repo.Favorites(ctx, 1, 0, 31)
	if err != nil || len(favorites) != 1 {
		t.Fatalf("favorite is not idempotent: %v", err)
	}
	other, err := repo.Favorites(ctx, 2, 0, 31)
	if err != nil || len(other) != 0 {
		t.Fatalf("favorites crossed accounts: %v", err)
	}
	list, err := repo.CreatePlaylist(ctx, 1, "First")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Playlist(ctx, 2, list.ID); !errors.Is(err, domainmusic.ErrPlaylistNotFound) {
		t.Fatal("other user read playlist")
	}
	if err := repo.SetPlaylistSong(ctx, 2, list.ID, entry("a"), true); !errors.Is(err, domainmusic.ErrPlaylistNotFound) {
		t.Fatal("other user changed playlist")
	}
	for _, id := range []string{"a", "b", "a"} {
		if err := repo.SetPlaylistSong(ctx, 1, list.ID, entry(id), true); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.ReorderPlaylist(ctx, 1, list.ID, []string{"b", "a"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReorderPlaylist(ctx, 1, list.ID, []string{"b", "a"}); !errors.Is(err, domainmusic.ErrInvalidOrder) {
		t.Fatal("stale reorder was accepted")
	}
	items, err := repo.PlaylistSongs(ctx, 1, list.ID, 0, 31)
	if err != nil || len(items) != 2 || items[0].SongID != "b" {
		t.Fatalf("wrong playlist order: %v", err)
	}
	if err := repo.DeletePlaylist(ctx, 2, list.ID); !errors.Is(err, domainmusic.ErrPlaylistNotFound) {
		t.Fatal("other user deleted playlist")
	}
	if err := repo.DeletePlaylist(ctx, 1, list.ID); err != nil {
		t.Fatal(err)
	}
	favorites, err = repo.Favorites(ctx, 1, 0, 31)
	if err != nil || len(favorites) != 1 {
		t.Fatal("deleting playlist affected favorites")
	}
}
