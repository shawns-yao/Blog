package ws

import (
	"context"
	"net/url"
	"path"
	"strings"

	"github.com/shawns-yao/shawn-blog/server/internal/app/sysconfig"
	"github.com/shawns-yao/shawn-blog/server/internal/domain/content"
)

const (
	presenceTypeMoment = "moment"
	presenceTypePage   = "page"
)

type PresenceTitleResolver struct {
	contentRepo content.Repository
	sysCfg      *sysconfig.Service
}

func NewPresenceTitleResolver(contentRepo content.Repository, sysCfg *sysconfig.Service) *PresenceTitleResolver {
	return &PresenceTitleResolver{
		contentRepo: contentRepo,
		sysCfg:      sysCfg,
	}
}

func (r *PresenceTitleResolver) Resolve(contentType string, rawURL string) (PresenceResolvedView, bool) {
	normalizedType := normalizePresenceType(contentType)
	if normalizedType == "" {
		return PresenceResolvedView{}, false
	}

	normalizedPath := normalizePresencePath(rawURL)
	switch normalizedType {
	case presenceTypeMoment:
		return r.resolveMoment(normalizedPath), true
	case presenceTypePage:
		return r.resolvePage(normalizedPath), true
	default:
		return PresenceResolvedView{}, false
	}
}

func normalizePresenceType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case presenceTypeMoment:
		return presenceTypeMoment
	case presenceTypePage:
		return presenceTypePage
	default:
		return ""
	}
}

func normalizePresencePath(rawURL string) string {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return "/"
	}

	parsed, err := url.Parse(value)
	if err == nil && parsed.Path != "" {
		value = parsed.Path
	}

	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}

	normalized := path.Clean(value)
	if normalized == "." {
		return "/"
	}
	return normalized
}

func splitPathSegments(pathname string) []string {
	trimmed := strings.Trim(pathname, "/")
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		decoded, err := url.PathUnescape(part)
		if err == nil {
			out = append(out, decoded)
			continue
		}
		out = append(out, part)
	}
	return out
}

func (r *PresenceTitleResolver) resolveMoment(pathname string) PresenceResolvedView {
	view := PresenceResolvedView{
		ContentType: presenceTypeMoment,
		Title:       "手记",
		URL:         pathname,
	}

	parts := splitPathSegments(pathname)
	if len(parts) < 2 || parts[0] != "moments" {
		return view
	}

	slug := parts[len(parts)-1]
	if slug == "" || slug == "moments" {
		return view
	}

	if r.contentRepo == nil {
		return view
	}

	item, err := r.contentRepo.GetMomentByShortURL(context.Background(), slug)
	if err != nil || item == nil || !item.IsPublished {
		return view
	}

	view.Title = strings.TrimSpace(item.Title)
	if view.Title == "" {
		view.Title = "手记"
	}
	siteTZ := r.sysCfg.Timezone(context.Background())
	view.URL = buildMomentPath(item.ShortURL, item.CreatedAt.In(siteTZ))
	return view
}

// resolvePage 覆盖首页与其余非内容详情页（标签、友链、时间轴等）。
// 页面内容模型已移除，此处不再按短链反查标题，统一回落到通用标题。
func (r *PresenceTitleResolver) resolvePage(pathname string) PresenceResolvedView {
	view := PresenceResolvedView{
		ContentType: presenceTypePage,
		Title:       "页面",
		URL:         pathname,
	}

	if pathname == "/" {
		view.Title = "首页"
		view.URL = "/"
		return view
	}

	parts := splitPathSegments(pathname)
	if len(parts) == 1 {
		view.URL = "/" + url.PathEscape(parts[0])
	}

	return view
}

func buildMomentPath(slug string, createdAt interface{ Format(string) string }) string {
	date := createdAt.Format("2006/01/02")
	return "/moments/" + date + "/" + url.PathEscape(slug)
}
