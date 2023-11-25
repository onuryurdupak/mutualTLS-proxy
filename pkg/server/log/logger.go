package log

import (
	gmlog "github.com/onuryurdupak/gomod/v2/log"
)

type logger struct {
}

func NewServerLogger() *logger {
	return &logger{}
}

func (l *logger) Write(p []byte) (n int, err error) {
	logger := gmlog.NewLogger("Server/Logger", "N/A")
	logger.Errorf(string(p))

	return len(p), nil
}
