package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

func InitLogger() {
	log.SetFlags(0)
	log.SetOutput(newLogWriter())
}

type logWriter struct{}

func (writer logWriter) Write(bytes []byte) (int, error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000000")
	return fmt.Fprintf(os.Stdout, "[%s] %s", timestamp, string(bytes))
}

func newLogWriter() *logWriter {
	return &logWriter{}
}
