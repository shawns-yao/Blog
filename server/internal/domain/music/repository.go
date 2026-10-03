package music

import (
	"context"
	"errors"
	"time"
)

var ErrPlaylistNotFound = errors.New("歌单不存在")
var ErrInvalidOrder = errors.New("歌单顺序已变化，请刷新后重试")
var ErrLibraryLimit = errors.New("歌单数量或歌曲数量已达到上限")

type Entry struct {
	SongID   string
	Snapshot []byte
}

type Playlist struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	SongCount int64     `json:"songCount"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Repository interface {
	Favorites(context.Context, int64, int, int) ([]Entry, error)
	FavoriteIDs(context.Context, int64, []string) ([]string, error)
	SetFavorite(context.Context, int64, Entry, bool) error
	Playlists(context.Context, int64, int, int) ([]Playlist, error)
	CreatePlaylist(context.Context, int64, string) (Playlist, error)
	Playlist(context.Context, int64, int64) (Playlist, error)
	RenamePlaylist(context.Context, int64, int64, string) error
	DeletePlaylist(context.Context, int64, int64) error
	PlaylistSongs(context.Context, int64, int64, int, int) ([]Entry, error)
	SetPlaylistSong(context.Context, int64, int64, Entry, bool) error
	ReorderPlaylist(context.Context, int64, int64, []string) error
}
