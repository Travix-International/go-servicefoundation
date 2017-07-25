package model

import "github.com/Travix-International/logger"

type Logger interface {
	Debug(event, formatOrMsg string, a ...interface{}) error
	Info(event, formatOrMsg string, a ...interface{}) error
	Warn(event, formatOrMsg string, a ...interface{}) error
	Error(event, formatOrMsg string, a ...interface{}) error
	GetLogger() *logger.Logger
}
