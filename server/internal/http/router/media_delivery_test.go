package router

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"io"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gen2brain/webp"
	"github.com/gofiber/fiber/v2"
	mediaapp "github.com/shawns-yao/shawn-blog/server/internal/app/media"
)

// 定向测试通过实际 /uploads HTTP 入口验证编码、缓存与原图回退，不连接数据库。
func TestMediaResponsiveDelivery(t *testing.T) {
	uploadDir := os.Getenv("BLOG_TEST_IMAGE_DELIVERY_DIR")
	if uploadDir == "" {
		// SendFile caches open files. Run the HTTP checks in a child process so
		// Windows releases those handles before the parent removes its fixtures.
		cmd := exec.Command(os.Args[0], "-test.run=^TestMediaResponsiveDelivery$", "-test.v")
		cmd.Env = append(os.Environ(), "BLOG_TEST_IMAGE_DELIVERY_DIR="+t.TempDir())
		output, err := cmd.CombinedOutput()
		t.Logf("%s", output)
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	if err := os.MkdirAll(filepath.Join(uploadDir, "pictures"), 0o755); err != nil {
		t.Fatal(err)
	}
	fixtures := map[string]string{
		"scene.webp": filepath.Join("..", "..", "..", "..", "web", "static", "home-scenes", "11.webp"),
		"avatar.png": filepath.Join("..", "..", "..", "..", "web", "static", "images", "rag-shuling-avatar_20260930.png"),
		"photo.jpg":  filepath.Join("..", "..", "..", "..", "web", "static", "home-scenes", "portrait-day_20261001.jpg"),
	}
	for name, path := range fixtures {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(uploadDir, "pictures", name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	app := fiber.New()
	registerMediaDelivery(app, Dependencies{Media: mediaapp.NewService(nil, uploadDir, nil)})
	t.Cleanup(func() { _ = app.Shutdown() })
	request := func(path string) ([]byte, string) {
		t.Helper()
		resp, err := app.Test(httptest.NewRequest("GET", path, nil), -1)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil || resp.StatusCode != 200 {
			t.Fatalf("GET %s: status=%d error=%v", path, resp.StatusCode, err)
		}
		if !strings.Contains(resp.Header.Get("Cache-Control"), "immutable") {
			t.Fatalf("GET %s missing immutable cache", path)
		}
		return body, resp.Header.Get("Content-Type")
	}
	original, _ := request("/uploads/pictures/scene.webp")
	originalHash := sha256.Sum256(original)
	for _, width := range []int{320, 640, 1280} {
		start := time.Now()
		body, contentType := request(fmt.Sprintf("/uploads/pictures/scene.webp?width=%d", width))
		config, format, err := image.DecodeConfig(bytes.NewReader(body))
		if err != nil || format != "webp" || config.Width != width || !strings.HasPrefix(contentType, "image/webp") {
			t.Fatalf("width=%d: config=%+v format=%s contentType=%s error=%v", width, config, format, contentType, err)
		}
		t.Logf("scene width=%d original=%d variant=%d request=%s", width, len(original), len(body), time.Since(start))
	}
	cachePath := filepath.Join(uploadDir, "thumbnails", "webp-v1", "pictures", "scene.webp.640.webp")
	before, err := os.Stat(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	var readers sync.WaitGroup
	for range 4 {
		readers.Go(func() { request("/uploads/pictures/scene.webp?width=640") })
	}
	readers.Wait()
	after, err := os.Stat(cachePath)
	if err != nil || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("cached variant was regenerated")
	}
	unchanged, _ := request("/uploads/pictures/scene.webp")
	if sha256.Sum256(unchanged) != originalHash {
		t.Fatal("original upload changed")
	}
	fallback, _ := request("/uploads/pictures/scene.webp?width=999999")
	if !bytes.Equal(fallback, original) {
		t.Fatal("unsupported width must retain the original")
	}
	avatar, _ := request("/uploads/pictures/avatar.png?width=1280")
	decoded, format, err := image.Decode(bytes.NewReader(avatar))
	if err != nil || format != "webp" || decoded.Bounds().Dx() != 256 {
		t.Fatalf("small image was enlarged or encoding failed: format=%s error=%v", format, err)
	}
	_, _, _, alpha := decoded.At(0, 0).RGBA()
	if alpha != 0 {
		t.Fatal("transparent pixels lost alpha")
	}
	jpegBody, jpegType := request("/uploads/pictures/photo.jpg?width=640")
	jpegConfig, jpegFormat, err := image.DecodeConfig(bytes.NewReader(jpegBody))
	if err != nil || jpegFormat != "webp" || jpegConfig.Width != 640 || !strings.HasPrefix(jpegType, "image/webp") {
		t.Fatalf("JPEG conversion failed: config=%+v format=%s error=%v", jpegConfig, jpegFormat, err)
	}
	palette := color.Palette{color.Black, color.White}
	frames := []*image.Paletted{image.NewPaletted(image.Rect(0, 0, 2, 2), palette), image.NewPaletted(image.Rect(0, 0, 2, 2), palette)}
	frames[1].SetColorIndex(0, 0, 1)
	var animation bytes.Buffer
	if err := gif.EncodeAll(&animation, &gif.GIF{Image: frames, Delay: []int{10, 10}}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(uploadDir, "pictures", "animated.gif"), animation.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	gifBody, _ := request("/uploads/pictures/animated.gif?width=320")
	if !bytes.Equal(gifBody, animation.Bytes()) {
		t.Fatal("animation was converted to a still image")
	}
	var animatedWebP bytes.Buffer
	if err := webp.EncodeAll(&animatedWebP, &webp.WEBP{Image: []image.Image{frames[0], frames[1]}, Delay: []int{100, 100}}); err != nil {
		t.Fatal(err)
	}
	webpFrames, err := webp.DecodeAll(bytes.NewReader(animatedWebP.Bytes()))
	if err != nil || len(webpFrames.Image) != 2 {
		t.Fatal("WebP fixture must contain two different frames")
	}
	if err := os.WriteFile(filepath.Join(uploadDir, "pictures", "animated.webp"), animatedWebP.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	webpBody, _ := request("/uploads/pictures/animated.webp?width=320")
	if !bytes.Equal(webpBody, animatedWebP.Bytes()) {
		t.Fatal("animated WebP was converted to a still image")
	}
}
