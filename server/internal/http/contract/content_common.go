package contract

// TOCNode 目录节点。
type TOCNode struct {
	Name     string    `json:"name"`
	Anchor   string    `json:"anchor"`
	Children []TOCNode `json:"children,omitempty"`
}

// TagResp 标签响应。
type TagResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// MetricsResp 指标响应。
type MetricsResp struct {
	Views    int64 `json:"views"`
	Likes    int   `json:"likes"`
	Comments int   `json:"comments"`
}
