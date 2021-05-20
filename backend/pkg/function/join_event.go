package function

import (
	"context"
	"log"

	"github.com/gcp-kit/line-bot-gcp-go/gcpine"
	"github.com/go-utils/caller"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// JoinEvent - handle join events
func JoinEvent(context.Context, *gcpine.GCPine, *linebot.Event) ([]linebot.SendingMessage, error) {
	name := caller.GetCallFuncName()
	log.Println("Call:", name)

	return nil, nil
}
