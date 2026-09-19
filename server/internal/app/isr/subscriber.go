package isr

import (
	"context"
	"fmt"
	"log"
	"strings"

	appalbum "github.com/grtsinry43/grtblog-v2/server/internal/app/album"
	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/moment"
	"github.com/redis/go-redis/v9"
)

type handlerFunc func(ctx context.Context, event appEvent.Event) error

func (h handlerFunc) Handle(ctx context.Context, event appEvent.Event) error {
	return h(ctx, event)
}

func RegisterMomentSubscribers(bus appEvent.Bus, service *Service) {
	if bus == nil || service == nil {
		return
	}

	register := func(eventName string) {
		bus.Subscribe(eventName, handlerFunc(func(ctx context.Context, event appEvent.Event) error {
			momentID, shortURL := extractMomentEventPayload(event)
			if momentID <= 0 {
				return nil
			}

			deps := []string{
				"home:recent-moments",
				"home:activity-pulse",
				"home:inspiration-stats",
				"column:list",
				"timeline:by-year",
				"moment:list:page:*",
				fmt.Sprintf("moment:detail:%d", momentID),
			}
			urls := []string{
				"/",
				"/timeline",
				"/moments",
			}
			// Brand-new moments are not tracked under any dep key yet, so the
			// date-segmented detail URL must be enqueued directly. Deleted
			// moments won't resolve; their tracked URL is handled by the dep.
			if detailURL, ok := service.MomentDetailURL(ctx, shortURL); ok {
				urls = append(urls, detailURL)
			}
			return service.Invalidate(ctx, deps, urls)
		}))
	}

	register(moment.MomentCreated{}.Name())
	register(moment.MomentUpdated{}.Name())
	register(moment.MomentPublished{}.Name())
	register(moment.MomentUnpublished{}.Name())
	register(moment.MomentDeleted{}.Name())
}

func RegisterFriendLinkSubscribers(bus appEvent.Bus, service *Service) {
	if bus == nil || service == nil {
		return
	}

	register := func(eventName string) {
		bus.Subscribe(eventName, handlerFunc(func(ctx context.Context, _ appEvent.Event) error {
			return service.Invalidate(ctx, []string{"friend:list"}, []string{"/friends"})
		}))
	}

	register("friendlink.application.approved")
	register("friendlink.application.rejected")
	register("friendlink.application.blocked")
	register("friendlink.link.changed")
}

func RegisterFriendTimelineSubscribers(bus appEvent.Bus, service *Service) {
	if bus == nil || service == nil {
		return
	}

	bus.Subscribe(federation.FederatedPostsCached{}.Name(), handlerFunc(func(ctx context.Context, _ appEvent.Event) error {
		deps := []string{
			"friend-timeline:list:page:*",
		}
		urls := []string{
			"/friends-timeline",
		}
		return service.Invalidate(ctx, deps, urls)
	}))
}

func RegisterLayoutSubscribers(bus appEvent.Bus, service *Service) {
	if bus == nil || service == nil {
		return
	}

	bus.Subscribe("sysconfig.updated", handlerFunc(func(ctx context.Context, event appEvent.Event) error {
		generic, ok := event.(appEvent.Generic)
		if !ok {
			return nil
		}
		keys, _ := generic.Payload["Keys"].([]string)
		for _, k := range keys {
			if len(k) > 5 && k[:5] == "site." {
				return service.Invalidate(ctx, []string{"layout:website-info"}, nil)
			}
		}
		return nil
	}))
}

func RegisterAlbumSubscribers(bus appEvent.Bus, service *Service) {
	if bus == nil || service == nil {
		return
	}

	register := func(eventName string) {
		bus.Subscribe(eventName, handlerFunc(func(ctx context.Context, event appEvent.Event) error {
			albumID, shortURL := extractAlbumEventPayload(event)
			if albumID <= 0 {
				return nil
			}

			deps := []string{
				"album:list:page:*",
				fmt.Sprintf("album:detail:%d", albumID),
			}
			urls := []string{"/albums"}
			// Resolve photo pages too: newly added photos have URLs that were
			// never rendered, so dep invalidation alone cannot reach them.
			urls = append(urls, service.AlbumURLs(ctx, shortURL)...)
			return service.Invalidate(ctx, deps, urls)
		}))
	}

	register(appalbum.AlbumCreated{}.Name())
	register(appalbum.AlbumUpdated{}.Name())
	register(appalbum.AlbumPublished{}.Name())
	register(appalbum.AlbumUnpublished{}.Name())
	register(appalbum.AlbumDeleted{}.Name())
}

func extractAlbumEventPayload(event appEvent.Event) (albumID int64, shortURL string) {
	switch e := event.(type) {
	case appalbum.AlbumCreated:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case appalbum.AlbumUpdated:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case appalbum.AlbumPublished:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case appalbum.AlbumUnpublished:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case appalbum.AlbumDeleted:
		return e.ID, strings.TrimSpace(e.ShortURL)
	default:
		return 0, ""
	}
}

func extractMomentEventPayload(event appEvent.Event) (momentID int64, shortURL string) {
	switch e := event.(type) {
	case moment.MomentCreated:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case moment.MomentUpdated:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case moment.MomentPublished:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case moment.MomentUnpublished:
		return e.ID, strings.TrimSpace(e.ShortURL)
	case moment.MomentDeleted:
		return e.ID, strings.TrimSpace(e.ShortURL)
	default:
		return 0, ""
	}
}

// RegisterTagContentCacheSubscribers subscribes to moment CRUD events
// and clears all tag:contents:* Redis keys so that the tag content API
// rebuilds its cache on the next request.
func RegisterTagContentCacheSubscribers(bus appEvent.Bus, redisClient *redis.Client, redisPrefix string) {
	if bus == nil || redisClient == nil {
		return
	}

	invalidate := handlerFunc(func(ctx context.Context, _ appEvent.Event) error {
		pattern := fmt.Sprintf("%stag:contents:*", redisPrefix)
		var cursor uint64
		for {
			keys, next, err := redisClient.Scan(ctx, cursor, pattern, 200).Result()
			if err != nil {
				log.Printf("tag content cache scan error: %v", err)
				return nil
			}
			if len(keys) > 0 {
				if err := redisClient.Del(ctx, keys...).Err(); err != nil {
					log.Printf("tag content cache del error: %v", err)
				}
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
		return nil
	})

	// Moment events
	bus.Subscribe(moment.MomentCreated{}.Name(), invalidate)
	bus.Subscribe(moment.MomentUpdated{}.Name(), invalidate)
	bus.Subscribe(moment.MomentPublished{}.Name(), invalidate)
	bus.Subscribe(moment.MomentUnpublished{}.Name(), invalidate)
	bus.Subscribe(moment.MomentDeleted{}.Name(), invalidate)
}
