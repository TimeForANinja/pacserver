package internal

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	MaxCacheAge       int64  `yaml:"maxCacheAge"`
	IPMapFile         string `yaml:"ipMapFile"`
	PACRoot           string `yaml:"pacRoot"`
	ContactInfo       string `yaml:"contactInfo"`
	AccessLogFile     string `yaml:"accessLogFile"`
	EventLogFile      string `yaml:"eventLogFile"`
	PidFile           string `yaml:"pidFile"`
	DoAutoRefresh     bool   `yaml:"doAutoRefresh"`
	Port              uint16 `yaml:"port"`
	PrometheusEnabled bool   `yaml:"prometheusEnabled"`
	PrometheusPath    string `yaml:"prometheusPath"`
}

var conf *Config

func LoadConfig(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	newConf := &Config{}
	err = yaml.Unmarshal(data, newConf)
	if err != nil {
		return err
	}

	conf = newConf
	return nil
}

func GetConfig() *Config {
	return conf
}

var accessLog *lumberjack.Logger

func InitEventLogger() {
	fileLogger := &lumberjack.Logger{
		Filename: GetConfig().EventLogFile,
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

// WritePidFile writes the current process ID to the configured PID file
func WritePidFile() error {
	pidFile := GetConfig().PidFile
	if pidFile == "" {
		return nil // No PID file configured, so nothing to do
	}

	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
}

// ReadPidFile reads the process ID from the configured PID file
func ReadPidFile() (int, error) {
	pidFile := GetConfig().PidFile
	if pidFile == "" {
		return 0, fmt.Errorf("no PID file configured")
	}

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}

	return pid, nil
}

// RemovePidFile removes the PID file if it exists
func RemovePidFile() error {
	pidFile := GetConfig().PidFile
	if pidFile == "" {
		return nil // No PID file configured, so nothing to do
	}

	// Check if the file exists before trying to remove it
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to do
	}

	return os.Remove(pidFile)
}
