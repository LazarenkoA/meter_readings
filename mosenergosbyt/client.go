package mosenergosbyt

import (
	"context"
	"net/http"
)

type Mosenergosbyt struct {
	ctx        context.Context
	session    string
	httpClient *http.Client
	login      string
	pass       string
}

type option func(m *Mosenergosbyt)

const (
	baseURL    = "https://my.mosenergosbyt.ru/gate_lkcomu"
	deviceInfo = `{"appver":"1.17.1","type":"browser","userAgent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36"}`
)

func NewClient(ctx context.Context, login, pass string, options ...option) *Mosenergosbyt {
	newClient := &Mosenergosbyt{
		ctx:        ctx,
		httpClient: http.DefaultClient,
		login:      login,
		pass:       pass,
	}

	for _, opt := range options {
		opt(newClient)
	}

	return newClient
}

func WithHttpClient(c *http.Client) option {
	return func(m *Mosenergosbyt) {
		m.httpClient = c
	}
}
