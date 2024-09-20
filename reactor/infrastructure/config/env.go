package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Environment struct {
	DB       DB
	Mastodon Mastodon
	Queue    Queue
	External External

	LogLevel slog.Level
	Port     string
	TLSCert  string
	TLSKey   string
	UserID   int64
}

type DB struct {
	Host        string
	Database    string
	Username    string
	Password    string
	SSLMode     string
	SSLCert     string
	SSLKey      string
	SSLRootCert string
	MaxLifetime time.Duration
}

type Mastodon struct {
	UserID      string
	ServerURL   string
	AccessToken string
}

type Queue struct {
	Host        string
	Username    string
	Password    string
	SSLCert     string
	SSLKey      string
	SSLRootCert string
}

type External struct {
	MpywAPIURL string
}

func Get() (env Environment, errs error) {
	for _, entry := range []struct {
		name     string
		field    any
		optional bool
	}{
		{name: "DB_HOST", field: &env.DB.Host},
		{name: "DB_DATABASE", field: &env.DB.Database},
		{name: "DB_USERNAME", field: &env.DB.Username},
		{name: "DB_PASSWORD", field: &env.DB.Password, optional: true},
		{name: "DB_SSL_MODE", field: &env.DB.SSLMode},
		{name: "DB_SSL_CERT", field: &env.DB.SSLCert, optional: true},
		{name: "DB_SSL_KEY", field: &env.DB.SSLKey, optional: true},
		{name: "DB_SSL_ROOT_CERT", field: &env.DB.SSLRootCert, optional: true},
		{name: "DB_MAX_LIFETIME_SEC", field: &env.DB.MaxLifetime, optional: true},
		{name: "MASTODON_USER_ID", field: &env.Mastodon.UserID},
		{name: "MASTODON_SERVER_URL", field: &env.Mastodon.ServerURL},
		{name: "MASTODON_ACCESS_TOKEN", field: &env.Mastodon.AccessToken},
		{name: "MQ_HOST", field: &env.Queue.Host},
		{name: "MQ_USERNAME", field: &env.Queue.Username, optional: true},
		{name: "MQ_PASSWORD", field: &env.Queue.Password, optional: true},
		{name: "MQ_SSL_CERT", field: &env.Queue.SSLCert, optional: true},
		{name: "MQ_SSL_KEY", field: &env.Queue.SSLKey, optional: true},
		{name: "MQ_SSL_ROOT_CERT", field: &env.Queue.SSLRootCert, optional: true},
		{name: "EXT_MPYW_API_URL", field: &env.External.MpywAPIURL},
		{name: "LOG_LEVEL", field: &env.External, optional: true},
		{name: "PORT", field: &env.Port},
		{name: "TLS_CERT", field: &env.TLSCert, optional: true},
		{name: "TLS_KEY", field: &env.TLSKey, optional: true},
		{name: "USER_ID", field: &env.UserID},
	} {
		v := os.Getenv(entry.name)
		if v == "" {
			if !entry.optional {
				errs = errors.Join(errs, fmt.Errorf("missing: %v", entry.name))
			}
			continue
		}

		switch field := entry.field.(type) {
		case *string:
			*field = v

		case *int64:
			v, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("%s is invalid: %w", entry.name, err))
				continue
			}
			*field = v

		case *time.Duration:
			v, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("%s is invalid: %w", entry.name, err))
				continue
			}
			*field = time.Duration(v) * time.Second

		case *slog.Level:
			v, err := parseLogLevel(v)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("%s is invalid: %w", entry.name, err))
				continue
			}
			*field = v
		}
	}

	return
}

func parseLogLevel(lvl string) (slog.Level, error) {
	switch strings.ToLower(lvl) {
	case "error":
		return slog.LevelError, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	}

	var l slog.Level
	return l, fmt.Errorf("not a valid log level: %q", lvl)
}
