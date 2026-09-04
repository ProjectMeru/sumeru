package web

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

type navActionKind int

const (
	navActionUnknown navActionKind = iota
	navActionWindow
	navActionURL
)

type navigationAction struct {
	kind       navActionKind
	url        string
	windowData map[string]interface{}
}

func resolveNavigationAction(ctx context.Context, actionID int, actionQuery string) (navigationAction, error) {
	modelName, coreID, err := orm.ResolveActionRecord(ctx, actionID, actionQuery)
	if err != nil {
		return navigationAction{}, err
	}
	switch modelName {
	case sysActionURLModel:
		row, err := orm.SearchOne(ctx, sysActionURLModel, map[string]interface{}{"id": coreID})
		if err != nil {
			return navigationAction{}, fmt.Errorf("action %d not found", coreID)
		}
		url := strings.TrimSpace(orm.AsString(row["url"]))
		if url == "" {
			return navigationAction{}, fmt.Errorf("action %d has empty url", coreID)
		}
		safe := render.SafeIframeURL(url)
		if config.AppConfig.DevMode {
			safe = render.SafeIframeURLAllowHTTP(url)
		}
		if !safe {
			return navigationAction{}, fmt.Errorf("action %d has unsafe url", coreID)
		}
		return navigationAction{kind: navActionURL, url: url}, nil
	case sysActionWindowModel:
		row, err := loadWindowAction(ctx, coreID)
		if err != nil {
			return navigationAction{}, fmt.Errorf("action %d not found", coreID)
		}
		return navigationAction{kind: navActionWindow, windowData: row}, nil
	default:
		return navigationAction{}, fmt.Errorf("unsupported action model %q", modelName)
	}
}
