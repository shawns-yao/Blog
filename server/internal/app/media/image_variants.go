package media

import (
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/webp"
)

var responsiveWidths = []int{320, 640, 1280}

const responsiveQuality = 82

func imageVariantStoredPath(storedPath string, width int) string {
	return fmt.Sprintf("/%s/webp-v1%s.%d.webp", thumbnailDir, storedPath, width)
}

// ResolveImageDelivery uses the original upload URL plus a bounded width preset.
// Old uploads are prepared on first use; subsequent reads only serve cached files.
func (s *Service) ResolveImageDelivery(ctx context.Context, storedPath string, width int) (Delivery, error) {
	normalized, err := normalizeStoredPath(storedPath)
	if err != nil {
		return Delivery{}, err
	}
	if !slices.Contains(responsiveWidths, width) || !isResponsiveImagePath(normalized) {
		return s.ResolveDelivery(ctx, normalized)
	}
	variantPath := imageVariantStoredPath(normalized, width)
	if !fileExists(s.diskPathFromStored(variantPath)) {
		unlock := s.beginMutation()
		_, err = s.ensureImageVariants(s.diskPathFromStored(normalized), normalized)
		unlock()
		if err != nil {
			return s.ResolveDelivery(ctx, normalized)
		}
	}
	if fileExists(s.diskPathFromStored(variantPath)) {
		return s.ResolveDelivery(ctx, variantPath)
	}
	return s.ResolveDelivery(ctx, normalized)
}

func isResponsiveImagePath(storedPath string) bool {
	if !strings.HasPrefix(storedPath, "/pictures/") {
		return false
	}
	switch strings.ToLower(filepath.Ext(storedPath)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	}
	return false
}

func (s *Service) ensureImageVariants(diskPath string, storedPath string) (*ImageMeta, error) {
	if !isResponsiveImagePath(storedPath) {
		return nil, nil
	}
	// Share one encoder slot between HTTP reads and the existing upload worker.
	s.imageMu.Lock()
	defer s.imageMu.Unlock()
	complete := true
	for _, width := range responsiveWidths {
		complete = complete && fileExists(s.diskPathFromStored(imageVariantStoredPath(storedPath, width)))
	}
	if complete {
		return nil, nil
	}
	f, err := os.Open(diskPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// Animated WebP must retain its frames instead of becoming a still thumbnail.
	var header [21]byte
	n, _ := io.ReadFull(f, header[:])
	if n == len(header) && string(header[8:12]) == "WEBP" && string(header[12:16]) == "VP8X" && header[20]&2 != 0 {
		return nil, nil
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	config, format, err := image.DecodeConfig(f)
	if err != nil {
		return nil, err
	}
	if format == "gif" {
		return nil, nil
	}
	// Decode one image at a time and reject oversized sources before allocation.
	if int64(config.Width)*int64(config.Height) > 40_000_000 {
		return nil, errors.New("image exceeds thumbnail pixel limit")
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	var src image.Image
	if format == "webp" {
		src, err = webp.Decode(f, webp.Options{AutoRotate: true})
	} else {
		src, err = imaging.Decode(f, imaging.AutoOrientation(true))
	}
	if err != nil {
		return nil, err
	}
	bounds := src.Bounds()
	meta := &ImageMeta{Width: bounds.Dx(), Height: bounds.Dy(), DominantColor: calcDominantColor(src)}
	for _, width := range responsiveWidths {
		target := s.diskPathFromStored(imageVariantStoredPath(storedPath, width))
		if fileExists(target) {
			continue
		}
		thumb := imaging.Resize(src, min(width, bounds.Dx()), 0, imaging.Lanczos)
		if err := writeWebP(target, thumb); err != nil {
			return meta, err
		}
	}
	return meta, nil
}

func writeWebP(target string, src image.Image) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(target), ".webp-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	encodeErr := webp.Encode(f, src, webp.Options{Quality: responsiveQuality, Method: 4})
	closeErr := f.Close()
	if encodeErr != nil {
		return encodeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), target)
}
