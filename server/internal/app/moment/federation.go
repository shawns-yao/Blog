package moment

import (
	"context"
	"log"
	"time"

	appEvent "github.com/shawns-yao/grtblog-v2/server/internal/app/event"
	appfed "github.com/shawns-yao/grtblog-v2/server/internal/app/federation"
	"github.com/shawns-yao/grtblog-v2/server/internal/domain/content"
)

func publishFederationSignals(ctx context.Context, bus appEvent.Bus, moment *content.Moment, contentBody string) {
	if bus == nil || moment == nil {
		return
	}
	mentions, citations := appfed.ParseSignals(contentBody)
	if len(mentions) == 0 && len(citations) == 0 {
		return
	}
	deliveredMentions, deliveredCitations := deliveredSignalKeys(moment.ExtInfo)
	now := time.Now()
	newMentionKeys := make([]string, 0, len(mentions))
	newCitationKeys := make([]string, 0, len(citations))
	for _, mention := range mentions {
		key := mention.User + "@" + mention.Instance
		if _, exists := deliveredMentions[key]; exists {
			continue
		}
		// Only record the signal as delivered when the dispatch handler
		// succeeded; otherwise the marker would be silently lost forever.
		if err := bus.Publish(ctx, appfed.MentionDetected{
			MomentID:        moment.ID,
			AuthorID:        moment.AuthorID,
			Title:           moment.Title,
			ShortURL:        moment.ShortURL,
			MomentCreatedAt: moment.CreatedAt,
			TargetUser:      mention.User,
			TargetInstance:  mention.Instance,
			Context:         mention.Context,
			MentionType:     "",
			At:              now,
		}); err != nil {
			log.Printf("[federation] 提及事件发布失败 moment_id=%d target=%s@%s err=%v", moment.ID, mention.User, mention.Instance, err)
			continue
		}
		newMentionKeys = append(newMentionKeys, key)
	}
	for _, citation := range citations {
		key := citation.Instance + "|" + citation.PostID
		if _, exists := deliveredCitations[key]; exists {
			continue
		}
		if err := bus.Publish(ctx, appfed.CitationDetected{
			MomentID:        moment.ID,
			AuthorID:        moment.AuthorID,
			Title:           moment.Title,
			ShortURL:        moment.ShortURL,
			MomentCreatedAt: moment.CreatedAt,
			TargetInstance:  citation.Instance,
			TargetPostID:    citation.PostID,
			Context:         citation.Context,
			CitationType:    "",
			At:              now,
		}); err != nil {
			log.Printf("[federation] 引用事件发布失败 moment_id=%d target=%s|%s err=%v", moment.ID, citation.Instance, citation.PostID, err)
			continue
		}
		newCitationKeys = append(newCitationKeys, key)
	}
	if len(newMentionKeys) == 0 && len(newCitationKeys) == 0 {
		return
	}
	if updated, changed := markDeliveredSignals(moment.ExtInfo, newMentionKeys, newCitationKeys); changed {
		moment.ExtInfo = updated
	}
}
