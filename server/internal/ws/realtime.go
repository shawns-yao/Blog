package ws

import (
	"fmt"
	"strings"
)

const realtimeRoomKey = "realtime:lobby"

func RealtimeRoomKey() string {
	return realtimeRoomKey
}

func ContentRoomKey(contentType string, contentID int64) (string, bool) {
	if contentID <= 0 {
		return "", false
	}

	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "moment":
		return fmt.Sprintf("moment:%d", contentID), true
	default:
		return "", false
	}
}
