package config

import "time"

type MusicConfig struct {
	Enabled        bool
	NavidromeURL   string
	Username       string
	Password       string
	LibraryDir     string
	StagingDir     string
	AllowedUserIDs []string
	MaxUploadBytes int64
	MaxBitRate     int
	RequestTimeout time.Duration
	ProbeBinary    string
}

func loadMusic() MusicConfig {
	return MusicConfig{
		Enabled:        getEnvAsBool("MUSIC_ENABLED", false),
		NavidromeURL:   getEnv("MUSIC_NAVIDROME_URL", "http://navidrome:4533"),
		Username:       getEnv("MUSIC_NAVIDROME_USER", ""),
		Password:       getEnv("MUSIC_NAVIDROME_PASSWORD", ""),
		LibraryDir:     getEnv("MUSIC_LIBRARY_DIR", "storage/music/library"),
		StagingDir:     getEnv("MUSIC_STAGING_DIR", "storage/music/staging"),
		AllowedUserIDs: getEnvAsSlice("MUSIC_ALLOWED_USER_IDS", nil),
		MaxUploadBytes: getEnvAsInt64("MUSIC_MAX_UPLOAD_BYTES", 100*1024*1024),
		MaxBitRate:     int(getEnvAsInt64("MUSIC_MAX_BIT_RATE", 192)),
		RequestTimeout: getEnvAsDuration("MUSIC_REQUEST_TIMEOUT", 15*time.Second),
		ProbeBinary:    getEnv("MUSIC_FFPROBE_BINARY", "ffprobe"),
	}
}
