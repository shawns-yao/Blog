package ws

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	appcomment "github.com/shawns-yao/shawn-blog/server/internal/app/comment"
	appEvent "github.com/shawns-yao/shawn-blog/server/internal/app/event"
	"github.com/shawns-yao/shawn-blog/server/internal/app/moment"
	domainalbum "github.com/shawns-yao/shawn-blog/server/internal/domain/album"
	domaincomment "github.com/shawns-yao/shawn-blog/server/internal/domain/comment"
	"github.com/shawns-yao/shawn-blog/server/internal/domain/content"
)

const siteActivityType = "site.activity"

type siteActivityPayload struct {
	Type          string `json:"type"`
	Event         string `json:"event"`
	ContentType   string `json:"contentType"`
	Title         string `json:"title"`
	Excerpt       string `json:"excerpt,omitempty"`
	URL           string `json:"url"`
	At            string `json:"at"`
	CommentAreaID int64  `json:"commentAreaId,omitempty"`
}

func RegisterSiteActivitySubscriber(
	bus appEvent.Bus,
	manager *Manager,
	contentRepo content.Repository,
	commentRepo domaincomment.CommentRepository,
	albumRepo domainalbum.Repository,
) {
	if bus == nil || manager == nil {
		return
	}

	bus.Subscribe(moment.MomentHotMarked{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		marked, ok := event.(moment.MomentHotMarked)
		if !ok || strings.TrimSpace(marked.ShortURL) == "" || contentRepo == nil || marked.ID <= 0 {
			return nil
		}
		item, err := contentRepo.GetMomentByID(ctx, marked.ID)
		if err != nil || item == nil || !item.IsPublished {
			return nil
		}
		broadcastSiteActivity(manager, siteActivityPayload{
			Type:        siteActivityType,
			Event:       marked.Name(),
			ContentType: "moment",
			Title:       normalizeActivityTitle(marked.Title, "手记"),
			URL:         buildMomentPath(marked.ShortURL, item.CreatedAt),
			At:          normalizeActivityAt(marked.At),
		})
		return nil
	}))

	bus.Subscribe(moment.MomentUpdated{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		updated, ok := event.(moment.MomentUpdated)
		if !ok || !updated.Published || contentRepo == nil || updated.ID <= 0 {
			return nil
		}
		item, err := contentRepo.GetMomentByID(ctx, updated.ID)
		if err != nil || item == nil || !item.IsPublished || strings.TrimSpace(item.ShortURL) == "" {
			return nil
		}
		broadcastSiteActivity(manager, siteActivityPayload{
			Type:        siteActivityType,
			Event:       updated.Name(),
			ContentType: "moment",
			Title:       normalizeActivityTitle(item.Title, "手记"),
			URL:         buildMomentPath(item.ShortURL, item.CreatedAt),
			At:          normalizeActivityAt(updated.At),
		})
		return nil
	}))

	bus.Subscribe(appcomment.CommentCreated{}.Name(), handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		created, ok := event.(appcomment.CommentCreated)
		if !ok || !strings.EqualFold(strings.TrimSpace(created.Status), domaincomment.CommentStatusApproved) {
			return nil
		}
		title, link, ok := resolveCommentTarget(ctx, contentRepo, commentRepo, albumRepo, created.AreaID)
		if !ok {
			return nil
		}
		broadcastSiteActivity(manager, siteActivityPayload{
			Type:          siteActivityType,
			Event:         created.Name(),
			ContentType:   "comment",
			Title:         title,
			Excerpt:       summarizeActivityText(created.Content, 56),
			URL:           link,
			At:            normalizeActivityAt(created.At),
			CommentAreaID: created.AreaID,
		})
		return nil
	}))

	bus.Subscribe("comment.updated", handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		generic, ok := event.(appEvent.Generic)
		if !ok || generic.Payload == nil {
			return nil
		}
		status, _ := generic.Payload["Status"].(string)
		if !strings.EqualFold(strings.TrimSpace(status), domaincomment.CommentStatusApproved) {
			return nil
		}

		areaID := anyToInt64(generic.Payload["AreaID"])
		if areaID <= 0 {
			return nil
		}

		title, link, ok := resolveCommentTarget(ctx, contentRepo, commentRepo, albumRepo, areaID)
		if !ok {
			return nil
		}

		excerpt := ""
		commentID := anyToInt64(generic.Payload["ID"])
		if commentRepo != nil && commentID > 0 {
			item, err := commentRepo.FindByID(ctx, commentID)
			if err == nil && item != nil {
				excerpt = summarizeActivityText(item.Content, 56)
			}
		}

		broadcastSiteActivity(manager, siteActivityPayload{
			Type:          siteActivityType,
			Event:         "comment.approved",
			ContentType:   "comment",
			Title:         title,
			Excerpt:       excerpt,
			URL:           link,
			At:            normalizeActivityAt(generic.At),
			CommentAreaID: areaID,
		})
		return nil
	}))
}

func resolveCommentTarget(
	ctx context.Context,
	contentRepo content.Repository,
	commentRepo domaincomment.CommentRepository,
	albumRepo domainalbum.Repository,
	areaID int64,
) (string, string, bool) {
	if commentRepo == nil || areaID <= 0 {
		return "", "", false
	}
	area, err := commentRepo.GetAreaByID(ctx, areaID)
	if err != nil || area == nil {
		return "", "", false
	}
	if area.Type == "guestbook" {
		return "留言板", "/message", true
	}
	if area.ContentID == nil || *area.ContentID <= 0 {
		return "", "", false
	}

	targetID := *area.ContentID
	switch strings.ToLower(strings.TrimSpace(area.Type)) {
	case "moment":
		if contentRepo == nil {
			return "", "", false
		}
		item, err := contentRepo.GetMomentByID(ctx, targetID)
		if err != nil || item == nil || !item.IsPublished || strings.TrimSpace(item.ShortURL) == "" {
			return "", "", false
		}
		return normalizeActivityTitle(item.Title, "手记"), buildMomentPath(item.ShortURL, item.CreatedAt), true
	case "album":
		if albumRepo == nil {
			return "", "", false
		}
		item, err := albumRepo.GetAlbumByID(ctx, targetID)
		if err != nil || item == nil || !item.IsPublished || strings.TrimSpace(item.ShortURL) == "" {
			return "", "", false
		}
		return normalizeActivityTitle(item.Title, "相册"), "/albums/" + url.PathEscape(strings.TrimSpace(item.ShortURL)), true
	default:
		return "", "", false
	}
}

func broadcastSiteActivity(manager *Manager, payload siteActivityPayload) {
	if manager == nil || strings.TrimSpace(payload.URL) == "" {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	manager.Broadcast(RealtimeRoomKey(), data)
}

func normalizeActivityTitle(raw string, fallback string) string {
	title := strings.TrimSpace(raw)
	if title == "" {
		return fallback
	}
	return summarizeActivityText(title, 60)
}

func normalizeActivityAt(raw time.Time) string {
	at := raw
	if at.IsZero() {
		at = time.Now().UTC()
	}
	return at.UTC().Format(time.RFC3339)
}

func summarizeActivityText(raw string, maxRunes int) string {
	text := strings.TrimSpace(raw)
	if text == "" || maxRunes <= 0 {
		return ""
	}
	text = strings.Join(strings.Fields(text), " ")
	if utf8.RuneCountInString(text) <= maxRunes {
		return text
	}

	runes := []rune(text)
	if len(runes) <= maxRunes {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[:maxRunes])) + "…"
}

func anyToInt64(raw any) int64 {
	switch v := raw.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	default:
		return 0
	}
}
