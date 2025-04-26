package internal

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

type YAMLConfig struct {
	// YAML unfortunately doesn't support default values
	// we can, however, use pointers to identify if a value is not set
	IPMapFile         *string `yaml:"ipMapFile"`
	PACRoot           *string `yaml:"pacRoot"`
	ContactInfo       *string `yaml:"contactInfo"`
	AccessLogFile     *string `yaml:"accessLogFile"`
	EventLogFile      *string `yaml:"eventLogFile"`
	MaxCacheAge       *int64  `yaml:"maxCacheAge"`
	PidFile           *string `yaml:"pidFile"`
	Port              *uint16 `yaml:"port"`
	PrometheusEnabled *bool   `yaml:"prometheusEnabled"`
	PrometheusPath    *string `yaml:"prometheusPath"`
	IgnoreMinors      *bool   `yaml:"ignoreMinors"`
	Loglevel          *string `yaml:"loglevel"`
}

type Config struct {
	IPMapFile         string
	PACRoot           string
	ContactInfo       string
	AccessLogFile     string
	EventLogFile      string
	MaxCacheAge       int64
	PidFile           string
	Port              uint16
	PrometheusEnabled bool
	PrometheusPath    string
	IgnoreMinors      bool
	Loglevel          string
}

var conf *Config

func LoadConfig(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	yamlConf := &YAMLConfig{}
	err = yaml.Unmarshal(data, yamlConf)
	if err != nil {
		return err
	}

	newConf := overloadDefaults(yamlConf)

	err = validateConfig(newConf)
	if err != nil {
		return err
	}

	// assign the new config to the global config
	// in case we ever start using hot-reload of the config
	// this will ensure we don't load a broken one
	conf = newConf
	return nil
}

func ifIsNil[T comparable](val *T, def T) T {
	if val == nil {
		return def
	}
	return *val
}

func overloadDefaults(conf *YAMLConfig) *Config {
	newConf := &Config{}
	newConf.IPMapFile = ifIsNil(conf.IPMapFile, "data/zones.csv")
	newConf.PACRoot = ifIsNil(conf.PACRoot, "data/pacs")
	newConf.ContactInfo = ifIsNil(conf.ContactInfo, "Your Help Desk")
	newConf.AccessLogFile = ifIsNil(conf.AccessLogFile, "access.log")
	newConf.EventLogFile = ifIsNil(conf.EventLogFile, "event.log")
	newConf.MaxCacheAge = ifIsNil(conf.MaxCacheAge, int64(900))
	newConf.PidFile = ifIsNil(conf.PidFile, "pacserver.pid")
	newConf.Port = ifIsNil(conf.Port, uint16(8080))
	newConf.PrometheusEnabled = ifIsNil(conf.PrometheusEnabled, false)
	newConf.PrometheusPath = ifIsNil(conf.PrometheusPath, "/metrics")
	newConf.IgnoreMinors = ifIsNil(conf.IgnoreMinors, false)
	newConf.Loglevel = ifIsNil(conf.Loglevel, "INFO")
	return newConf
}

func validateConfig(conf *Config) error {
	// Validate contact info - allow only alphanumeric, spaces, and basic punctuation
	contactRegex := regexp.MustCompile(`^[\w\s\-.,@() ]+$`)
	if !contactRegex.MatchString(conf.ContactInfo) {
		return fmt.Errorf("contact info contains invalid characters")
	}

	knownLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	fmt.Println(knownLevels, conf.Loglevel, strings.ToUpper(conf.Loglevel))
	if !slices.Contains(knownLevels, strings.ToUpper(conf.Loglevel)) {
		return fmt.Errorf("loglevel must be one of %v", knownLevels)
	}
	return nil
}

func GetConfig() *Config {
	return conf
}

var accessLog *lumberjack.Logger

func (conf *Config) getLoglevel() log.Level {
	switch strings.ToUpper(conf.Loglevel) {
	case "DEBUG":
		return log.LevelDebug
	case "INFO":
		return log.LevelInfo
	case "WARN":
		return log.LevelWarn
	case "ERROR":
		return log.LevelError
	}
	return log.LevelInfo
}

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
	log.SetLevel(conf.getLoglevel())
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
