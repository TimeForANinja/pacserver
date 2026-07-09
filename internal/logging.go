package internal

import (
	"io"
	"os"

	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/pkg/utils"
	"gopkg.in/natefinch/lumberjack.v2"
)

var accessLog *lumberjack.Logger

func (conf *Config) getLoglevel() log.Level {
	// Translate the configured string once so the logger setup stays simple.
	return utils.GetLoglevel(conf.Loglevel)
}

func InitEventLogger() {
	// Mirror logs to stdout and a rotating file so reload/test commands still stay visible.
	fileLogger := &lumberjack.Logger{
		Filename: GetConfig().EventLogFile,
		// TODO: offload to cfg
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
	multiLog := io.MultiWriter(os.Stdout, fileLogger)
	log.SetLevel(confStorage.getLoglevel())
	log.SetOutput(multiLog)
	log.Info("Application starting")
}

func getAccessLogger() io.Writer {
	// Reuse one rotating writer for the access log so repeated requests do not reopen files.
	if accessLog == nil {
		accessLog = &lumberjack.Logger{
			Filename: GetConfig().AccessLogFile,
			// TODO: offload to cfg
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   //days
			Compress:   true, // disabled by default
		}
	}
	return accessLog
}
