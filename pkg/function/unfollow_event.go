package function

import (
	"context"
	"log"

	"github.com/gcp-kit/line-bot-gcp-go/gcpine"
	"github.com/go-utils/caller"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// UnfollowEvent - handle unfollow events
func UnfollowEvent(_ context.Context, _ *gcpine.GCPine, event *linebot.Event) ([]linebot.SendingMessage, error) {
	name := caller.GetCallFuncName()
	log.Println("Call:", name)

	uid := event.Source.UserID
	log.Println("UID:", uid)

	return nil, nil
}
