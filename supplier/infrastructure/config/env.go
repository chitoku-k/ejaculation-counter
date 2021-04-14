package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type Environment struct {
	Mastodon Mastodon

	Queue Queue

	LogLevel logrus.Level
	Port     string
}

type Mastodon struct {
	ServerURL   string
	Stream      string
	AccessToken string
}

type Queue struct {
	Host     string
	Username string
	Password string
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
		"MQ_USERNAME":           &env.Queue.Username,
		"MQ_PASSWORD":           &env.Queue.Password,
		"PORT":                  &env.Port,
	} {
		*v = os.Getenv(k)

		if *v == "" {
			missing = append(missing, k)
		}
	}

	for k, v := range map[string]*string{
		"LOG_LEVEL": &logLevel,
	} {
		*v = os.Getenv(k)
	}

	env.LogLevel = logrus.InfoLevel
	if logLevel != "" {
		var err error
		env.LogLevel, err = logrus.ParseLevel(logLevel)
		if err != nil {
			return env, fmt.Errorf("failed to parse log level: %w", err)
		}
	}

	if len(missing) > 0 {
		return env, fmt.Errorf("missing env(s): %v", strings.Join(missing, ", "))
	}

	return env, nil
}
