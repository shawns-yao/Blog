package content

// MomentListOptions 公开的手记列表查询选项
type MomentListOptions struct {
	Page     int     `json:"page" validate:"min=1"`
	PageSize int     `json:"pageSize" validate:"min=1,max=100"`
	ColumnID *int64  `json:"columnId,omitempty"`
	TopicID  *int64  `json:"topicId,omitempty"`
	AuthorID *int64  `json:"authorId,omitempty"`
	Search   *string `json:"search,omitempty"`
}

// MomentListOptionsInternal 内部的手记列表查询选项（包含管理功能）
type MomentListOptionsInternal struct {
	Page      int     `json:"page" validate:"min=1"`
	PageSize  int     `json:"pageSize" validate:"min=1,max=100"`
	ColumnID  *int64  `json:"columnId,omitempty"`
	TopicID   *int64  `json:"topicId,omitempty"`
	AuthorID  *int64  `json:"authorId,omitempty"`
	Published *bool   `json:"published,omitempty"` // 仅管理员可用
	Search    *string `json:"search,omitempty"`
}
