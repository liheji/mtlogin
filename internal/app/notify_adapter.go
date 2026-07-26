package app

import (
	"mtlogin/internal/refresh"
	"mtlogin/pkg/apikit/apinotify"
	"mtlogin/pkg/log"
)

type notifyAdapter struct {
	manager *apinotify.Manager
}

func (a notifyAdapter) Notify(ctx *log.Context, msg refresh.Message) error {
	level := apinotify.LevelInfo
	switch msg.Level {
	case refresh.LevelSuccess:
		level = apinotify.LevelSuccess
	case refresh.LevelWarning:
		level = apinotify.LevelWarning
	case refresh.LevelError:
		level = apinotify.LevelError
	}
	return a.manager.Notify(ctx, apinotify.Message{Title: msg.Title, Body: msg.Body, Level: level})
}
