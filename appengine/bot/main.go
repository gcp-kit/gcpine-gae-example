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
	"github.com/gcp-kit/gcpen"
	"github.com/gcp-kit/gcpine-gae-example/pkg/function" // nolint: typecheck
	"github.com/gcp-kit/line-bot-gcp-go/gcpine"
	"github.com/gcp-kit/stalog"
	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

const locationID = "asia-northeast1"

func main() {
	defer func() {
		if rec := recover(); rec != nil {
			debug.PrintStack()
		}
	}()

	e := echo.New()
	e.HideBanner = true

	ctx := context.Background()

	cloudtasksClient, err := cloudtasks.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to initialize cloudtasks client: %+v", err)
	}

	lineSecret := os.Getenv(gcpine.EnvKeyChannelSecret)
	token := os.Getenv(gcpine.EnvKeyChannelAccessToken)

	lineClient, err := linebot.New(lineSecret, token)
	if err != nil {
		log.Fatalf("failed to initialize linebot client: %+v", err)
	}

	queuePath := filepath.Join("projects", gcpen.ProjectID, "locations", locationID, "queues")

	var (
		parentQueue = fmt.Sprintf("%s-parent", gcpen.ProjectID)
		childQueue  = fmt.Sprintf("%s-child", gcpen.ProjectID)
	)

	g := e.Group("/line/")
	{
		props := gcpine.NewAppEngineProps(
			cloudtasksClient,
			filepath.Join(queuePath, parentQueue),
			"/line/tq/parent",
		)
		props.SetSecret(lineSecret)

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
				cloudtasksClient,
				filepath.Join(queuePath, childQueue),
				"/line/tq/child",
			)
			props.SetGCPine(newPine(lineClient))

			tq.POST("parent", func(c echo.Context) error {
				body, err := ioutil.ReadAll(c.Request().Body)
				if err != nil {
					return err
				}
				return props.ParentEvent(ctx, body)
			})

			tq.POST("child", func(c echo.Context) error {
				body, err := ioutil.ReadAll(c.Request().Body)
				if err != nil {
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
		gcpine.EventTypeFollowEvent:     function.FollowEvent,
		gcpine.EventTypeUnfollowEvent:   function.UnfollowEvent,
		gcpine.EventTypeTextMessage:     function.TextEvent,
		gcpine.EventTypeLocationMessage: function.LocationEvent,
	}
	systemError := linebot.NewTextMessage("システムエラーです。")

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

	config := stalog.NewConfig(gcpen.ProjectID)
	config.RequestLogOut = os.Stderr               // request log to stderr
	config.ContextLogOut = os.Stdout               // context log to stdout
	config.Severity = severity                     // only over variable `severity` logs are logged
	config.AdditionalData = stalog.AdditionalData{ // set additional fields for all logs
		"service": gcpen.ServiceName,
		"version": gcpen.ServiceVersion,
	}

	return config
}
