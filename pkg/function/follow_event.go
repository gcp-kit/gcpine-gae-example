package function

import (
	"context"
	"fmt"
	"log"

	"github.com/gcp-kit/line-bot-gcp-go/gcpine"
	"github.com/go-utils/caller"
	"github.com/line/line-bot-sdk-go/linebot"
	"golang.org/x/xerrors"
)

// FollowEvent - handle follow events
func FollowEvent(ctx context.Context, pine *gcpine.GCPine, event *linebot.Event) ([]linebot.SendingMessage, error) {
	name := caller.GetCallFuncName()
	log.Println("Call:", name)

	uid := event.Source.UserID
	log.Println("UID:", uid)

	prof, err := pine.GetProfile(uid).WithContext(ctx).Do()
	if err != nil {
		log.Println("Error:", err.Error())
		return nil, xerrors.Errorf("error in GetProfile method: %w", err)
	}

	log.Println("Name:", prof.DisplayName)
	log.Println("Picture:", prof.PictureURL)

	text := fmt.Sprintf("%s, thanks for following!", prof.DisplayName)
	stack := []linebot.SendingMessage{
		linebot.NewTextMessage(text),
	}

	return stack, nil
}
