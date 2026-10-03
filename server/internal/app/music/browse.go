package music

import (
	"context"
	"net/url"
)

// Browse exposes metadata only; media endpoints retain their own access checks.
func (s *Service) Browse(ctx context.Context, query string, offset, limit int) (Catalog, error) {
	catalog, err := s.Catalog(ctx, query, offset, limit)
	if err != nil {
		return Catalog{}, err
	}
	catalog.Songs = browseSongs(catalog.Songs)
	for i := range catalog.Albums {
		catalog.Albums[i].CoverArt = ""
		catalog.Albums[i].Songs = nil
	}
	return catalog, nil
}

func (s *Service) BrowseAlbum(ctx context.Context, id string) (Album, error) {
	album, err := s.Album(ctx, id)
	if err != nil {
		return Album{}, err
	}
	album.CoverArt = ""
	album.Songs = browseSongs(album.Songs)
	return album, nil
}

func browseSongs(songs []Song) []Song {
	for i := range songs {
		if !songs[i].Public {
			songs[i].CoverArt = ""
		}
	}
	return songs
}

func (s *Service) GetSong(ctx context.Context, id string) (Song, error) {
	if !validID(id) {
		return Song{}, &InputError{"音乐编号无效"}
	}
	result, err := s.call(ctx, "getSong", url.Values{"id": {id}})
	if err != nil {
		return Song{}, err
	}
	if result.Song.ID != id {
		return Song{}, ErrUpstream
	}
	songs := []Song{result.Song}
	if err := s.markPublic(songs); err != nil {
		return Song{}, err
	}
	return songs[0], nil
}
