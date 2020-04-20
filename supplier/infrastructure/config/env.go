package config

import (
	"errors"
	"os"
	"strings"
)

type Environment struct {
	Mastodon Mastodon

	Reactor Reactor

	Queue Queue

	Port string
}

type Mastodon struct {
	UserID      string
	ServerURL   string
	AccessToken string
}

type Queue struct {
	Host     string
	Username string
	Password string
}

type Reactor struct {
	Host string
}

func Get() (Environment, error) {
	var missing []string
	var env Environment

	for k, v := range map[string]*string{
		"MASTODON_USER_ID":      &env.Mastodon.UserID,
		"MASTODON_SERVER_URL":   &env.Mastodon.ServerURL,
		"MASTODON_ACCESS_TOKEN": &env.Mastodon.AccessToken,
		"MQ_HOST":               &env.Queue.Host,
		"MQ_USERNAME":           &env.Queue.Username,
		"MQ_PASSWORD":           &env.Queue.Password,
		"REACTOR_HOST":          &env.Reactor.Host,
		"PORT":                  &env.Port,
	} {
		*v = os.Getenv(k)

		if *v == "" {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		return env, errors.New("missing env(s): " + strings.Join(missing, ", "))
	}

	return env, nil
}
