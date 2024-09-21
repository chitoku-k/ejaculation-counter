package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Environment struct {
	Mastodon Mastodon
	Queue    Queue

	LogLevel slog.Level
	Port     string
	TLSCert  string
	TLSKey   string
}

type Mastodon struct {
	ServerURL   string
	Stream      string
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

func Get() (env Environment, errs error) {
	for _, entry := range []struct {
		name     string
		field    any
		optional bool
	}{
		{name: "MASTODON_SERVER_URL", field: &env.Mastodon.ServerURL},
		{name: "MASTODON_STREAM", field: &env.Mastodon.Stream},
		{name: "MASTODON_ACCESS_TOKEN", field: &env.Mastodon.AccessToken},
		{name: "MQ_HOST", field: &env.Queue.Host},
		{name: "MQ_USERNAME", field: &env.Queue.Username, optional: true},
		{name: "MQ_PASSWORD", field: &env.Queue.Password, optional: true},
		{name: "MQ_SSL_CERT", field: &env.Queue.SSLCert, optional: true},
		{name: "MQ_SSL_KEY", field: &env.Queue.SSLKey, optional: true},
		{name: "MQ_SSL_ROOT_CERT", field: &env.Queue.SSLRootCert, optional: true},
		{name: "LOG_LEVEL", field: &env.LogLevel, optional: true},
		{name: "PORT", field: &env.Port},
		{name: "TLS_CERT", field: &env.TLSCert, optional: true},
		{name: "TLS_KEY", field: &env.TLSKey, optional: true},
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
