package mos_ru

import (
	"compress/gzip"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Константы для авторизации.
const (
	OAUTH_URL = "https://login.mos.ru/sps/oauth/ae?" +
		"scope=openid+profile+blitz_user_rights+snils+contacts+blitz_change_password&" +
		"access_type=offline&" +
		"response_type=code&" +
		"redirect_uri=https://dnevnik.mos.ru/sudir" +
		"&client_id=dnevnik.mos.ru"
	EVENT         = "361e8e40a297ed04767fb8c93b686874942ae3159e4445afa9fd864b4b501ee1"
	CSRF_ENDPOINT = "https://login.mos.ru/7ccd851171c76f27e541d264c4186df3"
	FORM_ACTION   = "https://login.mos.ru/sps/login/methods/password"
)

var (
	ErrCredentialsInvalid = errors.New("credentials are invalid")
	ErrUnknown            = errors.New("unknown error")
	ErrIpBan              = errors.New("IP banned")
)

type RequestsAuthorization struct {
	login          string
	password       string
	client         *http.Client
	requestsParams map[string]string
}

func NewRequestsAuthorization(login, password string, params map[string]string) *RequestsAuthorization {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 20 {
				return fmt.Errorf("остановлено после %d редиректов", len(via))
			}
			return nil
		},
	}

	return &RequestsAuthorization{
		login:          login,
		password:       password,
		client:         client,
		requestsParams: params,
	}
}

// Функция для проведения авторизации, возвращает ответ сервера с токеном и профайлами
func (r *RequestsAuthorization) ProceedAuthorization() (string, error) {
	req, _ := http.NewRequest("GET", OAUTH_URL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.91 Safari/537.36")
	for k, v := range r.requestsParams {
		req.Header.Set(k, v)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 403 {
		return "", ErrIpBan
	}

	time.Sleep(time.Duration(rand.Intn(21)+10) * 100 * time.Millisecond)

	statsUrl := fmt.Sprintf("https://stats.mos.ru/handler/handler.js?time=%d", time.Now().Unix())
	r.client.Get(statsUrl)
	time.Sleep(time.Duration(rand.Intn(21)+10) * 100 * time.Millisecond)

	body, _ := io.ReadAll(resp.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}

	proofOfWork, exists := doc.Find("input[name=proofOfWork]").Attr("value")
	if !exists {
		return "", ErrUnknown
	}

	pow, err := buildPow(proofOfWork)
	if err != nil {
		return "", errors.Wrap(err, "pow build error")
	}

	form := url.Values{}
	form.Add("login", r.login)
	form.Add("password", r.password)
	form.Add("proofOfWork", pow)
	form.Add("isDelayed", "false")
	form.Add("bfp", "d04b73b5005e02b49eb8cee52841e864")

	loginReq, _ := http.NewRequest(http.MethodPost, FORM_ACTION, strings.NewReader(form.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 YaBrowser/25.6.0.0 Safari/537.36")
	loginReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	loginReq.Header.Set("Accept-Language", "ru,en;q=0.9")
	loginReq.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Header.Set("Origin", "https://login.mos.ru")
	loginReq.Header.Set("Referer", "https://login.mos.ru/sps/login/methods/password?bo=%2Fsps%2Foauth%2Fae%3Fclient_id%3Dmy.mos.ru%26response_type%3Dcode%26redirect_uri%3Dhttps%3A%2F%2Fmy.mos.ru%2Foauth%26scope%3Dopenid%2Bprofile%2Bblitz_api_user%2Bblitz_api_user_chg%2Bblitz_api_usec%2Bblitz_api_usec_chg%2Bblitz_api_uapps%2Bblitz_api_uapps_chg%2Bblitz_api_uaud%2Bblitz_user_rights%2Bblitz_change_password")
	loginReq.Header.Set("Sec-Fetch-Site", "same-origin")
	loginReq.Header.Set("Sec-Fetch-Mode", "navigate")
	loginReq.Header.Set("Sec-Fetch-Dest", "document")
	loginReq.Header.Set("Sec-CH-UA", `"Chromium";v="136", "YaBrowser";v="25.6", "Not.A/Brand";v="99", "Yowser";v="2.5"`)
	loginReq.Header.Set("Sec-CH-UA-Mobile", "?0")
	loginReq.Header.Set("Sec-CH-UA-Platform", `"Windows"`)
	loginReq.Header.Set("Sec-CH-UA-Platform-Version", `"19.0.0"`)
	loginReq.Header.Set("Upgrade-Insecure-Requests", "1")
	loginReq.Header.Set("DNT", "1")

	resp2, err := r.client.Do(loginReq)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()

	if resp2.StatusCode >= 301 && resp2.StatusCode < 400 {
		redirectURL := resp2.Header.Get("Location")
		u, _ := url.Parse(redirectURL)
		code := u.Query().Get("code")
		if code == "" {
			return "", ErrUnknown
		}
		apiURL := fmt.Sprintf("https://dnevnik.mos.ru/lms/api/sudir/oauth/te?code=%s", code)
		apiReq, _ := http.NewRequest("GET", apiURL, nil)
		apiReq.Header.Set("Accept", "application/vnd.api.v3+json")
		apiResp, err := r.client.Do(apiReq)
		if err != nil {
			return "", err
		}
		defer apiResp.Body.Close()
		result, _ := io.ReadAll(apiResp.Body)
		return string(result), nil
	} else if resp2.StatusCode == 200 {
		//body, _ := io.ReadAll(resp2.Body)

		var reader io.ReadCloser
		if resp2.Header.Get("Content-Encoding") == "gzip" {
			reader, _ = gzip.NewReader(resp2.Body)
			defer reader.Close()
		} else {
			reader = resp2.Body
		}

		body, _ := io.ReadAll(reader)

		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		block := doc.Find("blockquote.blockquote-danger").Text()
		if trimBlock := strings.TrimSpace(block); trimBlock != "" {
			return "", fmt.Errorf("%w: %s", ErrCredentialsInvalid, trimBlock)
		}
		return "", ErrUnknown
	} else if resp2.StatusCode == 403 {
		return "", ErrIpBan
	} else {
		return "", ErrUnknown
	}
}

func (r *RequestsAuthorization) generateCSRFToken(csrftokenw string) (string, error) {
	req, _ := http.NewRequest("GET", CSRF_ENDPOINT, nil)
	req.Header.Set("x-csrftokenw", csrftokenw)
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("x-ajax-token", EVENT)
	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", ErrUnknown
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrf-token-value" {
			return cookie.Value, nil
		}
	}
	return "", ErrUnknown
}
