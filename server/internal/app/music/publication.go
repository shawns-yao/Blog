package music

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxPublicationBytes = 8 * 1024 * 1024

// 发布记录属于音乐访问权限数据；取消公开保留记录，不删除原始音乐。
type publication struct {
	Song Song `json:"song"`
}

type publicationFile struct {
	Version int                    `json:"version"`
	Songs   map[string]publication `json:"songs"`
}

func (s *Service) loadPublications() {
	s.publications = make(map[string]publication)
	file, err := os.Open(s.publicationPath)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		s.publicationErr = ErrPublication
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxPublicationBytes+1))
	var saved publicationFile
	if err != nil || len(data) > maxPublicationBytes || json.Unmarshal(data, &saved) != nil || saved.Version != 1 || saved.Songs == nil {
		s.publicationErr = ErrPublication
		return
	}
	for id, entry := range saved.Songs {
		if !validID(id) || entry.Song.ID != id || (entry.Song.CoverArt != "" && !validID(entry.Song.CoverArt)) || (entry.Song.AlbumID != "" && !validID(entry.Song.AlbumID)) {
			s.publicationErr = ErrPublication
			return
		}
	}
	s.publications = saved.Songs
}

func (s *Service) markPublic(songs []Song) error {
	s.publicationMu.RLock()
	defer s.publicationMu.RUnlock()
	if s.publicationErr != nil {
		return s.publicationErr
	}
	for i := range songs {
		songs[i].Public = s.publications[songs[i].ID].Song.Public
		s.enrichCover(&songs[i])
	}
	return nil
}

func (s *Service) SetPublic(ctx context.Context, id string, public bool) (Song, error) {
	if !validID(id) {
		return Song{}, &InputError{"音乐编号无效"}
	}
	result, err := s.call(ctx, "getSong", url.Values{"id": {id}})
	if err != nil {
		return Song{}, err
	}
	song := result.Song
	if song.ID != id || (song.CoverArt != "" && !validID(song.CoverArt)) || (song.AlbumID != "" && !validID(song.AlbumID)) {
		return Song{}, ErrUpstream
	}
	song.Public = public
	s.publicationMu.Lock()
	defer s.publicationMu.Unlock()
	if s.publicationErr != nil {
		return Song{}, s.publicationErr
	}
	next := make(map[string]publication, len(s.publications)+1)
	for key, entry := range s.publications {
		next[key] = entry
	}
	// Fiber 请求参数的底层缓冲区会复用，持久记录使用上游 JSON 解码后的独立编号。
	next[song.ID] = publication{Song: song}
	data, err := json.Marshal(publicationFile{Version: 1, Songs: next})
	if err != nil || len(data) > maxPublicationBytes {
		return Song{}, ErrPublication
	}
	if os.MkdirAll(filepath.Dir(s.publicationPath), 0700) != nil {
		return Song{}, ErrPublication
	}
	file, err := os.CreateTemp(filepath.Dir(s.publicationPath), ".public-catalog-*")
	if err != nil {
		return Song{}, ErrPublication
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil || closeErr != nil || os.Rename(name, s.publicationPath) != nil {
		return Song{}, ErrPublication
	}
	s.publications = next
	return song, nil
}

func (s *Service) publishedSongs() ([]Song, error) {
	if !s.configured {
		return nil, ErrUnavailable
	}
	s.publicationMu.RLock()
	defer s.publicationMu.RUnlock()
	if s.publicationErr != nil {
		return nil, s.publicationErr
	}
	songs := []Song{}
	for _, entry := range s.publications {
		if entry.Song.Public {
			song := entry.Song
			s.enrichCover(&song)
			songs = append(songs, song)
		}
	}
	sort.Slice(songs, func(i, j int) bool {
		if songs[i].Title != songs[j].Title {
			return songs[i].Title < songs[j].Title
		}
		return songs[i].ID < songs[j].ID
	})
	return songs, nil
}

func (s *Service) PublicCatalog(query string, offset, limit int) (Catalog, error) {
	if offset < 0 || limit < 1 || limit > 100 || len(query) > 300 {
		return Catalog{}, &InputError{"曲库查询参数无效"}
	}
	songs, err := s.publishedSongs()
	if err != nil {
		return Catalog{}, err
	}
	query = strings.ToLower(strings.TrimSpace(query))
	filtered := []Song{}
	for _, song := range songs {
		if query == "" || strings.Contains(strings.ToLower(song.Title+"\n"+song.Artist+"\n"+song.Album), query) {
			filtered = append(filtered, song)
		}
	}
	albums := albumsFromSongs(filtered)
	start := min(offset, len(filtered))
	end := start + min(limit, len(filtered)-start)
	return Catalog{Songs: filtered[start:end], Albums: albums, HasMore: end < len(filtered)}, nil
}

func albumsFromSongs(songs []Song) []Album {
	byID := map[string]Album{}
	for _, song := range songs {
		if song.AlbumID == "" {
			continue
		}
		album := byID[song.AlbumID]
		if album.ID == "" {
			album = Album{ID: song.AlbumID, Name: song.Album, Artist: song.Artist, CoverArt: song.CoverArt}
		}
		album.SongCount++
		byID[song.AlbumID] = album
	}
	albums := []Album{}
	for _, album := range byID {
		albums = append(albums, album)
	}
	sort.Slice(albums, func(i, j int) bool {
		if albums[i].Name != albums[j].Name {
			return albums[i].Name < albums[j].Name
		}
		return albums[i].ID < albums[j].ID
	})
	return albums
}

func (s *Service) PublicAlbum(id string) (Album, error) {
	songs, err := s.publishedSongs()
	if err != nil {
		return Album{}, err
	}
	selected := []Song{}
	for _, song := range songs {
		if song.AlbumID == id {
			selected = append(selected, song)
		}
	}
	if !validID(id) || len(selected) == 0 {
		return Album{}, ErrNotPublic
	}
	album := albumsFromSongs(selected)[0]
	album.Songs = selected
	return album, nil
}

func (s *Service) PublicBinary(ctx context.Context, id, byteRange, ifRange string, cover bool) (*http.Response, error) {
	songs, err := s.publishedSongs()
	if err != nil {
		return nil, err
	}
	allowed := false
	for _, song := range songs {
		if (!cover && song.ID == id) || (cover && song.CoverArt == id && id != "") {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, ErrNotPublic
	}
	return s.Binary(ctx, id, byteRange, ifRange, cover)
}
