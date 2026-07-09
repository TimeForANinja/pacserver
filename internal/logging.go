package internal

import (
	"io"
	"os"
	"runtime/debug"

	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/pkg/utils"
	"gopkg.in/natefinch/lumberjack.v2"
)

var accessLog *lumberjack.Logger

func (conf *Config) getLoglevel() log.Level {
	if conf == nil {
		return log.LevelInfo
	}
	// Translate the configured string once so the logger setup stays simple.
	return utils.GetLoglevel(conf.Loglevel)
}

// InitEventLogger routes application logs to stdout and the rotating event log file.
func InitEventLogger() {
	conf := GetConfig()
	if conf == nil {
		log.SetLevel(log.LevelInfo)
		log.SetOutput(os.Stdout)
		log.Warn("Event logger initialized without loaded config; falling back to stdout only")
		return
	}

	// Mirror logs to stdout and a rotating file so reload/test commands still stay visible.
	fileLogger := &lumberjack.Logger{
		Filename: conf.EventLogFile,
		// TODO: offload to cfg
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
	multiLog := io.MultiWriter(os.Stdout, fileLogger)
	log.SetLevel(conf.getLoglevel())
	log.SetOutput(multiLog)
	log.Info("Application starting")
}

// LogUnexpectedError writes a contextual error and stack trace to the event log.
func LogUnexpectedError(context string, err error) {
	if err == nil {
		return
	}

	log.Errorf("%s: %v\n%s", context, err, debug.Stack())
}

func getAccessLogger() io.Writer {
	conf := GetConfig()
	if conf == nil {
		return os.Stdout
	}

	// Reuse one rotating writer for the access log so repeated requests do not reopen files.
	if accessLog == nil {
		accessLog = &lumberjack.Logger{
			Filename: conf.AccessLogFile,
			// TODO: offload to cfg
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   //days
			Compress:   true, // disabled by default
		}
	}
	return accessLog
}
