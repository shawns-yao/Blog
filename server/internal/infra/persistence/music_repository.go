package persistence

import (
	"context"
	"errors"
	"time"

	domainmusic "github.com/shawns-yao/shawn-blog/server/internal/domain/music"
	"github.com/shawns-yao/shawn-blog/server/internal/infra/persistence/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MusicRepository struct{ db *gorm.DB }

func NewMusicRepository(db *gorm.DB) *MusicRepository { return &MusicRepository{db: db} }

func (r *MusicRepository) Favorites(ctx context.Context, user int64, offset, limit int) ([]domainmusic.Entry, error) {
	var rows []model.MusicFavorite
	err := r.db.WithContext(ctx).Where("user_id = ?", user).Order("created_at DESC, song_id").Offset(offset).Limit(limit).Find(&rows).Error
	items := make([]domainmusic.Entry, 0, len(rows))
	for _, row := range rows {
		items = append(items, domainmusic.Entry{SongID: row.SongID, Snapshot: row.Snapshot})
	}
	return items, err
}
func (r *MusicRepository) FavoriteIDs(ctx context.Context, user int64, ids []string) ([]string, error) {
	items := []string{}
	if len(ids) == 0 {
		return items, nil
	}
	err := r.db.WithContext(ctx).Model(&model.MusicFavorite{}).Where("user_id = ? AND song_id IN ?", user, ids).Pluck("song_id", &items).Error
	return items, err
}
func (r *MusicRepository) SetFavorite(ctx context.Context, user int64, entry domainmusic.Entry, add bool) error {
	q := r.db.WithContext(ctx)
	if !add {
		return q.Where("user_id = ? AND song_id = ?", user, entry.SongID).Delete(&model.MusicFavorite{}).Error
	}
	row := model.MusicFavorite{UserID: user, SongID: entry.SongID, Snapshot: entry.Snapshot}
	return q.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}
func playlistQuery(db *gorm.DB, user int64) *gorm.DB {
	return db.Model(&model.MusicPlaylist{}).Select("music_playlist.id, music_playlist.name, music_playlist.updated_at, (SELECT count(*) FROM music_playlist_song WHERE playlist_id = music_playlist.id) AS song_count").Where("music_playlist.user_id = ?", user)
}
func (r *MusicRepository) Playlists(ctx context.Context, user int64, offset, limit int) ([]domainmusic.Playlist, error) {
	items := []domainmusic.Playlist{}
	err := playlistQuery(r.db.WithContext(ctx), user).Order("music_playlist.id DESC").Offset(offset).Limit(limit).Scan(&items).Error
	return items, err
}
func (r *MusicRepository) Playlist(ctx context.Context, user, id int64) (domainmusic.Playlist, error) {
	var item domainmusic.Playlist
	result := playlistQuery(r.db.WithContext(ctx), user).Where("music_playlist.id = ?", id).Scan(&item)
	if result.Error != nil {
		return item, result.Error
	}
	if result.RowsAffected == 0 {
		return item, domainmusic.ErrPlaylistNotFound
	}
	return item, nil
}
func (r *MusicRepository) CreatePlaylist(ctx context.Context, user int64, name string) (domainmusic.Playlist, error) {
	row := model.MusicPlaylist{UserID: user, Name: name}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var owner model.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&owner, user).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.MusicPlaylist{}).Where("user_id = ?", user).Count(&count).Error; err != nil {
			return err
		}
		if count >= 100 {
			return domainmusic.ErrLibraryLimit
		}
		return tx.Create(&row).Error
	})
	return domainmusic.Playlist{ID: row.ID, Name: row.Name, UpdatedAt: row.UpdatedAt}, err
}
func (r *MusicRepository) RenamePlaylist(ctx context.Context, user, id int64, name string) error {
	result := r.db.WithContext(ctx).Model(&model.MusicPlaylist{}).Where("id = ? AND user_id = ?", id, user).Updates(map[string]any{"name": name, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainmusic.ErrPlaylistNotFound
	}
	return nil
}
func (r *MusicRepository) DeletePlaylist(ctx context.Context, user, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, user).Delete(&model.MusicPlaylist{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainmusic.ErrPlaylistNotFound
	}
	return nil
}
func (r *MusicRepository) PlaylistSongs(ctx context.Context, user, id int64, offset, limit int) ([]domainmusic.Entry, error) {
	if _, err := r.Playlist(ctx, user, id); err != nil {
		return nil, err
	}
	var rows []model.MusicPlaylistSong
	err := r.db.WithContext(ctx).Where("playlist_id = ?", id).Order("sort_order, song_id").Offset(offset).Limit(limit).Find(&rows).Error
	items := make([]domainmusic.Entry, 0, len(rows))
	for _, row := range rows {
		items = append(items, domainmusic.Entry{SongID: row.SongID, Snapshot: row.Snapshot})
	}
	return items, err
}
func lockMusicPlaylist(tx *gorm.DB, user, id int64) error {
	var row model.MusicPlaylist
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND user_id = ?", id, user).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainmusic.ErrPlaylistNotFound
	}
	return err
}
func (r *MusicRepository) SetPlaylistSong(ctx context.Context, user, id int64, entry domainmusic.Entry, add bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockMusicPlaylist(tx, user, id); err != nil {
			return err
		}
		if !add {
			if err := tx.Where("playlist_id = ? AND song_id = ?", id, entry.SongID).Delete(&model.MusicPlaylistSong{}).Error; err != nil {
				return err
			}
		} else {
			var rows []model.MusicPlaylistSong
			if err := tx.Where("playlist_id = ?", id).Order("sort_order DESC").Find(&rows).Error; err != nil {
				return err
			}
			for _, row := range rows {
				if row.SongID == entry.SongID {
					return nil
				}
			}
			if len(rows) >= 500 {
				return domainmusic.ErrLibraryLimit
			}
			next := 0
			if len(rows) > 0 {
				next = rows[0].SortOrder + 1
			}
			row := model.MusicPlaylistSong{PlaylistID: id, SongID: entry.SongID, Snapshot: entry.Snapshot, SortOrder: next}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.MusicPlaylist{}).Where("id = ? AND user_id = ?", id, user).Update("updated_at", time.Now()).Error
	})
}

// ReorderPlaylist swaps adjacent songs, checking the expected pair under the owner lock.
func (r *MusicRepository) ReorderPlaylist(ctx context.Context, user, id int64, ids []string) error {
	if len(ids) != 2 || ids[0] == ids[1] {
		return domainmusic.ErrInvalidOrder
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockMusicPlaylist(tx, user, id); err != nil {
			return err
		}
		var rows []model.MusicPlaylistSong
		if err := tx.Where("playlist_id = ?", id).Order("sort_order, song_id").Find(&rows).Error; err != nil {
			return err
		}
		index := -1
		for i := 0; i+1 < len(rows); i++ {
			if rows[i].SongID == ids[1] && rows[i+1].SongID == ids[0] {
				index = i
				break
			}
		}
		if index < 0 {
			return domainmusic.ErrInvalidOrder
		}
		for i, songID := range ids {
			if err := tx.Model(&model.MusicPlaylistSong{}).Where("playlist_id = ? AND song_id = ?", id, songID).Update("sort_order", rows[index+i].SortOrder).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.MusicPlaylist{}).Where("id = ?", id).Update("updated_at", time.Now()).Error
	})
}
