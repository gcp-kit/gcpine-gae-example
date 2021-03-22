package function

import (
	"context"
	"fmt"

	"github.com/gcp-kit/line-bot-gcp-go/gcpine"
	"github.com/line/line-bot-sdk-go/linebot"
	"golang.org/x/xerrors"
)

// LocationEvent - handle location message events
func LocationEvent(_ context.Context, _ *gcpine.GCPine, event *linebot.Event) ([]linebot.SendingMessage, error) {
	message, ok := event.Message.(*linebot.LocationMessage)
	if !ok {
		return nil, xerrors.New("couldn't cast")
	}

	text := fmt.Sprintf("Latitude: %f\nLongitude: %f", message.Latitude, message.Longitude)
	stack := []linebot.SendingMessage{
		linebot.NewTextMessage(text),
	}

	return stack, nil
}
