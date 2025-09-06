package services

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

type LogService struct{}

func (l *LogService) LogError(message string) {
	application.Get().Logger.Error(message)
}
