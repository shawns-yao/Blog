package music

import "errors"

var ErrSongNotFound = errors.New("歌曲已不在曲库中")
var ErrUnavailable = errors.New("音乐服务未启用或配置不完整")
var ErrUpstream = errors.New("音乐服务暂时无法访问，请稍后重试")
var ErrNotPublic = errors.New("音乐未公开")
var ErrPublication = errors.New("公开试听名单无法读取或保存，请检查音乐目录权限与名单文件")
var ErrAsset = errors.New("音乐封面或歌词无法读取或保存，请检查音乐目录权限")

type InputError struct{ Message string }

func (e *InputError) Error() string { return e.Message }

type Song struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Album       string  `json:"album"`
	AlbumID     string  `json:"albumId"`
	CoverArt    string  `json:"coverArt,omitempty"`
	Duration    float64 `json:"duration"`
	Track       int     `json:"track"`
	Year        int     `json:"year"`
	Public      bool    `json:"public"`
	Unavailable bool    `json:"unavailable,omitempty"`
}

type Album struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Artist    string `json:"artist"`
	CoverArt  string `json:"coverArt,omitempty"`
	SongCount int    `json:"songCount"`
	Songs     []Song `json:"song,omitempty"`
}

type Catalog struct {
	Songs   []Song  `json:"songs"`
	Albums  []Album `json:"albums"`
	HasMore bool    `json:"hasMore"`
}

type ScanStatus struct {
	Scanning bool  `json:"scanning"`
	Count    int64 `json:"count"`
}

type Status struct {
	Enabled        bool        `json:"enabled"`
	Configured     bool        `json:"configured"`
	Available      bool        `json:"available"`
	Message        string      `json:"message,omitempty"`
	Scan           *ScanStatus `json:"scan,omitempty"`
	MaxUploadBytes int64       `json:"maxUploadBytes"`
	MaxBitRate     int         `json:"maxBitRate"`
}

type UploadResult struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Size      int64  `json:"size"`
	Duplicate bool   `json:"duplicate"`
}

type LyricLine struct {
	Start *int64 `json:"start,omitempty"`
	Value string `json:"value"`
}

type Lyrics struct {
	DisplayArtist string      `json:"displayArtist,omitempty"`
	DisplayTitle  string      `json:"displayTitle,omitempty"`
	Lang          string      `json:"lang"`
	Synced        bool        `json:"synced"`
	Offset        int64       `json:"offset"`
	Line          []LyricLine `json:"line"`
}

type LyricsResult struct {
	Lyrics []Lyrics `json:"lyrics"`
}

type AssetResult struct {
	SongID string `json:"songId"`
	Kind   string `json:"kind"`
}
