package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Environment struct {
	Mastodon Mastodon

	Queue Queue

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

func Get() (Environment, error) {
	var missing []string
	var env Environment
	var logLevel string

	for k, v := range map[string]*string{
		"MASTODON_SERVER_URL":   &env.Mastodon.ServerURL,
		"MASTODON_STREAM":       &env.Mastodon.Stream,
		"MASTODON_ACCESS_TOKEN": &env.Mastodon.AccessToken,
		"MQ_HOST":               &env.Queue.Host,
		"PORT":                  &env.Port,
	} {
		*v = os.Getenv(k)

		if *v == "" {
			missing = append(missing, k)
		}
	}

	for k, v := range map[string]*string{
		"TLS_CERT":         &env.TLSCert,
		"TLS_KEY":          &env.TLSKey,
		"MQ_SSL_CERT":      &env.Queue.SSLCert,
		"MQ_SSL_KEY":       &env.Queue.SSLKey,
		"MQ_SSL_ROOT_CERT": &env.Queue.SSLRootCert,
		"MQ_USERNAME":      &env.Queue.Username,
		"MQ_PASSWORD":      &env.Queue.Password,
		"LOG_LEVEL":        &logLevel,
	} {
		*v = os.Getenv(k)
	}

	if logLevel != "" {
		var err error
		env.LogLevel, err = parseLogLevel(logLevel)
		if err != nil {
			return env, fmt.Errorf("failed to parse log level: %w", err)
		}
	}

	if len(missing) > 0 {
		return env, fmt.Errorf("missing env(s): %v", strings.Join(missing, ", "))
	}

	return env, nil
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
