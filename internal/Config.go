package internal

import (
	"io"
	"os"

	"github.com/gofiber/fiber/v2/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

type config struct {
	MaxCacheAge   int64  `yaml:"maxCacheAge"`
	IPMapFile     string `yaml:"ipMapFile"`
	PACRoot       string `yaml:"pacRoot"`
	ContactInfo   string `yaml:"contactInfo"`
	AccessLogFile string `yaml:"accessLogFile"`
	EventLogFile  string `yaml:"eventLogFile"`
	DoAutoRefresh bool   `yaml:"doAutoRefresh"`
	Port          uint16 `yaml:"port"`
}

var conf *config

func LoadConfig(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	newConf := &config{}
	err = yaml.Unmarshal(data, newConf)
	if err != nil {
		return err
	}

	conf = newConf
	return nil
}

func getConfig() *config {
	return conf
}

var accessLog *lumberjack.Logger

func InitEventLogger() {
	fileLogger := &lumberjack.Logger{
		Filename: getConfig().EventLogFile,
		// TODO: offload to cfg
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
	multiLog := io.MultiWriter(os.Stdout, fileLogger)
	log.SetOutput(multiLog)
	log.Info("Application starting")
}

func getAccessLogger() io.Writer {
	if accessLog == nil {
		accessLog = &lumberjack.Logger{
			Filename: getConfig().AccessLogFile,
			// TODO: offload to cfg
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   //days
			Compress:   true, // disabled by default
		}
	}
	return accessLog
}
