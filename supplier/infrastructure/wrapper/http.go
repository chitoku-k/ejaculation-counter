//go:generate mockgen -source=http.go -destination=http_mock.go -package=wrapper -self_package=github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper

package wrapper

import (
	"io"
	"net/http"
)

type HttpClient interface {
	Get(url string) (*http.Response, error)
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}
