package music

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

const MaxCoverBytes int64 = 10 * 1024 * 1024
const MaxLyricsBytes int64 = 1024 * 1024

func AssetLimit(kind string) int64 {
	if kind == "cover" {
		return MaxCoverBytes
	}
	return MaxLyricsBytes
}

// 补充资源按歌曲编号独立保存，Navidrome 的虚拟路径不用于定位原始音频。
func (s *Service) assetPath(id, filename string, create bool) (string, error) {
	if !s.configured {
		return "", ErrUnavailable
	}
	if !validID(id) {
		return "", &InputError{"音乐编号无效"}
	}
	for _, dir := range []string{s.assetRoot, filepath.Join(s.assetRoot, id)} {
		if create {
			if err := os.Mkdir(dir, 0700); err != nil && !errors.Is(err, os.ErrExist) {
				return "", ErrAsset
			}
		}
		info, err := os.Lstat(dir)
		if errors.Is(err, os.ErrNotExist) && !create {
			return "", os.ErrNotExist
		}
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return "", ErrAsset
		}
	}
	path := filepath.Join(s.assetRoot, id, filename)
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) && !create {
		return "", os.ErrNotExist
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) || err == nil && !info.Mode().IsRegular() {
		return "", ErrAsset
	}
	return path, nil
}

func (s *Service) saveAsset(id, filename string, data []byte) error {
	s.uploadMu.Lock()
	defer s.uploadMu.Unlock()
	path, err := s.assetPath(id, filename, true)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".asset-*")
	if err != nil {
		return ErrAsset
	}
	defer os.Remove(file.Name())
	_, err = file.Write(data)
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil || closeErr != nil || os.Rename(file.Name(), path) != nil {
		return ErrAsset
	}
	return nil
}

func (s *Service) UploadAsset(ctx context.Context, id, kind, filename string, reader io.Reader) (AssetResult, error) {
	if !s.configured {
		return AssetResult{}, ErrUnavailable
	}
	if !validID(id) || kind != "cover" && kind != "lyrics" {
		return AssetResult{}, &InputError{"音乐资源参数无效"}
	}
	data, err := io.ReadAll(io.LimitReader(reader, AssetLimit(kind)+1))
	if err != nil || len(data) == 0 || int64(len(data)) > AssetLimit(kind) {
		return AssetResult{}, &InputError{"文件为空、无法读取或超过大小限制"}
	}
	storedName := "lyrics.json"
	if kind == "cover" {
		data, err = normalizeCover(filename, data)
		storedName = "cover.jpg"
	} else {
		var lyrics Lyrics
		lyrics, err = parseLyrics(filename, data)
		if err == nil {
			data, err = json.Marshal(LyricsResult{Lyrics: []Lyrics{lyrics}})
			if len(data) > 8*int(MaxLyricsBytes) {
				err = &InputError{"歌词展开后的内容过多，请缩减重复时间标签"}
			}
		}
	}
	if err != nil {
		return AssetResult{}, err
	}
	result, err := s.call(ctx, "getSong", url.Values{"id": {id}})
	if err != nil {
		return AssetResult{}, err
	}
	if result.Song.ID != id || !validID(result.Song.ID) {
		return AssetResult{}, ErrUpstream
	}
	if err = s.saveAsset(result.Song.ID, storedName, data); err != nil {
		return AssetResult{}, err
	}
	return AssetResult{SongID: result.Song.ID, Kind: kind}, nil
}

func normalizeCover(filename string, data []byte) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png") || (ext != ".jpg" && ext != ".jpeg" && ext != ".png") || (format == "png") != (ext == ".png") || config.Width < 1 || config.Height < 1 || config.Width > 4096 || config.Height > 4096 || config.Width*config.Height > 16*1024*1024 {
		return nil, &InputError{"请选择有效的 JPG 或 PNG 封面，尺寸不超过 4096 × 4096"}
	}
	cover, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, &InputError{"封面文件损坏，无法读取"}
	}
	var output bytes.Buffer
	if jpeg.Encode(&output, imaging.Fit(cover, 720, 720, imaging.Lanczos), &jpeg.Options{Quality: 90}) != nil {
		return nil, ErrAsset
	}
	return output.Bytes(), nil
}

func (s *Service) enrichCover(song *Song) {
	path, err := s.assetPath(song.ID, "cover.jpg", false)
	if err == nil {
		if info, err := os.Stat(path); err == nil {
			song.CoverArt = "local-" + song.ID + "-" + strconv.FormatInt(info.ModTime().UnixNano(), 10)
		}
	}
}

func (s *Service) assetCover(id string) (*http.Response, error) {
	value := strings.TrimPrefix(id, "local-")
	index := strings.LastIndex(value, "-")
	if index <= 0 {
		return nil, &InputError{"封面编号无效"}
	}
	path, err := s.assetPath(value[:index], "cover.jpg", false)
	if err != nil {
		return nil, ErrAsset
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, ErrAsset
	}
	info, err := file.Stat()
	if err != nil || value[index+1:] != strconv.FormatInt(info.ModTime().UnixNano(), 10) || info.Size() > MaxCoverBytes {
		file.Close()
		return nil, ErrAsset
	}
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"image/jpeg"}, "Content-Length": {strconv.FormatInt(info.Size(), 10)}}, Body: file, ContentLength: info.Size()}, nil
}
