package contract

import (
	"time"

	"github.com/shawns-yao/shawn-blog/server/internal/http/response"
)

type ColumnResp struct {
	ParentID  *int64    `json:"parentId"`
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	ShortURL  string    `json:"shortUrl"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TagItemResp struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TagPublicResp struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	MomentCount int64  `json:"momentCount"`
}

// EmptyRespEnvelope 仅用于 swagger 展示。
type EmptyRespEnvelope struct {
	Code   int           `json:"code"`
	BizErr string        `json:"bizErr"`
	Msg    string        `json:"msg"`
	Data   any           `json:"data"`
	Meta   response.Meta `json:"meta"`
}
