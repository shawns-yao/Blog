package media

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

type fakeRemoteStorage struct {
	exists      bool
	existsErr   error
	readURL     string
	putStarted  chan struct{}
	putRelease  chan struct{}
	putComplete chan struct{}
	active      atomic.Int32
	maxActive   atomic.Int32
}

func (f *fakeRemoteStorage) Exists(context.Context, string) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeRemoteStorage) PutFile(context.Context, string, string) error {
	active := f.active.Add(1)
	for {
		current := f.maxActive.Load()
		if active <= current || f.maxActive.CompareAndSwap(current, active) {
			break
		}
	}
	if f.putStarted != nil {
		select {
		case f.putStarted <- struct{}{}:
		default:
		}
	}
	if f.putRelease != nil {
		<-f.putRelease
	}
	f.active.Add(-1)
	if f.putComplete != nil {
		f.putComplete <- struct{}{}
	}
	return nil
}

func (f *fakeRemoteStorage) ReadURL(context.Context, string) (string, error) {
	return f.readURL, nil
}

func (f *fakeRemoteStorage) Delete(context.Context, string) error { return nil }

func TestUploadReturnsBeforeBackgroundMirrorCompletes(t *testing.T) {
	uploadDir := t.TempDir()
	repo := newMemoryRepo()
	remote := &fakeRemoteStorage{
		putStarted: make(chan struct{}, 1),
		putRelease: make(chan struct{}),
	}
	svc := NewService(repo, uploadDir, nil)
	svc.SetRemoteStorage(remote)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc.StartBackground(ctx)

	sourcePath := filepath.Join(t.TempDir(), "photo.png")
	writePNG(t, sourcePath)
	header := multipartHeaderForFile(t, sourcePath)

	done := make(chan *UploadResult, 1)
	go func() {
		result, err := svc.Upload(context.Background(), header, "picture")
		if err != nil {
			t.Errorf("Upload() error = %v", err)
			done <- nil
			return
		}
		done <- result
	}()

	select {
	case result := <-done:
		if result == nil || !result.Created {
			t.Fatalf("Upload() result = %#v, want created record", result)
		}
	case <-time.After(time.Second):
		t.Fatal("Upload() waited for the background media job")
	}

	select {
	case <-remote.putStarted:
	case <-time.After(time.Second):
		t.Fatal("background mirror did not start")
	}
	close(remote.putRelease)
}

func TestMediaWorkerProcessesJobsWithSingleConcurrency(t *testing.T) {
	uploadDir := t.TempDir()
	writeText(t, filepath.Join(uploadDir, "files", "a.txt"), "a")
	writeText(t, filepath.Join(uploadDir, "files", "b.txt"), "b")

	remote := &fakeRemoteStorage{
		putRelease:  make(chan struct{}),
		putComplete: make(chan struct{}, 2),
	}
	svc := NewService(newMemoryRepo(), uploadDir, nil)
	svc.SetRemoteStorage(remote)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc.StartBackground(ctx)
	svc.enqueueMedia("/files/a.txt", "file")
	svc.enqueueMedia("/files/b.txt", "file")

	deadline := time.After(time.Second)
	for remote.active.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("worker did not start")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	close(remote.putRelease)
	for range 2 {
		select {
		case <-remote.putComplete:
		case <-time.After(time.Second):
			t.Fatal("worker did not complete queued jobs")
		}
	}
	if got := remote.maxActive.Load(); got != 1 {
		t.Fatalf("maximum worker concurrency = %d, want 1", got)
	}
}

func TestResolveDeliveryFallsBackToLocalWhenRemoteFails(t *testing.T) {
	uploadDir := t.TempDir()
	localPath := filepath.Join(uploadDir, "pictures", "photo.jpg")
	writeText(t, localPath, "image")

	svc := NewService(newMemoryRepo(), uploadDir, nil)
	svc.SetRemoteStorage(&fakeRemoteStorage{existsErr: errors.New("R2 unavailable")})
	delivery, err := svc.ResolveDelivery(context.Background(), "/pictures/photo.jpg")
	if err != nil {
		t.Fatalf("ResolveDelivery() error = %v", err)
	}
	if delivery.LocalPath != localPath {
		t.Fatalf("LocalPath = %q, want %q", delivery.LocalPath, localPath)
	}
	if delivery.RemoteURL != "" {
		t.Fatalf("RemoteURL = %q, want empty fallback URL", delivery.RemoteURL)
	}
}

func multipartHeaderForFile(t *testing.T, path string) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	source, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := io.Copy(part, source); err != nil {
		_ = source.Close()
		t.Fatalf("Copy() error = %v", err)
	}
	_ = source.Close()
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart writer close error = %v", err)
	}

	reader := multipart.NewReader(bytes.NewReader(body.Bytes()), writer.Boundary())
	form, err := reader.ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("ReadForm() error = %v", err)
	}
	t.Cleanup(func() { _ = form.RemoveAll() })
	return form.File["file"][0]
}

var _ remoteStorage = (*fakeRemoteStorage)(nil)
