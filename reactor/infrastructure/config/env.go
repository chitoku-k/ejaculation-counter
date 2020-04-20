package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Environment struct {
	DB DB

	Mastodon Mastodon

	Queue Queue

	Port   string
	UserID int64
}

type DB struct {
	Host     string
	Database string
	Username string
	Password string
}

type Mastodon struct {
	ServerURL   string
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

	for k, v := range map[string]*string{
		"DB_HOST":               &env.DB.Host,
		"DB_DATABASE":           &env.DB.Database,
		"DB_USERNAME":           &env.DB.Username,
		"DB_PASSWORD":           &env.DB.Password,
		"MASTODON_SERVER_URL":   &env.Mastodon.ServerURL,
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

	if len(missing) > 0 {
		return env, errors.New("missing env(s): " + strings.Join(missing, ", "))
	}

	return env, nil
}
