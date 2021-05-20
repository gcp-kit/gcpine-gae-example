package component

import (
	"fmt"

	"github.com/gcp-kit/gcpine-gae-example/backend/pkg/environ"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func separator(marginType ...linebot.FlexComponentMarginType) *linebot.SeparatorComponent {
	margin := linebot.FlexComponentMarginTypeMd
	if len(marginType) > 0 {
		margin = marginType[0]
	}
	return &linebot.SeparatorComponent{
		Type:   linebot.FlexComponentTypeSeparator,
		Margin: margin,
	}
}

// nolint
var footer = func() *linebot.BoxComponent {
	return &linebot.BoxComponent{
		Type:   linebot.FlexComponentTypeBox,
		Layout: linebot.FlexBoxLayoutTypeVertical,
		Margin: linebot.FlexComponentMarginTypeMd,
		Contents: []linebot.FlexComponent{
			separator(),
			&linebot.TextComponent{
				Type: linebot.FlexComponentTypeText,
				Text: func() string {
					branchName := environ.BranchName
					switch branchName {
					case "main", "":
						branchName = ""
					default:
						branchName = fmt.Sprintf("%s / ", branchName)
					}
					return branchName + environ.CommitHash
				}(),
				Color:   "#f0f0f0",
				Align:   linebot.FlexComponentAlignTypeEnd,
				Gravity: linebot.FlexComponentGravityTypeBottom,
				Margin:  linebot.FlexComponentMarginTypeSm,
				Size:    linebot.FlexTextSizeTypeXxs,
				Weight:  linebot.FlexTextWeightTypeBold,
				Style:   linebot.FlexTextStyleTypeItalic,
			},
		},
	}
}()
