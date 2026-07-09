package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"github.com/timeforaninja/pacserver/internal/storage"
	"github.com/timeforaninja/pacserver/pkg/utils"
)

type Config struct {
	IPMapFile         string `mapstructure:"ipMapFile"`
	PACRoot           string `mapstructure:"pacRoot"`
	DefaultPACFile    string `mapstructure:"defaultPACFile"`
	WPADFile          string `mapstructure:"wpadFile"`
	ContactInfo       string `mapstructure:"contactInfo"`
	AccessLogFile     string `mapstructure:"accessLogFile"`
	EventLogFile      string `mapstructure:"eventLogFile"`
	Port              uint16 `mapstructure:"port"`
	AdminSecret       string `mapstructure:"adminSecret"`
	PrometheusEnabled bool   `mapstructure:"prometheusEnabled"`
	PrometheusPath    string `mapstructure:"prometheusPath"`
	IgnoreMinors      bool   `mapstructure:"ignoreMinors"`
	Loglevel          string `mapstructure:"loglevel"`
}

var confStorage *Config

// LoadConfig reads, validates, and stores the active application configuration.
func LoadConfig(filename string) error {
	// Load into a temporary config first so validation can reject bad input before it goes live.
	newConf, err := loadConfigWithViper(filename)
	if err != nil {
		return err
	}

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

func loadConfigWithViper(filename string) (*Config, error) {
	// Set the defaults here so the YAML file only needs to override the values it cares about.
	v := viper.New()
	v.SetConfigFile(filename)
	v.SetConfigType("yaml")
	v.SetDefault("ipMapFile", "data/zones.csv")
	v.SetDefault("pacRoot", "data/pacs")
	v.SetDefault("contactInfo", "Your Help Desk")
	v.SetDefault("accessLogFile", "access.log")
	v.SetDefault("eventLogFile", "event.log")
	v.SetDefault("port", uint16(8080))
	v.SetDefault("adminSecret", "")
	v.SetDefault("prometheusEnabled", false)
	v.SetDefault("prometheusPath", "/metrics")
	v.SetDefault("ignoreMinors", false)
	v.SetDefault("loglevel", "INFO")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	newConf := &Config{}
	if err := v.Unmarshal(newConf); err != nil {
		return nil, err
	}

	// Derive the special PAC file paths from the root unless the config already provided them.
	if newConf.DefaultPACFile == "" {
		newConf.DefaultPACFile = filepath.Join(newConf.PACRoot, "default.pac")
	}
	if newConf.WPADFile == "" {
		newConf.WPADFile = filepath.Join(newConf.PACRoot, "wpad.dat")
	}

	return newConf, nil
}

func validateConfig(conf *Config) error {
	if conf == nil {
		return fmt.Errorf("config must not be nil")
	}

	// Keep the contact info printable because it is injected into generated PAC content.
	contactRegex := regexp.MustCompile(`^[\w\s\-.,@() ]+$`)
	if !contactRegex.MatchString(conf.ContactInfo) {
		return fmt.Errorf("contact info contains invalid characters")
	}

	err := utils.ValidateLogLevel(conf.Loglevel)
	if err != nil {
		return err
	}

	if strings.TrimSpace(conf.AdminSecret) == "" {
		return fmt.Errorf("admin secret must be configured")
	}

	// Validate every file input up front so startup fails fast and predictably.
	zoneInfo, err := os.Stat(conf.IPMapFile)
	if err != nil || zoneInfo.IsDir() {
		return fmt.Errorf("Zone-File does not exist or is not a file: %s", conf.IPMapFile)
	}

	// PACRoot must be a directory because the storage package scans it for templates.
	pacRootInfo, err := os.Stat(conf.PACRoot)
	if err != nil || !pacRootInfo.IsDir() {
		return fmt.Errorf("PACRoot directory does not exist or is not a directory: %s", conf.PACRoot)
	}

	// The default PAC is loaded directly from disk, so it must be an actual file.
	fileInfo, err := os.Stat(conf.DefaultPACFile)
	if err != nil || fileInfo.IsDir() {
		return fmt.Errorf("DefaultPACFile does not exist or is not a file: %s", conf.DefaultPACFile)
	}

	// WPAD follows the same rule as the default PAC and is served directly.
	fileInfo, err = os.Stat(conf.WPADFile)
	if err != nil || fileInfo.IsDir() {
		return fmt.Errorf("WPADFile does not exist or is not a file: %s", conf.WPADFile)
	}

	return nil
}

// GetConfig returns the currently loaded configuration.
func GetConfig() *Config {
	return confStorage
}

// ToStorageConfig converts the app config into the subset required by storage.
func (conf *Config) ToStorageConfig() storage.StorageConfig {
	if conf == nil {
		return storage.StorageConfig{}
	}

	return storage.StorageConfig{
		IPMapFile:      conf.IPMapFile,
		PACRoot:        conf.PACRoot,
		DefaultPACFile: conf.DefaultPACFile,
		WPADFile:       conf.WPADFile,
		ContactInfo:    conf.ContactInfo,
		IgnoreMinors:   conf.IgnoreMinors,
	}
}
