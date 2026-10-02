package music

import (
	"context"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shawns-yao/shawn-blog/server/internal/config"
)

type Service struct {
	cfg             config.MusicConfig
	baseURL         *url.URL
	http            *http.Client
	configured      bool
	uploadMu        sync.Mutex
	publicationMu   sync.RWMutex
	publications    map[string]publication
	publicationPath string
	publicationErr  error
	assetRoot       string
}

func NewService(cfg config.MusicConfig) *Service {
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 15 * time.Second
	}
	if cfg.MaxUploadBytes <= 0 {
		cfg.MaxUploadBytes = 100 * 1024 * 1024
	}
	if cfg.MaxBitRate < 64 || cfg.MaxBitRate > 320 {
		cfg.MaxBitRate = 192
	}
	s := &Service{cfg: cfg, http: newHTTPClient()}
	base, err := url.Parse(cfg.NavidromeURL)
	if err == nil && (base.Scheme == "http" || base.Scheme == "https") && base.Host != "" && base.User == nil && base.RawQuery == "" && base.Fragment == "" {
		s.baseURL = base
		library, libraryErr := filepath.Abs(cfg.LibraryDir)
		staging, stagingErr := filepath.Abs(cfg.StagingDir)
		rel, relErr := filepath.Rel(library, staging)
		separate := relErr == nil && (rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)))
		s.configured = cfg.Enabled && cfg.Username != "" && cfg.Password != "" && cfg.LibraryDir != "" && cfg.StagingDir != "" && libraryErr == nil && stagingErr == nil && separate
		if s.configured {
			s.assetRoot = filepath.Join(filepath.Dir(library), "assets")
			s.publicationPath = filepath.Join(filepath.Dir(library), "public-catalog.json")
			s.loadPublications()
		}
	}
	return s
}

func (s *Service) Allowed(userID int64, admin bool) bool {
	if admin {
		return true
	}
	for _, value := range s.cfg.AllowedUserIDs {
		if strings.TrimSpace(value) == strconv.FormatInt(userID, 10) {
			return true
		}
	}
	return false
}

func (s *Service) StatusLimit() int64 { return s.cfg.MaxUploadBytes }

func (s *Service) Status(ctx context.Context, detailed bool) Status {
	status := Status{Enabled: s.cfg.Enabled, Configured: s.configured, MaxUploadBytes: s.cfg.MaxUploadBytes, MaxBitRate: s.cfg.MaxBitRate}
	if !s.configured {
		status.Message = ErrUnavailable.Error()
		return status
	}
	if _, err := s.call(ctx, "ping", nil); err != nil {
		status.Message = err.Error()
		return status
	}
	status.Available = true
	if detailed {
		scan, err := s.Scan(ctx, false)
		if err != nil {
			status.Message = err.Error()
		} else {
			status.Scan = &scan
		}
	}
	return status
}

func (s *Service) Catalog(ctx context.Context, query string, offset, limit int) (Catalog, error) {
	if offset < 0 || limit < 1 || limit > 100 || len(query) > 300 {
		return Catalog{}, &InputError{"曲库查询参数无效"}
	}
	values := url.Values{"query": {query}, "songCount": {strconv.Itoa(limit + 1)}, "songOffset": {strconv.Itoa(offset)}, "artistCount": {"0"}, "albumCount": {"12"}}
	result, err := s.call(ctx, "search3", values)
	if err != nil {
		return Catalog{}, err
	}
	songs := result.Search.Songs
	hasMore := len(songs) > limit
	if hasMore {
		songs = songs[:limit]
	}
	if songs == nil {
		songs = []Song{}
	}
	albums := result.Search.Albums
	if albums == nil {
		albums = []Album{}
	}
	if err := s.markPublic(songs); err != nil {
		return Catalog{}, err
	}
	return Catalog{Songs: songs, Albums: albums, HasMore: hasMore}, nil
}

func (s *Service) Album(ctx context.Context, id string) (Album, error) {
	if !validID(id) {
		return Album{}, &InputError{"专辑编号无效"}
	}
	result, err := s.call(ctx, "getAlbum", url.Values{"id": {id}})
	if err == nil {
		err = s.markPublic(result.Album.Songs)
	}
	return result.Album, err
}

func (s *Service) Scan(ctx context.Context, start bool) (ScanStatus, error) {
	endpoint := "getScanStatus"
	if start {
		endpoint = "startScan"
	}
	result, err := s.call(ctx, endpoint, nil)
	return result.Scan, err
}

func validID(id string) bool {
	if id == "" || len(id) > 200 {
		return false
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// Binary 保留 Range 请求和响应头，音频不载入 Go 的内存缓冲区。
func (s *Service) Binary(ctx context.Context, id, byteRange, ifRange string, cover bool) (*http.Response, error) {
	if !s.configured {
		return nil, ErrUnavailable
	}
	if !validID(id) {
		return nil, &InputError{"音乐编号无效"}
	}
	if cover && strings.HasPrefix(id, "local-") {
		return s.assetCover(id)
	}
	endpoint := "stream"
	values := url.Values{"id": {id}, "maxBitRate": {strconv.Itoa(s.cfg.MaxBitRate)}, "format": {"mp3"}}
	if cover {
		endpoint = "getCoverArt"
		values = url.Values{"id": {id}, "size": {"360"}}
	}
	// 认证参数由 request 统一生成；二进制请求额外透传范围头。
	resp, err := s.requestWithRange(ctx, endpoint, values, byteRange, ifRange)
	if err != nil {
		return nil, err
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if resp.StatusCode != 200 && resp.StatusCode != 206 && resp.StatusCode != 416 || (resp.StatusCode != 416 && !(strings.HasPrefix(contentType, "audio/") || strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "application/octet-stream"))) {
		resp.Body.Close()
		return nil, ErrUpstream
	}
	return resp, nil
}
