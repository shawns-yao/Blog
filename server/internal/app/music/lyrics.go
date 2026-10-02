package music

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

var lyricTimestamp = regexp.MustCompile(`^\[(\d{1,5}):([0-5]\d)(?:\.(\d{1,3}))?\]`)
var lyricMetadata = regexp.MustCompile(`^\[(ar|ti|al|by|re|ve|length):.*\]$`)

func parseLyrics(filename string, data []byte) (Lyrics, error) {
	invalid := &InputError{"请选择有效的 UTF-8 歌词文件：LRC 需要时间标签，TXT 用于纯文本歌词"}
	ext := strings.ToLower(filepath.Ext(filename))
	if (ext != ".lrc" && ext != ".txt") || !utf8.Valid(data) || strings.ContainsRune(string(data), '\x00') {
		return Lyrics{}, invalid
	}
	lyrics := Lyrics{Lang: "und", Synced: ext == ".lrc", Line: []LyricLine{}}
	text := strings.TrimPrefix(string(data), "\ufeff")
	total := 0
	for _, value := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		value = strings.TrimSpace(value)
		if len(value) > 16000 {
			return Lyrics{}, invalid
		}
		if value == "" {
			continue
		}
		if !lyrics.Synced {
			lyrics.Line = append(lyrics.Line, LyricLine{Value: value})
		} else if strings.HasPrefix(value, "[offset:") && strings.HasSuffix(value, "]") {
			offset, err := strconv.ParseInt(strings.TrimSpace(value[8:len(value)-1]), 10, 64)
			if err != nil || offset < -86400000 || offset > 86400000 {
				return Lyrics{}, invalid
			}
			lyrics.Offset = offset
		} else if !lyricMetadata.MatchString(value) {
			starts := []int64{}
			for {
				match := lyricTimestamp.FindStringSubmatch(value)
				if match == nil {
					break
				}
				minutes, _ := strconv.ParseInt(match[1], 10, 64)
				seconds, _ := strconv.ParseInt(match[2], 10, 64)
				fraction, _ := strconv.ParseInt(match[3]+strings.Repeat("0", 3-len(match[3])), 10, 64)
				starts = append(starts, minutes*60000+seconds*1000+fraction)
				value = value[len(match[0]):]
			}
			if len(starts) == 0 {
				return Lyrics{}, invalid
			}
			for _, start := range starts {
				total += len(value)
				if total > 4*int(MaxLyricsBytes) {
					return Lyrics{}, invalid
				}
				lyrics.Line = append(lyrics.Line, LyricLine{Start: &start, Value: strings.TrimSpace(value)})
			}
		}
		if len(lyrics.Line) > 10000 {
			return Lyrics{}, invalid
		}
	}
	if len(lyrics.Line) == 0 {
		return Lyrics{}, invalid
	}
	if lyrics.Synced {
		sort.SliceStable(lyrics.Line, func(i, j int) bool { return *lyrics.Line[i].Start < *lyrics.Line[j].Start })
	}
	return lyrics, nil
}

func (s *Service) Lyrics(ctx context.Context, id string) (LyricsResult, error) {
	if !validID(id) {
		return LyricsResult{}, &InputError{"音乐编号无效"}
	}
	path, err := s.assetPath(id, "lyrics.json", false)
	if err == nil {
		file, openErr := os.Open(path)
		if openErr != nil {
			return LyricsResult{}, ErrAsset
		}
		defer file.Close()
		var lyrics LyricsResult
		// 多个时间标签会展开为多行，规范化文件的上限高于原始文本。
		if json.NewDecoder(io.LimitReader(file, 8*MaxLyricsBytes)).Decode(&lyrics) != nil || len(lyrics.Lyrics) == 0 {
			return LyricsResult{}, ErrAsset
		}
		return lyrics, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return LyricsResult{}, err
	}
	result, err := s.call(ctx, "getLyricsBySongId", url.Values{"id": {id}})
	if err != nil {
		return LyricsResult{}, err
	}
	lyrics := result.LyricsList.Lyrics
	if lyrics == nil {
		lyrics = []Lyrics{}
	}
	return LyricsResult{Lyrics: lyrics}, nil
}

func (s *Service) PublicLyrics(ctx context.Context, id string) (LyricsResult, error) {
	if !s.configured {
		return LyricsResult{}, ErrUnavailable
	}
	s.publicationMu.RLock()
	allowed := s.publications[id].Song.Public
	err := s.publicationErr
	s.publicationMu.RUnlock()
	if err != nil {
		return LyricsResult{}, err
	}
	if !allowed {
		return LyricsResult{}, ErrNotPublic
	}
	return s.Lyrics(ctx, id)
}
