package media

import (
	"context"
	"log"
	"sync"
)

type mediaJob struct {
	storedPath string
	fileType   string
}

type mediaJobQueue struct {
	mu      sync.Mutex
	jobs    []mediaJob
	pending map[string]struct{}
	wake    chan struct{}
}

func newMediaJobQueue() *mediaJobQueue {
	return &mediaJobQueue{
		pending: make(map[string]struct{}),
		wake:    make(chan struct{}, 1),
	}
}

func (q *mediaJobQueue) enqueue(job mediaJob) {
	q.mu.Lock()
	if _, exists := q.pending[job.storedPath]; exists {
		q.mu.Unlock()
		return
	}
	q.pending[job.storedPath] = struct{}{}
	q.jobs = append(q.jobs, job)
	q.mu.Unlock()
	select {
	case q.wake <- struct{}{}:
	default:
	}
}

func (q *mediaJobQueue) pop() (mediaJob, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.jobs) == 0 {
		return mediaJob{}, false
	}
	job := q.jobs[0]
	q.jobs[0] = mediaJob{}
	q.jobs = q.jobs[1:]
	return job, true
}

func (q *mediaJobQueue) done(job mediaJob) {
	q.mu.Lock()
	delete(q.pending, job.storedPath)
	q.mu.Unlock()
}

func (s *Service) StartBackground(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	s.workerOnce.Do(func() {
		go s.runWorker(ctx)
	})
}

// EnqueueExisting schedules every indexed upload after a restart. Existing
// thumbnails and remote objects are detected and skipped, so interrupted work
// converges without a separate persistent job table.
func (s *Service) EnqueueExisting(ctx context.Context) error {
	items, err := s.repo.ListAll(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		s.enqueueMedia(item.Path, item.Type)
	}
	return nil
}

func (s *Service) enqueueMedia(storedPath string, fileType string) {
	if storedPath == "" {
		return
	}
	s.StartBackground(context.Background())
	s.queue.enqueue(mediaJob{storedPath: storedPath, fileType: fileType})
}

func (s *Service) runWorker(ctx context.Context) {
	for {
		for {
			job, ok := s.queue.pop()
			if !ok {
				break
			}
			s.processMediaJob(ctx, job)
			s.queue.done(job)
		}
		select {
		case <-ctx.Done():
			return
		case <-s.queue.wake:
		}
	}
}

func (s *Service) processMediaJob(ctx context.Context, job mediaJob) {
	unlock := s.beginMutation()
	defer unlock()

	diskPath := s.diskPathFromStored(job.storedPath)
	if !fileExists(diskPath) {
		return
	}

	var thumbStoredPath string
	if job.fileType == "picture" {
		thumbURL, _ := s.processImage(diskPath, job.storedPath, "pictures")
		thumbStoredPath = storedPathFromPublicURL(thumbURL)
	}
	if s.remote == nil {
		return
	}

	if err := s.ensureRemoteFile(ctx, job.storedPath, diskPath); err != nil {
		log.Printf("[media-worker] mirror original failed path=%s: %v", job.storedPath, err)
		return
	}
	if thumbStoredPath == "" {
		return
	}
	thumbDiskPath := s.diskPathFromStored(thumbStoredPath)
	if err := s.ensureRemoteFile(ctx, thumbStoredPath, thumbDiskPath); err != nil {
		log.Printf("[media-worker] mirror thumbnail failed path=%s: %v", thumbStoredPath, err)
	}
}

func (s *Service) ensureRemoteFile(ctx context.Context, storedPath string, diskPath string) error {
	exists, err := s.remote.Exists(ctx, storedPath)
	if err == nil && exists {
		return nil
	}
	return s.remote.PutFile(ctx, storedPath, diskPath)
}

func storedPathFromPublicURL(publicURL string) string {
	const prefix = "/uploads"
	if len(publicURL) <= len(prefix) || publicURL[:len(prefix)] != prefix {
		return ""
	}
	return publicURL[len(prefix):]
}
