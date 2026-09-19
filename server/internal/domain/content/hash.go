package content

import (
	"crypto/md5"
	"encoding/hex"
)

func MomentContentHash(title string, summary string, content string) string {
	return hashContentParts(title, summary, content)
}

func hashContentParts(parts ...string) string {
	hasher := md5.New()
	for i, part := range parts {
		if i > 0 {
			_, _ = hasher.Write([]byte{0})
		}
		_, _ = hasher.Write([]byte(part))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}
