package internal

import (
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/pkg/utils"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"regexp"
	"strconv"
)

type YAMLConfig struct {
	// YAML unfortunately doesn't support default values
	// we can, however, use pointers to identify if a value is not set
	IPMapFile         *string `yaml:"ipMapFile"`
	PACRoot           *string `yaml:"pacRoot"`
	DefaultPACFile    *string `yaml:"defaultPACFile"`
	WPADFile          *string `yaml:"wpadFile"`
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
	DefaultPACFile    string
	WPADFile          string
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

var confStorage *Config

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
	confStorage = newConf
	return nil
}

func overloadDefaults(conf *YAMLConfig) *Config {
	newConf := &Config{}
	// Set defaults and map to Config
	newConf.IPMapFile = utils.IfIsNil(conf.IPMapFile, "data/zones.csv")
	newConf.PACRoot = utils.IfIsNil(conf.PACRoot, "data/pacs")
	newConf.DefaultPACFile = utils.IfIsNil(conf.DefaultPACFile, newConf.PACRoot+"\\default.pac")
	newConf.WPADFile = utils.IfIsNil(conf.WPADFile, newConf.PACRoot+"\\wpad.dat")
	newConf.ContactInfo = utils.IfIsNil(conf.ContactInfo, "Your Help Desk")
	newConf.AccessLogFile = utils.IfIsNil(conf.AccessLogFile, "access.log")
	newConf.EventLogFile = utils.IfIsNil(conf.EventLogFile, "event.log")
	newConf.MaxCacheAge = utils.IfIsNil(conf.MaxCacheAge, int64(900))
	newConf.PidFile = utils.IfIsNil(conf.PidFile, "pacserver.pid")
	newConf.Port = utils.IfIsNil(conf.Port, uint16(8080))
	newConf.PrometheusEnabled = utils.IfIsNil(conf.PrometheusEnabled, false)
	newConf.PrometheusPath = utils.IfIsNil(conf.PrometheusPath, "/metrics")
	newConf.IgnoreMinors = utils.IfIsNil(conf.IgnoreMinors, false)
	newConf.Loglevel = utils.IfIsNil(conf.Loglevel, "INFO")
	return newConf
}

func validateConfig(conf *Config) error {
	// Validate contact info - allow only alphanumeric, spaces, and basic punctuation
	contactRegex := regexp.MustCompile(`^[\w\s\-.,@() ]+$`)
	if !contactRegex.MatchString(conf.ContactInfo) {
		return fmt.Errorf("contact info contains invalid characters")
	}

	err := utils.ValidateLogLevel(conf.Loglevel)
	if err != nil {
		return err
	}

	// Validate the Zone-File exists
	zoneInfo, err := os.Stat(conf.IPMapFile)
	if err != nil || zoneInfo.IsDir() {
		return fmt.Errorf("Zone-File does not exist or is not a file: %s", conf.IPMapFile)
	}

	// Validate that PACRoot exists and is a directory
	pacRootInfo, err := os.Stat(conf.PACRoot)
	if err != nil || !pacRootInfo.IsDir() {
		return fmt.Errorf("PACRoot directory does not exist or is not a directory: %s", conf.PACRoot)
	}

	// Check if DefaultPACFile exists, and if it does, ensure it's a file
	fileInfo, err := os.Stat(conf.DefaultPACFile)
	if err != nil || fileInfo.IsDir() {
		return fmt.Errorf("DefaultPACFile does not exist or is not a file: %s", conf.DefaultPACFile)
	}

	// Check if WPADFile exists, and if it does, ensure it's a file
	fileInfo, err = os.Stat(conf.WPADFile)
	if err != nil || fileInfo.IsDir() {
		return fmt.Errorf("WPADFile does not exist or is not a file: %s", conf.WPADFile)
	}

	return nil
}

func GetConfig() *Config {
	return confStorage
}

var accessLog *lumberjack.Logger

func (conf *Config) getLoglevel() log.Level {
	return utils.GetLoglevel(conf.Loglevel)
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
	log.SetLevel(confStorage.getLoglevel())
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
