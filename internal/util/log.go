package util

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func GroupLogger(group string) *logrus.Entry {
	logger := logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   false,
		DisableQuote:    true,
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceQuote:      false,
		PadLevelText:    true,
	})

	logger.Formatter = &customFormatter{group: group}

	return logger.WithField("group", group)
}

type customFormatter struct {
	group string
}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	level := entry.Level.String()
	msg := entry.Message
	return []byte(fmt.Sprintf("[%s] [%s] [%s] %s\n", timestamp, f.group, level, msg)), nil
}
