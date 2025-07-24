package salutespeech

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	OAuthUrl     = "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
	UploadUrl    = "https://smartspeech.sber.ru/rest/v1/data:upload"
	RecognizeUrl = "https://smartspeech.sber.ru/rest/v1/speech:async_recognize"
	StatusUrl    = "https://smartspeech.sber.ru/rest/v1/task:get"
	DownloadUrl  = "https://smartspeech.sber.ru/rest/v1/data:download"
)

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

type option func(m *Client)

type Client struct {
	ctx        context.Context
	httpClient *http.Client
	authKey    string
	token      string
	expire     time.Time
}

func NewClient(ctx context.Context, authKey string, options ...option) *Client {
	newClient := &Client{
		httpClient: http.DefaultClient,
		authKey:    authKey,
		ctx:        ctx,
	}

	for _, opt := range options {
		opt(newClient)
	}

	return newClient
}

func (c *Client) GetToken() error {
	if !c.expire.IsZero() && c.expire.After(time.Now()) {
		return nil
	}

	req, err := http.NewRequestWithContext(c.ctx, http.MethodPost, OAuthUrl, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+c.authKey)
	req.Header.Set("RqUID", uuid.New().String())

	data, err := c.request(req)
	if err != nil {
		return errors.Wrap(err, "GetToken request error")
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return errors.Wrap(err, "failed to unmarshal response")
	}

	c.token = token.AccessToken
	c.expire = time.Now().Add(time.Duration(token.ExpiresAt-60000) * time.Millisecond)

	return nil
}

