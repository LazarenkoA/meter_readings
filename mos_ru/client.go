package mos_ru

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

type MosruConnector struct {
	httpClient *http.Client
	CookieFile string
	Authorized bool
	ctx        context.Context
	login      string
	pass       string
}

type option func(m *MosruConnector)

const (
	baseURL = "https://login.mos.ru/sps/login"
	authURL = "https://my.mos.ru/login"
)

func NewMosruConnector(ctx context.Context, login, pass string, options ...option) *MosruConnector {
	newClient := &MosruConnector{
		httpClient: http.DefaultClient,
		ctx:        ctx,
		login:      login,
		pass:       pass,
	}

	for _, opt := range options {
		opt(newClient)
	}

	newClient.httpClient.Jar, _ = cookiejar.New(nil)
	return newClient
}

func (mc *MosruConnector) Auth() error {
	resp, err := mc.httpClient.Get(authURL)
	if err != nil {
		return errors.Wrap(err, "my.mos.ru/login error")
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Извлекаем login_url через regexp
	re := regexp.MustCompile(`login_url = "(https:\/\/login\.mos\.ru\/sps\/oauth\/ae[^"]+)"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return errors.New("login_url not found")
	}
	loginURL := matches[1]

	// Второй запрос на loginURL
	resp, err = mc.httpClient.Get(loginURL)
	if err != nil {
		return fmt.Errorf("%s req error: %w", loginURL, err)
	}
	resp.Body.Close()

	// POST на https://login.mos.ru/sps/login/methods/password
	form := url.Values{}
	form.Set("login", mc.login)
	form.Set("password", mc.pass)
	form.Set("isDelayed", "false")
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/methods/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = mc.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "auth request error")
	}
	resp.Body.Close()

	// Завершающий запрос
	resp, err = mc.httpClient.Get("https://my.mos.ru/my/#/")
	if err != nil {
		log.Println("Ошибка завершающего запроса:", err)
		return errors.Wrap(err, "finally request error")
	}
	resp.Body.Close()

	// Проверяем наличие Ltpatoken2 в cookie
	for _, cookie := range mc.httpClient.Jar.Cookies(&url.URL{Scheme: "https", Host: "my.mos.ru"}) {
		if cookie.Name == "Ltpatoken2" {
			mc.Authorized = true
			return nil
		}
	}

	return errors.New("couldn't log in")
}

func WithHttpClient(c *http.Client) option {
	return func(m *MosruConnector) {
		m.httpClient = c
	}
}
