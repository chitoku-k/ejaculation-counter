//go:generate mockgen -source=websocket.go -destination=websocket_mock.go -package=wrapper -self_package=github.com/chitoku-k/ejaculation-counter/supplier/infrastructure/wrapper

package wrapper

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
)

type Dialer interface {
	DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (Conn, *http.Response, error)
}

type dialer struct {
	d *websocket.Dialer
}

func NewDialer(d *websocket.Dialer) Dialer {
	return &dialer{d}
}

func (d *dialer) DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (Conn, *http.Response, error) {
	return d.d.DialContext(ctx, urlStr, requestHeader)
}

type Conn interface {
	Close() error
	ReadJSON(v interface{}) error
}
