package music

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (s *Service) musicRoot() string { return filepath.Dir(s.cfg.LibraryDir) }

func validUploadID(id string) bool {
	_, err := hex.DecodeString(id)
	return len(id) == 64 && err == nil
}

func (s *Service) UploadStatus(id string) (UploadResult, error) {
	if !validUploadID(id) {
		return UploadResult{}, &InputError{"上传任务编号无效"}
	}
	data, err := os.ReadFile(filepath.Join(s.musicRoot(), "jobs", id+".json"))
	if err != nil {
		return UploadResult{}, err
	}
	var job UploadResult
	if err := json.Unmarshal(data, &job); err != nil {
		return UploadResult{}, errors.New("上传任务记录无法读取")
	}
	if job.ID != id || filepath.Base(job.Filename) != job.Filename || (job.PlaybackFilename != "" && filepath.Base(job.PlaybackFilename) != job.PlaybackFilename) {
		return UploadResult{}, errors.New("上传任务记录无效")
	}
	return job, nil
}

func (s *Service) saveUpload(job UploadResult) error {
	dir := filepath.Join(s.musicRoot(), "jobs")
	if err := os.MkdirAll(dir, 0750); err != nil {
		return errors.New("无法创建上传任务目录")
	}
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return writeMusicFile(filepath.Join(dir, job.ID+".json"), data, 0600)
}

func writeMusicFile(path string, data []byte, mode os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".music-*")
	if err != nil {
		return errors.New("无法保存音乐处理文件")
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if _, err = file.Write(data); err != nil {
		return errors.New("无法保存音乐处理文件")
	}
	if err = file.Sync(); err != nil {
		return errors.New("无法保存音乐处理文件")
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err = os.Chmod(file.Name(), mode); err != nil {
		return errors.New("无法设置音乐处理文件权限")
	}
	if err = os.Rename(file.Name(), path); err != nil {
		return errors.New("无法发布音乐处理文件")
	}
	return nil
}

func (s *Service) nextUpload() (UploadResult, bool, error) {
	entries, err := os.ReadDir(filepath.Join(s.musicRoot(), "jobs"))
	if errors.Is(err, os.ErrNotExist) {
		return UploadResult{}, false, nil
	}
	if err != nil {
		return UploadResult{}, false, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		job, err := s.UploadStatus(strings.TrimSuffix(entry.Name(), ".json"))
		if err != nil {
			slog.Error("音乐上传任务记录无法读取", "task", strings.TrimSuffix(entry.Name(), ".json"))
			continue
		}
		if job.State == "queued" || job.State == "processing" || job.State == "scanning" {
			return job, true, nil
		}
	}
	return UploadResult{}, false, nil
}

func (s *Service) startUploadWorker() {
	s.workerMu.Lock()
	defer s.workerMu.Unlock()
	if s.processing {
		return
	}
	s.processing = true
	go func() {
		for {
			// 与新上传的启动操作串行，避免队列刚变空时遗漏新任务。
			s.workerMu.Lock()
			job, found, err := s.nextUpload()
			if err != nil || !found {
				if err != nil {
					slog.Error("音乐上传任务队列无法读取")
				}
				s.processing = false
				s.workerMu.Unlock()
				return
			}
			s.workerMu.Unlock()
			job.State, job.Error = "processing", ""
			if s.saveUpload(job) != nil {
				slog.Error("音乐上传任务状态无法保存", "task", job.ID)
				s.stopUploadWorker()
				return
			}
			if err := s.prepareUpload(&job); err != nil {
				job.State, job.Error = "failed", err.Error()
			} else {
				job.State = "ready"
			}
			if s.saveUpload(job) != nil {
				slog.Error("音乐上传处理结果无法保存", "task", job.ID)
				s.stopUploadWorker()
				return
			}
		}
	}()
}

func (s *Service) stopUploadWorker() {
	s.workerMu.Lock()
	s.processing = false
	s.workerMu.Unlock()
}

func (s *Service) prepareUpload(job *UploadResult) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	original := filepath.Join(s.musicRoot(), "originals", job.ID, job.Filename)
	metadata, err := s.probe(ctx, original, strings.ToLower(filepath.Ext(job.Filename)))
	if err != nil {
		return err
	}
	if metadata["title"] != "" {
		job.Title = metadata["title"]
	}
	job.Artist = metadata["artist"]
	dir := filepath.Join(s.cfg.LibraryDir, job.ID)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return errors.New("无法创建播放缓存目录")
	}
	output := filepath.Join(dir, uploadPlaybackFilename(*job))
	if info, err := os.Stat(output); err != nil || info.Size() == 0 {
		file, err := os.CreateTemp(s.cfg.StagingDir, "playback-*.mp3")
		if err != nil {
			return errors.New("无法创建播放缓存文件")
		}
		file.Close()
		defer os.Remove(file.Name())
		cmd := exec.CommandContext(ctx, s.cfg.TranscodeBinary,
			"-nostdin", "-v", "error", "-y", "-protocol_whitelist", "file,pipe", "-i", original,
			"-map", "0:a:0", "-map", "0:v?", "-c:a", "libmp3lame", "-b:a", fmt.Sprintf("%dk", s.cfg.MaxBitRate),
			"-c:v", "copy", "-map_metadata", "0", "-id3v2_version", "3", "-write_xing", "1", file.Name())
		if err := cmd.Run(); err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				return errors.New("服务器缺少 ffmpeg，无法生成播放缓存")
			}
			return errors.New("播放缓存生成失败，请重新上传")
		}
		if err := os.Chmod(file.Name(), 0640); err != nil {
			return errors.New("无法设置播放缓存权限")
		}
		if err := os.Rename(file.Name(), output); err != nil {
			return errors.New("无法保存播放缓存")
		}
	}
	job.PlaybackFilename = filepath.Base(output)
	if err := s.publishUploadLyrics(*job); err != nil {
		return err
	}
	job.State = "scanning"
	if err := s.saveUpload(*job); err != nil {
		return err
	}
	if _, err := s.Scan(ctx, true); err != nil {
		return errors.New("播放缓存已保存，音乐扫描失败，请重新上传重试")
	}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return errors.New("处理音乐超时，请重新上传重试")
		case <-ticker.C:
			scan, err := s.Scan(ctx, false)
			if err != nil {
				return errors.New("音乐扫描状态无法读取，请重新上传重试")
			}
			if !scan.Scanning {
				return nil
			}
		}
	}
}

func uploadPlaybackFilename(job UploadResult) string {
	if job.PlaybackFilename != "" {
		return job.PlaybackFilename
	}
	return strings.TrimSuffix(job.Filename, filepath.Ext(job.Filename)) + ".mp3"
}