//
//func (c *Client) GetStatus(taskId string) (string, error) {
//	req := fasthttp.AcquireRequest()
//	req.SetRequestURI(StatusUrl + "?id=" + taskId)
//	req.Header.SetMethod(fasthttp.MethodGet)
//	req.Header.Add("Accept", "application/json")
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("Authorization", "Bearer "+c.token)
//
//	defer fasthttp.ReleaseRequest(req)
//
//	resp := fasthttp.AcquireResponse()
//	defer fasthttp.ReleaseResponse(resp)
//
//	if err := c.httpClient.DoTimeout(req, resp, time.Duration(10)*time.Second); err != nil {
//		return "", fmt.Errorf("timeout, error: %v", err)
//	}
//
//	if resp.StatusCode() != fasthttp.StatusOK {
//		return "", fmt.Errorf("wrong status code: %v\n%v", resp.StatusCode(), resp.Body())
//	}
//
//	var res StatusResponse
//
//	if err := json.Unmarshal(resp.Body(), &res); err != nil {
//		return "", fmt.Errorf("failed to unmarshal response: %w", err)
//	}
//
//	if res.Status != 200 {
//		return "", fmt.Errorf("wrong status code: %v", res.Status)
//	}
//
//	if res.Result.Status == "ERROR" {
//		return "ERROR", fmt.Errorf("error: %v", res.Result.Error)
//	}
//
//	return res.Result.ResponseFileId, nil
//}
//
//func (c *Client) UploadFile(filename string) (string, error) {
//	file, err := os.Open(filename)
//
//	if err != nil {
//		return "", fmt.Errorf("failed to open file: %w", err)
//	}
//
//	defer file.Close()
//
//	fileInfo, err := file.Stat()
//
//	if err != nil {
//		return "", fmt.Errorf("failed to get file info: %w", err)
//	}
//
//	req := fasthttp.AcquireRequest()
//	req.SetRequestURI(UploadUrl)
//	req.Header.SetMethod(fasthttp.MethodPost)
//	req.Header.Add("Accept", "application/json")
//	req.Header.Set("Content-Type", "binary/octet-stream")
//	req.Header.Set("Authorization", "Bearer "+c.token)
//	req.SetBodyStream(file, int(fileInfo.Size()))
//
//	defer fasthttp.ReleaseRequest(req)
//
//	resp := fasthttp.AcquireResponse()
//	defer fasthttp.ReleaseResponse(resp)
//
//	if err := c.httpClient.DoTimeout(req, resp, time.Duration(10)*time.Second); err != nil {
//		return "", fmt.Errorf("timeout, error: %w", err)
//	}
//
//	if resp.StatusCode() != fasthttp.StatusOK {
//		return "", fmt.Errorf("wrong status code: %v", resp.StatusCode())
//	}
//
//	var res UploadResponse
//
//	if err := json.Unmarshal(resp.Body(), &res); err != nil {
//		return "", fmt.Errorf("failed to unmarshal response: %w", err)
//	}
//
//	if res.Status != 200 {
//		return "", fmt.Errorf("wrong status code: %v", res.Status)
//	}
//
//	return res.Result.RequestFileId, nil
//}
//
//func (c *Client) RecognizeFile(reqFileId string) (string, error) {
//	payload := map[string]any{
//		"request_file_id": reqFileId,
//		"options": map[string]any{
//			"language":                "ru-RU",
//			"audio_encoding":          "OPUS",
//			"sample_rate":             48000,
//			"hypotheses_count":        1,
//			"enable_profanity_filter": false,
//			"max_speech_timeout":      "20s",
//			"channels_count":          1,
//			"no_speech_timeout":       "7s",
//			"speaker_separation_options": map[string]any{
//				"enable":                   false,
//				"enable_only_main_speaker": false,
//				"count":                    2,
//			},
//		},
//	}
//
//	body, err := json.Marshal(payload)
//
//	if err != nil {
//		return "", fmt.Errorf("failed to marshal payload: %w", err)
//	}
//
//	req := fasthttp.AcquireRequest()
//	req.SetRequestURI(RecognizeUrl)
//	req.Header.SetMethod(fasthttp.MethodPost)
//	req.Header.Add("Accept", "application/json")
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("Authorization", "Bearer "+c.token)
//	req.SetBody(body)
//
//	defer fasthttp.ReleaseRequest(req)
//
//	resp := fasthttp.AcquireResponse()
//	defer fasthttp.ReleaseResponse(resp)
//
//	if err := c.httpClient.DoTimeout(req, resp, time.Duration(10)*time.Second); err != nil {
//		return "", fmt.Errorf("timeout, error: %v", err)
//	}
//
//	if resp.StatusCode() != fasthttp.StatusOK {
//		return "", fmt.Errorf("wrong status code: %v", resp.StatusCode())
//	}
//
//	var res RecognizeResponse
//
//	if err := json.Unmarshal(resp.Body(), &res); err != nil {
//		return "", fmt.Errorf("failed to unmarshal response: %w", err)
//	}
//
//	if res.Status != 200 {
//		return "", fmt.Errorf("wrong status code: %v", res.Status)
//	}
//
//	return res.Result.TaskId, nil
//}
//
//func (c *Client) DownloadFile(reqFileId string) (string, error) {
//	req := fasthttp.AcquireRequest()
//	req.SetRequestURI(DownloadUrl + "?response_file_id=" + reqFileId)
//	req.Header.SetMethod(fasthttp.MethodGet)
//	req.Header.Add("Accept", "application/json")
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("Authorization", "Bearer "+c.token)
//
//	defer fasthttp.ReleaseRequest(req)
//
//	resp := fasthttp.AcquireResponse()
//	defer fasthttp.ReleaseResponse(resp)
//
//	if err := c.httpClient.DoTimeout(req, resp, time.Duration(10)*time.Second); err != nil {
//		return "", fmt.Errorf("timeout, error: %v", err)
//	}
//
//	if resp.StatusCode() != fasthttp.StatusOK {
//		return "", fmt.Errorf("wrong status code: %v", resp.StatusCode())
//	}
//
//	var res []DownloadResponse
//
//	if err := json.Unmarshal(resp.Body(), &res); err != nil {
//		return "", fmt.Errorf("failed to unmarshal response: %w", err)
//	}
//
//	for _, r := range res {
//		return r.Results[0].Text, nil
//	}
//
//	return "", nil
//}

func WithHttpClient(c *http.Client) option {
	return func(m *Client) {
		m.httpClient = c
	}
}

func (c *Client) request(request *http.Request) ([]byte, error) {
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("response status isn't 200 OK returns: %d", resp.StatusCode))
	}

	return io.ReadAll(resp.Body)
}
