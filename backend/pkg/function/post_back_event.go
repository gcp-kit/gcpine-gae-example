package function

import (
	"context"
	"fmt"
	"log"

	"github.com/gcp-kit/line-bot-gcp-go/gcpine"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// PostBackEvent - handle post back events
func PostBackEvent(ctx context.Context, _ *gcpine.GCPine, event *linebot.Event) ([]linebot.SendingMessage, error) {
	fmt.Println("PostBackEvent")
	var (
		uid  = event.Source.UserID
		data = event.Postback.Data
	)
	log.Println("UID:", uid)
	log.Println("PostBackData:", data)
	return nil, nil
}
