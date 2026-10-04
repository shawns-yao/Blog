package music

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

func (s *Service) Upload(ctx context.Context, filename string, input io.Reader) (UploadResult, error) {
	if !s.configured {
		return UploadResult{}, ErrUnavailable
	}
	filename = filepath.Base(strings.ReplaceAll(filename, "\\", "/"))
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".mp3" && ext != ".flac" && ext != ".m4a" {
		return UploadResult{}, &InputError{"仅支持 MP3、FLAC 和 M4A 音乐文件"}
	}
	name := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || strings.ContainsRune("<>:\"/\\|?*", r) {
			return '_'
		}
		return r
	}, strings.TrimSuffix(filename, filepath.Ext(filename)))
	if len([]rune(name)) > 80 {
		name = string([]rune(name)[:80])
	}
	name = strings.Trim(name, ". ")
	if name == "" {
		name = "track"
	}
	filename = name + ext
	if err := os.MkdirAll(s.cfg.StagingDir, 0750); err != nil {
		return UploadResult{}, errors.New("无法创建音乐临时目录")
	}
	temp, err := os.CreateTemp(s.cfg.StagingDir, "import-*")
	if err != nil {
		return UploadResult{}, errors.New("无法保存音乐文件")
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(temp, hash), io.LimitReader(input, s.cfg.MaxUploadBytes+1))
	if err != nil {
		return UploadResult{}, errors.New("音乐文件写入失败")
	}
	if size == 0 || size > s.cfg.MaxUploadBytes {
		return UploadResult{}, &InputError{"音乐文件为空或超过单文件大小限制"}
	}
	if err := temp.Close(); err != nil {
		return UploadResult{}, errors.New("音乐文件写入失败")
	}
	id := hex.EncodeToString(hash.Sum(nil))
	result := UploadResult{ID: id, Filename: filename, Title: name, Size: size, State: "queued"}
	s.uploadMu.Lock()
	defer s.uploadMu.Unlock()
	retry := false
	if job, err := s.UploadStatus(id); err == nil {
		if job.State != "failed" {
			job.Duplicate = true
			if job.State != "ready" {
				s.startUploadWorker()
			}
			return job, nil
		}
		result = job
		result.State, result.Error, result.Duplicate = "queued", "", false
		filename = result.Filename
		retry = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return UploadResult{}, err
	}
	if files, err := os.ReadDir(filepath.Join(s.cfg.LibraryDir, id)); !retry && err == nil {
		for _, file := range files {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if file.IsDir() || (ext != ".mp3" && ext != ".flac" && ext != ".m4a") {
				continue
			}
			result.PlaybackFilename = file.Name()
			result.State, result.Duplicate = "ready", true
			if err := s.saveUpload(result); err != nil {
				return UploadResult{}, err
			}
			return result, nil
		}
	}
	if err := ctx.Err(); err != nil {
		return UploadResult{}, err
	}
	dir := filepath.Join(s.musicRoot(), "originals", id)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return UploadResult{}, errors.New("无法创建原始音乐目录")
	}
	// 原文件先保存在扫描目录之外；请求返回后由后台生成播放副本。
	if err := os.Chmod(temp.Name(), 0640); err != nil {
		return UploadResult{}, errors.New("无法设置音乐文件权限")
	}
	if err := os.Rename(temp.Name(), filepath.Join(dir, filename)); err != nil {
		return UploadResult{}, errors.New("无法保存原始音乐文件，请确认音乐目录位于同一文件系统")
	}
	if err := s.saveUpload(result); err != nil {
		return UploadResult{}, err
	}
	s.startUploadWorker()
	return result, nil
}

func (s *Service) probe(ctx context.Context, filename, extension string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.cfg.ProbeBinary, "-v", "error", "-protocol_whitelist", "file,pipe", "-show_entries", "stream=codec_type,codec_name:format=duration:format_tags=title,artist", "-of", "json", filename)
	output, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, errors.New("音乐上传需要安装 ffprobe")
		}
		return nil, &InputError{"文件无法识别为有效音乐"}
	}
	var data struct {
		Streams []struct {
			Type  string `json:"codec_type"`
			Codec string `json:"codec_name"`
		} `json:"streams"`
		Format struct {
			Tags map[string]string `json:"tags"`
		} `json:"format"`
	}
	if json.Unmarshal(output, &data) != nil {
		return nil, &InputError{"音乐文件信息无法读取"}
	}
	valid := false
	for _, stream := range data.Streams {
		if stream.Type == "audio" && ((extension == ".mp3" && stream.Codec == "mp3") || (extension == ".flac" && stream.Codec == "flac") || (extension == ".m4a" && (stream.Codec == "aac" || stream.Codec == "alac"))) {
			valid = true
		}
	}
	if !valid {
		return nil, &InputError{"文件扩展名与音乐编码不匹配"}
	}
	tags := make(map[string]string)
	for key, value := range data.Format.Tags {
		tags[strings.ToLower(key)] = value
	}
	return tags, nil
}
