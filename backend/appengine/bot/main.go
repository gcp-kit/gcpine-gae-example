package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/gcp-kit/gcpen"
	"github.com/gcp-kit/gcpine-gae-example/backend/pkg/config"
	"github.com/gcp-kit/gcpine-gae-example/backend/pkg/ctxkeys"
	"github.com/gcp-kit/gcpine-gae-example/backend/pkg/environ"
	"github.com/gcp-kit/gcpine-gae-example/backend/pkg/function" // nolint: typecheck
	"github.com/gcp-kit/gcpine-gae-example/backend/pkg/secret"
	"github.com/gcp-kit/line-bot-gcp-go/gcpine"
	"github.com/gcp-kit/stalog"
	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

const locationID = "asia-northeast1"

func main() {
	environ.IsTest = false
	defer func() {
		if rec := recover(); rec != nil {
			debug.PrintStack()
		}
	}()

	e := echo.New()
	e.HideBanner = true

	ctx := context.Background()

	cloudTasksClient, err := cloudtasks.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to initialize cloudtasks client: %+v", err)
	}

	cfg := new(config.Config)
	{
		smClient, err := secretmanager.NewClient(ctx)
		if err != nil {
			log.Fatalf("failed to initialize secretmanager client: %+v", err)
		}
		smProvider := secret.NewProvider(smClient, gcpen.ProjectID)
		cfg.ChannelSecret = smProvider.GetSecret(ctx, "CHANNEL_SECRET")
		cfg.ChannelAccessToken = smProvider.GetSecret(ctx, "CHANNEL_ACCESS_TOKEN")
		if cfg.ChannelSecret == "" || cfg.ChannelAccessToken == "" {
			log.Fatalf("secret and token are required")
		}
		ctx = context.WithValue(ctx, ctxkeys.ConfigKey{}, cfg)
	}

	lineClient, err := linebot.New(cfg.ChannelSecret, cfg.ChannelAccessToken)
	if err != nil {
		log.Fatalf("failed to initialize linebot client: %+v", err)
	}

	queuePath := filepath.Join("projects", gcpen.ProjectID, "locations", locationID, "queues")

	g := e.Group("/line/")
	{
		props := gcpine.NewAppEngineProps(
			cloudTasksClient,
			filepath.Join(queuePath, "parent"),
			"/line/tq/parent",
		)
		props.SetSecret(cfg.ChannelSecret)

		g.POST("webhook", func(c echo.Context) error {
			return props.ReceiveWebHook(c.Request(), c.Response().Writer)
		})

		tq := g.Group("tq/")
		{
			tq.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c echo.Context) error {
					t, ok := c.Request().Header["X-Appengine-Taskname"]
					if !ok || len(t[0]) == 0 {
						log.Println("Invalid Task: No X-Appengine-Taskname request header found")
						return c.String(http.StatusBadRequest, "Bad Request - Invalid Task\n")
					}

					var queueName string
					if q, ok := c.Request().Header["X-Appengine-Queuename"]; ok {
						queueName = q[0]
					}

					fmt.Printf("Completed task: task queue(%s), task name(%s)\n", queueName, t[0])
					return next(c)
				}
			})

			props := gcpine.NewAppEngineProps(
				cloudTasksClient,
				filepath.Join(queuePath, "child"),
				"/line/tq/child",
			)

			pine := newPine(lineClient)
			props.SetGCPine(pine)

			config.BotInfo, err = pine.GetBotInfo().WithContext(ctx).Do()
			if err != nil {
				log.Fatalf("failed to get bot info: %+v", err)
			}

			tq.POST("parent", func(c echo.Context) error {
				body, err := ioutil.ReadAll(c.Request().Body)
				if err != nil {
					return err
				}
				return props.ParentEvent(ctx, body)
			})

			tq.POST("child", func(c echo.Context) error {
				logger := stalog.RequestContextLogger(c.Request())
				body, err := ioutil.ReadAll(c.Request().Body)
				if err != nil {
					logger.Errorf("error: %+v", err)
					return err
				}
				return props.ChildEvent(ctx, body)
			})
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Use(
		stalog.RequestLoggingWithEcho(newStalogConfig()),
	)

	e.Logger.Fatal(e.Start(":" + port))
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

func newStalogConfig(severities ...stalog.Severity) *stalog.Config {
	severity := stalog.SeverityDefault
	if len(severities) > 0 {
		severity = severities[0]
	}

	cfg := stalog.NewConfig(gcpen.ProjectID)
	cfg.RequestLogOut = os.Stderr               // request log to stderr
	cfg.ContextLogOut = os.Stdout               // context log to stdout
	cfg.Severity = severity                     // only over variable `severity` logs are logged
	cfg.AdditionalData = stalog.AdditionalData{ // set additional fields for all logs
		"service": gcpen.ServiceName,
		"version": gcpen.ServiceVersion,
	}

	return cfg
}
