package contract

import "time"

// SamePeriodMomentItemResp 同一时期手记项。
type SamePeriodMomentItemResp struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	ShortURL  string    `json:"shortUrl"`
	Summary   string    `json:"summary"`
	Cover     *string   `json:"cover,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// SamePeriodMomentListResp 同一时期手记列表响应。
type SamePeriodMomentListResp struct {
	Items []SamePeriodMomentItemResp `json:"items"`
}
