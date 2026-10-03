package music

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	domainmusic "github.com/shawns-yao/shawn-blog/server/internal/domain/music"
)

type LibraryService struct {
	repo  domainmusic.Repository
	music *Service
}

func NewLibraryService(repo domainmusic.Repository, music *Service) *LibraryService {
	return &LibraryService{repo: repo, music: music}
}

type SongPage struct {
	Songs   []Song `json:"songs"`
	HasMore bool   `json:"hasMore"`
}

func libraryPage(user int64, offset, limit int) error {
	if user <= 0 || offset < 0 || offset > 100000 || limit < 1 || limit > 30 {
		return &InputError{"分页参数无效"}
	}
	return nil
}
func playlistName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 100 {
		return "", &InputError{"歌单名称需要 1 至 100 个字符"}
	}
	return name, nil
}
func (s *LibraryService) hydrate(ctx context.Context, entries []domainmusic.Entry, limit int) (SongPage, error) {
	page := SongPage{Songs: []Song{}, HasMore: len(entries) > limit}
	if page.HasMore {
		entries = entries[:limit]
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	for _, entry := range entries {
		song, err := s.music.GetSong(ctx, entry.SongID)
		if errors.Is(err, ErrSongNotFound) {
			if json.Unmarshal(entry.Snapshot, &song) != nil {
				return SongPage{}, errors.New("个人音乐记录无法读取")
			}
			song.ID, song.Public, song.CoverArt, song.Unavailable = entry.SongID, false, "", true
		} else if err != nil {
			return SongPage{}, err
		}
		if !song.Public {
			song.CoverArt = ""
		}
		page.Songs = append(page.Songs, song)
	}
	return page, nil
}
func (s *LibraryService) Favorites(ctx context.Context, user int64, offset, limit int) (SongPage, error) {
	if err := libraryPage(user, offset, limit); err != nil {
		return SongPage{}, err
	}
	entries, err := s.repo.Favorites(ctx, user, offset, limit+1)
	if err != nil {
		return SongPage{}, err
	}
	return s.hydrate(ctx, entries, limit)
}
func (s *LibraryService) FavoriteIDs(ctx context.Context, user int64, ids []string) ([]string, error) {
	if len(ids) > 100 {
		return nil, &InputError{"一次最多查询 100 首歌曲"}
	}
	for _, id := range ids {
		if !validID(id) {
			return nil, &InputError{"音乐编号无效"}
		}
	}
	return s.repo.FavoriteIDs(ctx, user, ids)
}
func (s *LibraryService) entry(ctx context.Context, id string, add bool) (domainmusic.Entry, error) {
	entry := domainmusic.Entry{SongID: id}
	if !validID(id) {
		return entry, &InputError{"音乐编号无效"}
	}
	if !add {
		return entry, nil
	}
	song, err := s.music.GetSong(ctx, id)
	if err != nil {
		return entry, err
	}
	song.CoverArt, song.Public = "", false
	entry.Snapshot, err = json.Marshal(song)
	return entry, err
}
func (s *LibraryService) SetFavorite(ctx context.Context, user int64, id string, add bool) error {
	entry, err := s.entry(ctx, id, add)
	if err != nil {
		return err
	}
	return s.repo.SetFavorite(ctx, user, entry, add)
}
func (s *LibraryService) Playlists(ctx context.Context, user int64, offset, limit int) ([]domainmusic.Playlist, error) {
	if err := libraryPage(user, offset, limit); err != nil {
		return nil, err
	}
	return s.repo.Playlists(ctx, user, offset, limit+1)
}
func (s *LibraryService) CreatePlaylist(ctx context.Context, user int64, name string) (domainmusic.Playlist, error) {
	name, err := playlistName(name)
	if err != nil {
		return domainmusic.Playlist{}, err
	}
	return s.repo.CreatePlaylist(ctx, user, name)
}
func (s *LibraryService) RenamePlaylist(ctx context.Context, user, id int64, name string) error {
	name, err := playlistName(name)
	if err != nil {
		return err
	}
	return s.repo.RenamePlaylist(ctx, user, id, name)
}
func (s *LibraryService) DeletePlaylist(ctx context.Context, user, id int64) error {
	return s.repo.DeletePlaylist(ctx, user, id)
}
func (s *LibraryService) Playlist(ctx context.Context, user, id int64, offset, limit int) (domainmusic.Playlist, SongPage, error) {
	if err := libraryPage(user, offset, limit); err != nil {
		return domainmusic.Playlist{}, SongPage{}, err
	}
	list, err := s.repo.Playlist(ctx, user, id)
	if err != nil {
		return list, SongPage{}, err
	}
	entries, err := s.repo.PlaylistSongs(ctx, user, id, offset, limit+1)
	if err != nil {
		return list, SongPage{}, err
	}
	page, err := s.hydrate(ctx, entries, limit)
	return list, page, err
}
func (s *LibraryService) SetPlaylistSong(ctx context.Context, user, id int64, songID string, add bool) error {
	if _, err := s.repo.Playlist(ctx, user, id); err != nil {
		return err
	}
	entry, err := s.entry(ctx, songID, add)
	if err != nil {
		return err
	}
	return s.repo.SetPlaylistSong(ctx, user, id, entry, add)
}
func (s *LibraryService) ReorderPlaylist(ctx context.Context, user, id int64, ids []string) error {
	if len(ids) != 2 || !validID(ids[0]) || !validID(ids[1]) || ids[0] == ids[1] {
		return domainmusic.ErrInvalidOrder
	}
	return s.repo.ReorderPlaylist(ctx, user, id, ids)
}
