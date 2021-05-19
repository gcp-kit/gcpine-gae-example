package function

import (
	"context"
	"log"

	"github.com/gcp-kit/line-bot-gcp-go/gcpine"
	"github.com/go-utils/caller"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"golang.org/x/xerrors"
)

// TextEvent - handle text message events
func TextEvent(ctx context.Context, pine *gcpine.GCPine, event *linebot.Event) ([]linebot.SendingMessage, error) {
	name := caller.GetCallFuncName()
	log.Println("Call:", name)

	message, ok := event.Message.(*linebot.TextMessage)
	if !ok {
		return nil, xerrors.New("couldn't cast")
	}

	text := message.Text

	replyMessage := linebot.SendingMessage(linebot.NewTextMessage(text))

	if text == "ping" {
		items := &linebot.QuickReplyItems{
			Items: []*linebot.QuickReplyButton{
				{
					Action: linebot.QuickReplyAction(&linebot.MessageAction{
						Label: "ping",
						Text:  "ping",
					}),
				},
			},
		}

		prof, err := pine.GetProfile(event.Source.UserID).WithContext(ctx).Do()
		if err != nil {
			return nil, xerrors.Errorf("error in GetProfile method: %w", err)
		}

		sender := &linebot.Sender{
			Name:    prof.DisplayName,
			IconURL: prof.PictureURL,
		}

		replyMessage = linebot.NewTextMessage("pong").
			WithQuickReplies(items).
			WithSender(sender)
	}

	stack := []linebot.SendingMessage{
		replyMessage,
	}

	return stack, nil
}
