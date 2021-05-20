package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gcp-kit/gcpine-gae-example/backend/pkg/config"
	"github.com/gcp-kit/gcpine-gae-example/backend/pkg/ctxkeys"
	"github.com/gcp-kit/gcpine-gae-example/backend/pkg/environ"
	"github.com/gcp-kit/gcpine-gae-example/backend/pkg/function"
	"github.com/gcp-kit/line-bot-gcp-go/gcpine"
	"github.com/heetch/confita"
	"github.com/joho/godotenv"
	"github.com/k0kubun/pp"
	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"golang.org/x/xerrors"
)

func main() {
	environ.IsTest = false
	defer func() {
		if rec := recover(); rec != nil {
			debug.PrintStack()
		}
	}()

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatalf("failed to read .env file: %+v", err)
	}

	cfg := new(config.Config)
	ctx := context.Background()
	{
		loader := confita.NewLoader()
		if err := loader.Load(context.TODO(), cfg); err != nil {
			log.Fatalf("failed to load config: %+v", err)
		}
		if cfg.ChannelSecret == "" || cfg.ChannelAccessToken == "" {
			log.Fatalf("secret and token are required")
		}
		ctx = context.WithValue(ctx, ctxkeys.ConfigKey{}, cfg)
	}

	lineClient, err := linebot.New(cfg.ChannelSecret, cfg.ChannelAccessToken)
	if err != nil {
		log.Fatalf("failed to initialize linebot client: %+v", err)
	}

	pine := newPine(lineClient)

	config.BotInfo, err = pine.GetBotInfo().WithContext(ctx).Do()
	if err != nil {
		log.Fatalf("failed to get bot info: %+v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.POST("webhook", func(c echo.Context) error {
		events, err := pine.ParseRequest(c.Request())
		if err != nil {
			if xerrors.Is(err, linebot.ErrInvalidSignature) {
				log.Print(err)
			}
			return c.String(http.StatusBadRequest, "NG")
		}

		var wg sync.WaitGroup
		for _, event := range events {
			pp.Println(event)
			// nolint
			wg.Add(1)
			go func(ev *linebot.Event) {
				defer wg.Done()
				err = pine.Execute(ctx, ev)
				if err != nil {
					if len(pine.ErrMessages) > 0 {
						if er := pine.SendReplyMessage(ev.ReplyToken, pine.ErrMessages); er != nil {
							log.Printf("failed to send reply: %+v", err)
						}
					}
					log.Printf("failed to execute: %+v", err)
				}
			}(event)
		}
		wg.Wait()
		return c.String(http.StatusOK, "OK")
	})

	go func() {
		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 10 seconds.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
	}()

	e.Logger.Fatal(e.Start(":8080"))
}

func newPine(client *linebot.Client) *gcpine.GCPine {
	functionMap := map[gcpine.EventType]gcpine.PineFunction{
		gcpine.EventTypeFollowEvent:       function.FollowEvent,
		gcpine.EventTypeUnfollowEvent:     function.UnfollowEvent,
		gcpine.EventTypeTextMessage:       function.TextEvent,
		gcpine.EventTypeLocationMessage:   function.LocationEvent,
		gcpine.EventTypePostBackEvent:     function.PostBackEvent,
		gcpine.EventTypeJoinEvent:         function.JoinEvent,
		gcpine.EventTypeMemberJoinedEvent: function.MemberJoinedEvent,
		gcpine.EventTypeMemberLeftEvent:   function.MemberLeftEvent,
		gcpine.EventTypeLeaveEvent:        function.LeaveEvent,
	}
	systemError := linebot.NewTextMessage("System error.")

	return &gcpine.GCPine{
		ErrMessages: []linebot.SendingMessage{systemError},
		Function:    functionMap,
		Client:      client,
	}
}
