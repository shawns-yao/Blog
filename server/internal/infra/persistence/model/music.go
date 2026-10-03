package model

import "time"

type MusicFavorite struct {
	UserID    int64     `gorm:"column:user_id;primaryKey"`
	SongID    string    `gorm:"column:song_id;primaryKey"`
	Snapshot  []byte    `gorm:"column:snapshot;type:jsonb"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (MusicFavorite) TableName() string { return "music_favorite" }

type MusicPlaylist struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	UserID    int64     `gorm:"column:user_id"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (MusicPlaylist) TableName() string { return "music_playlist" }

type MusicPlaylistSong struct {
	PlaylistID int64  `gorm:"column:playlist_id;primaryKey"`
	SongID     string `gorm:"column:song_id;primaryKey"`
	SortOrder  int    `gorm:"column:sort_order"`
	Snapshot   []byte `gorm:"column:snapshot;type:jsonb"`
}

func (MusicPlaylistSong) TableName() string { return "music_playlist_song" }
