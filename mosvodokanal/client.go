package mosvodokanal

import (
	"context"
	"net/http"
	"net/http/cookiejar"
)

type Mosvodokanal struct {
	ctx        context.Context
	httpClient *http.Client
	login      string
	pass       string
}

type option func(m *Mosvodokanal)

const (
	baseURL    = "https://onewind.mosvodokanal.ru/api"
	deviceInfo = `{"appver":"1.17.1","type":"browser","userAgent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36"}`
)

func NewClient(ctx context.Context, login, pass string, options ...option) *Mosvodokanal {
	newClient := &Mosvodokanal{
		ctx:        ctx,
		httpClient: http.DefaultClient,
		login:      login,
		pass:       pass,
	}

	for _, opt := range options {
		opt(newClient)
	}

	newClient.httpClient.Jar, _ = cookiejar.New(nil)
	return newClient
}

func WithHttpClient(c *http.Client) option {
	return func(m *Mosvodokanal) {
		m.httpClient = c
	}
}
