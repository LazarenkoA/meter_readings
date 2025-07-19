package giga

import (
	"context"
	"encoding/json"
	"github.com/paulrzcz/go-gigachat"
	"github.com/pkg/errors"
	"log"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/mock.go
type IGigaClient interface {
	AuthWithContext(ctx context.Context) error
	ChatWithContext(ctx context.Context, in *gigachat.ChatRequest) (*gigachat.ChatResponse, error)
	UploadFile(ctx context.Context, filePath string) (string, error)
	DeleteFile(ctx context.Context, fileID string) error
}

type MeterValue struct {
	Value        float64 `json:"value"`
	SerialNumber string  `json:"serial_number"`
	Percent      int     `json:"percent"`
}

type Client struct {
	ctx      context.Context
	client   IGigaClient
	countReq int
}

var (
	ErrFailRecognize = errors.New("failed to recognize the value")
)

func NewGigaClient(ctx context.Context, authKey string) (*Client, error) {
	client, err := gigachat.NewInsecureClientWithAuthKey(authKey)
	if err != nil {
		return nil, errors.Wrap(err, "newGigaClient error")
	}

	return &Client{
		ctx:    ctx,
		client: client,
	}, nil
}

func (c *Client) GetWaterMeter(imgPath string) (float64, error) {
	err := c.client.AuthWithContext(c.ctx)
	if err != nil {
		return 0, errors.Wrap(err, "auth error")
	}

	fileID, err := c.client.UploadFile(c.ctx, imgPath)
	if err != nil {
		return 0, errors.Wrap(err, "upload file error")
	}

	defer c.client.DeleteFile(c.ctx, fileID)

	req := &gigachat.ChatRequest{
		Model: "GigaChat-Max",
		Messages: []gigachat.Message{
			{
				Role:    "system",
				Content: "Ты — языковая модель, выполняющая визуальное распознавание с фото счётчиков воды. Твоя задача — извлечь показания счётчика воды (с точностью до одного знака после запятой) и номер счётчика. Результат верни строго в формате JSON.",
			},
			{
				Role:        "user",
				Content:     waterMeterPrompt(),
				Attachments: []string{fileID},
			},
		},
		Temperature: ptr(0.7),
		//MaxTokens:   ptr[int64](200),
	}

	resp, err := c.client.ChatWithContext(c.ctx, req)
	if err != nil {
		return 0, errors.Wrap(err, "request error")
	}

	var v MeterValue
	if len(resp.Choices) > 0 {
		content := resp.Choices[0].Message.Content
		if err := json.Unmarshal([]byte(content), &v); err != nil {
			log.Println("GIGA content:", content)
			return 0, errors.Wrap(err, "json unmarshal error")
		}
	}

	if v.Percent < 90 {
		log.Println("GIGA percent:", v.Percent)
		return 0, ErrFailRecognize
	}

	return v.Value, nil
}

func ptr[T any](v T) *T {
	return &v
}

func waterMeterPrompt() string {
	return `На изображении фото счетчика воды. Если на изображении нет счетчика, верни пустой json.
Твоя задача найти и распознать показания на счетчике.
Показания отображаются НА МЕХАНИЧЕСКОМ ТАБЛО с чёрными и красными цифрами. Необходимо учитывать все чёрные цифры, включая одну красную после запятой.

Верни данные строго в формате JSON (пример ниже).
JSON должен содержать поля: 
value - показания счетчика
percent - процент уверенности распознавания значений (0 — ничего не удалось распознать, 100 — абсолютно уверен).

Если ты не уверен, не придумывай значения. Лучше верни "value": null.
Никогда не генерируй значения от себя — только с изображения.

	ПРИМЕР успешного распознавания:
	{
		"value": 2.3,
		"percent":  99
	}

	ПРИМЕР если не удалось распознать показания:
	{
		"value": null,
		"percent": 0
	}

Убедись, что возвращаешь только JSON, без пояснений, комментариев и текста вокруг. Начинай распознавание`
}
