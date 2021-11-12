package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Environment struct {
	DB DB

	Mastodon Mastodon

	Queue Queue

	External External

	LogLevel logrus.Level
	Port     string
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

func Get() (Environment, error) {
	var missing []string
	var env Environment
	var logLevel string

	for k, v := range map[string]*string{
		"DB_HOST":               &env.DB.Host,
		"DB_DATABASE":           &env.DB.Database,
		"DB_USERNAME":           &env.DB.Username,
		"DB_SSL_MODE":           &env.DB.SSLMode,
		"MASTODON_USER_ID":      &env.Mastodon.UserID,
		"MASTODON_SERVER_URL":   &env.Mastodon.ServerURL,
		"MASTODON_ACCESS_TOKEN": &env.Mastodon.AccessToken,
		"MQ_HOST":               &env.Queue.Host,
		"MQ_USERNAME":           &env.Queue.Username,
		"MQ_PASSWORD":           &env.Queue.Password,
		"EXT_MPYW_API_URL":      &env.External.MpywAPIURL,
		"PORT":                  &env.Port,
	} {
		*v = os.Getenv(k)

		if *v == "" {
			missing = append(missing, k)
		}
	}

	for k, v := range map[string]*string{
		"DB_PASSWORD":      &env.DB.Password,
		"DB_SSL_CERT":      &env.DB.SSLCert,
		"DB_SSL_KEY":       &env.DB.SSLKey,
		"DB_SSL_ROOT_CERT": &env.DB.SSLRootCert,
		"MQ_SSL_CERT":      &env.Queue.SSLCert,
		"MQ_SSL_KEY":       &env.Queue.SSLKey,
		"MQ_SSL_ROOT_CERT": &env.Queue.SSLRootCert,
		"LOG_LEVEL":        &logLevel,
	} {
		*v = os.Getenv(k)
	}

	for k, v := range map[string]*time.Duration{
		"DB_MAX_LIFETIME_SEC": &env.DB.MaxLifetime,
	} {
		s := os.Getenv(k)
		if s == "" {
			continue
		}

		t, err := strconv.Atoi(s)
		if err != nil {
			return env, fmt.Errorf("%s is invalid: %w", k, err)
		}

		*v = time.Duration(t) * time.Second
	}

	for k, v := range map[string]*int64{
		"USER_ID": &env.UserID,
	} {
		s := os.Getenv(k)
		if s == "" {
			missing = append(missing, k)
			continue
		}

		fmt.Sscanf(s, "%d", v)
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
		return env, errors.New("missing env(s): " + strings.Join(missing, ", "))
	}

	return env, nil
}
