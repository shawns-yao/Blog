package content

import "encoding/json"

const (
	KindNote         = "note"
	KindArticle      = "article"
	KindUnclassified = "unclassified"
)

// ContentKind leaves historical content unclassified until its owner chooses a type.
func ContentKind(extInfo []byte) string {
	var info struct {
		ContentKind string `json:"contentKind"`
	}
	if json.Unmarshal(extInfo, &info) == nil &&
		(info.ContentKind == KindNote || info.ContentKind == KindArticle) {
		return info.ContentKind
	}
	return KindUnclassified
}

func ValidContentKindFilter(kind string) bool {
	return kind == "" || kind == KindNote || kind == KindArticle || kind == KindUnclassified
}
