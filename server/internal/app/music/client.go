package music

import (
	"context"
	"crypto/md5" // Subsonic 协议的令牌算法。
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type subsonicResponse struct {
	Status string `json:"status"`
	Error  struct {
		Code int `json:"code"`
	} `json:"error"`
	Search struct {
		Songs  []Song  `json:"song"`
		Albums []Album `json:"album"`
	} `json:"searchResult3"`
	Album      Album      `json:"album"`
	Song       Song       `json:"song"`
	Scan       ScanStatus `json:"scanStatus"`
	LyricsList struct {
		Lyrics []Lyrics `json:"structuredLyrics"`
	} `json:"lyricsList"`
}

func (s *Service) request(ctx context.Context, endpoint string, values url.Values) (*http.Response, error) {
	return s.requestWithRange(ctx, endpoint, values, "", "")
}

func (s *Service) requestWithRange(ctx context.Context, endpoint string, values url.Values, byteRange, ifRange string) (*http.Response, error) {
	if !s.configured {
		return nil, ErrUnavailable
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, ErrUpstream
	}
	saltString := hex.EncodeToString(salt)
	token := md5.Sum([]byte(s.cfg.Password + saltString))
	if values == nil {
		values = make(url.Values)
	}
	values.Set("u", s.cfg.Username)
	values.Set("s", saltString)
	values.Set("t", hex.EncodeToString(token[:]))
	values.Set("v", "1.16.1")
	values.Set("c", "grtblog")
	values.Set("f", "json")
	target := *s.baseURL
	target.Path = strings.TrimRight(target.Path, "/") + "/rest/" + endpoint
	target.RawQuery = values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, ErrUpstream
	}
	if byteRange != "" {
		req.Header.Set("Range", byteRange)
	}
	if ifRange != "" {
		req.Header.Set("If-Range", ifRange)
	}
	resp, err := s.http.Do(req)
	// 不向调用方返回携带认证查询参数的 HTTP 错误。
	if err != nil {
		return nil, ErrUpstream
	}
	return resp, nil
}

func (s *Service) call(ctx context.Context, endpoint string, values url.Values) (subsonicResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancel()
	resp, err := s.request(ctx, endpoint, values)
	if err != nil {
		return subsonicResponse{}, err
	}
	defer resp.Body.Close()
	var envelope struct {
		Response subsonicResponse `json:"subsonic-response"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(resp.Body, 8*1024*1024)).Decode(&envelope) != nil {
		return subsonicResponse{}, ErrUpstream
	}
	if envelope.Response.Status != "ok" {
		if endpoint == "getSong" && envelope.Response.Error.Code == 70 {
			return subsonicResponse{}, ErrSongNotFound
		}
		return subsonicResponse{}, ErrUpstream
	}
	return envelope.Response, nil
}

func newHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 15 * time.Second
	return &http.Client{
		Transport: transport,
		// 不将令牌发送到重定向目标，也不将上游登录页作为音乐返回。
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
}
