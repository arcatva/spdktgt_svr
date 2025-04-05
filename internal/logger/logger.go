package logger

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func Init() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&spdkFormatter{})
	logrus.SetOutput(os.Stdout)
}

type spdkFormatter struct{}

func (f *spdkFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000000")
	message := entry.Message

	var callerInfo string
	if entry.Caller != nil {
		callerInfo = fmt.Sprintf("%s:%d (%s)",
			entry.Caller.File, entry.Caller.Line, entry.Caller.Function)
	} else {
		callerInfo = "unknown"
	}

	logLine := fmt.Sprintf("[%s] [%s] [%s] %s\n",
		timestamp, entry.Level.String(), callerInfo, message)

	return []byte(logLine), nil
}
