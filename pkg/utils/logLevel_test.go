package utils

import (
	"github.com/gofiber/fiber/v2/log"
	"testing"
)

func TestGetLoglevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level string
		want  log.Level
	}{
		{
			name:  "DEBUG level",
			level: "DEBUG",
			want:  log.LevelDebug,
		},
		{
			name:  "INFO level",
			level: "INFO",
			want:  log.LevelInfo,
		},
		{
			name:  "WARN level",
			level: "WARN",
			want:  log.LevelWarn,
		},
		{
			name:  "ERROR level",
			level: "ERROR",
			want:  log.LevelError,
		},
		{
			name:  "Lowercase debug",
			level: "debug",
			want:  log.LevelDebug,
		},
		{
			name:  "Mixed case Info",
			level: "Info",
			want:  log.LevelInfo,
		},
		{
			name:  "Unknown level defaults to INFO",
			level: "UNKNOWN",
			want:  log.LevelInfo,
		},
		{
			name:  "Empty string defaults to INFO",
			level: "",
			want:  log.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetLoglevel(tt.level)
			if got != tt.want {
				t.Errorf("GetLoglevel(%q) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestValidateLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "Valid DEBUG level",
			level:   "DEBUG",
			wantErr: false,
		},
		{
			name:    "Valid INFO level",
			level:   "INFO",
			wantErr: false,
		},
		{
			name:    "Valid WARN level",
			level:   "WARN",
			wantErr: false,
		},
		{
			name:    "Valid ERROR level",
			level:   "ERROR",
			wantErr: false,
		},
		{
			name:    "Valid lowercase debug",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "Valid mixed case Info",
			level:   "Info",
			wantErr: false,
		},
		{
			name:    "Invalid level",
			level:   "UNKNOWN",
			wantErr: true,
		},
		{
			name:    "Empty string is invalid",
			level:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLogLevel(%q) error = %v, wantErr %v", tt.level, err, tt.wantErr)
			}
		})
	}
}