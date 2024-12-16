package client

import (
	"net/http"
	"net/http/cookiejar"
)

func NewHttpClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	transport := http.DefaultTransport.(*http.Transport)
	transport.ForceAttemptHTTP2 = false
	transport.MaxIdleConnsPerHost = 16

	return &http.Client{
		Jar:       jar,
		Transport: transport,
	}, nil
}
