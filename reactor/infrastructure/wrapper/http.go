//go:generate mockgen -source=http.go -destination=http_mock.go -package=wrapper -self_package=github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/wrapper

package wrapper

import (
	"net/http"
	"net/url"
)

type HttpClient interface {
	Get(url string) (*http.Response, error)
	PostForm(url string, data url.Values) (*http.Response, error)
}
