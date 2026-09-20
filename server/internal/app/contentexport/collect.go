package contentexport

import (
	"context"
	"time"

	"github.com/shawns-yao/shawn-blog/server/internal/app/moment"
	"github.com/shawns-yao/shawn-blog/server/internal/app/sysconfig"
	"github.com/shawns-yao/shawn-blog/server/internal/app/taxonomy"
	"github.com/shawns-yao/shawn-blog/server/internal/domain/content"
)

// exportPageSize 是批量加载内容时的分页大小。必须 >= 1（GORM Limit(0) 会返回空集）。
const exportPageSize = 100

// Snapshot 是一次导出任务加载到的全量内容快照。
type Snapshot struct {
	Moments  []*content.Moment
	Columns  []*content.MomentColumn
	Tags     []*content.Tag
	SiteTZ   *time.Location
	SiteName string
	SiteURL  string
	// PublicHost 是站点 public_url 的主机名，用于把手记里的
	// https://<本站域名>/uploads/... 绝对引用判定为站内图片。
	PublicHost string
}

// Collector 通过现有内容应用服务批量加载全部内容（含未发布/禁用）。
type Collector struct {
	momentSvc *moment.Service
	columnSvc *taxonomy.ColumnService
	tagSvc    *taxonomy.TagService
	sysCfg    *sysconfig.Service
}

func NewCollector(
	momentSvc *moment.Service,
	columnSvc *taxonomy.ColumnService,
	tagSvc *taxonomy.TagService,
	sysCfg *sysconfig.Service,
) *Collector {
	return &Collector{
		momentSvc: momentSvc,
		columnSvc: columnSvc,
		tagSvc:    tagSvc,
		sysCfg:    sysCfg,
	}
}

// Collect 加载 admin 全集：Published/Enabled/Builtin 过滤一律传 nil。
func (c *Collector) Collect(ctx context.Context) (*Snapshot, error) {
	snap := &Snapshot{
		Moments: make([]*content.Moment, 0),
		Columns: make([]*content.MomentColumn, 0),
		Tags:    make([]*content.Tag, 0),
		SiteTZ:  time.UTC,
	}

	for pg := 1; ; pg++ {
		items, _, err := c.momentSvc.ListMoments(ctx, content.MomentListOptionsInternal{Page: pg, PageSize: exportPageSize})
		if err != nil {
			return nil, err
		}
		snap.Moments = append(snap.Moments, items...)
		if len(items) < exportPageSize {
			break
		}
	}

	columns, err := c.columnSvc.List(ctx)
	if err != nil {
		return nil, err
	}
	snap.Columns = columns
	tags, err := c.tagSvc.List(ctx)
	if err != nil {
		return nil, err
	}
	snap.Tags = tags

	if c.sysCfg != nil {
		snap.SiteTZ = c.sysCfg.Timezone(ctx)
		if info, infoErr := c.sysCfg.WebsiteInfo(ctx); infoErr == nil {
			snap.SiteName = info["website_name"]
			snap.SiteURL = info["public_url"]
			snap.PublicHost = hostFromURL(info["public_url"])
		}
	}

	return snap, nil
}
