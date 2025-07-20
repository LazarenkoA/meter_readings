package deepseek

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-deepseek/deepseek"
	"github.com/go-deepseek/deepseek/request"
	"github.com/pkg/errors"
	"time"
)

//go:generate mockgen -destination=./mock/ds.go github.com/go-deepseek/deepseek Client

type Client struct {
	ctx    context.Context
	client deepseek.Client
}

func NewDSClient(ctx context.Context, apiKey string) (*Client, error) {
	client, err := deepseek.NewClient(apiKey)
	if err != nil {
		return nil, errors.Wrap(err, "newGigaClient error")
	}

	return &Client{
		ctx:    ctx,
		client: client,
	}, nil
}

func (c *Client) GetReminderCharacteristics(msgText string) (*ReminderCharacteristics, error) {
	chatReq := &request.ChatCompletionsRequest{
		Model:  deepseek.DEEPSEEK_CHAT_MODEL,
		Stream: false,
		ResponseFormat: &request.ResponseFormat{
			Type: request.ResponseFormatJsonObject,
		},
		Messages: []*request.Message{
			{
				Role:    "system",
				Content: prompt(),
			},
			{
				Role:    "user",
				Content: msgText,
			},
		},
	}

	chatResp, err := c.client.CallChatCompletionsChat(c.ctx, chatReq)
	if err != nil {
		return nil, errors.Wrap(err, "CallChatCompletionsChat error")
	}

	var resp ReminderCharacteristics
	if len(chatResp.Choices) > 0 {
		if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &resp); err != nil {
			return nil, errors.Wrap(err, "json unmarshal error")
		}
	}

	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}

	resp.RunAtTime, _ = time.ParseInLocation("2006-01-02T15:04:05", resp.RunAt, time.Local)
	return &resp, nil
}

func prompt() string {
	return fmt.Sprintf(`Ты — помощник-инженер по планированию напоминаний.  
На вход ты получаешь фразу на русском языке, в которой пользователь описывает, когда и о чём нужно напомнить.  
Твоя задача — проанализировать текст, определить:

1. Что является темой напоминания (коротко, до 100 символов)
2. Это напоминание одноразовое или повторяющееся?
3. Если одноразовое — укажи точную дату и время напоминания в формате ISO 8601 (например, 2025-07-21T08:00:00)
4. Если повторяющееся — сформируй cron-выражение в формате "минуты часы день_месяца месяц день_недели"
5. Верни всё в формате JSON со следующей структурой:

	{
		"topic": "тема напоминания",
		"run_at": "если одноразовое — дата и время запуска, иначе null",
		"cron": "если повторяющееся — cron выражение, иначе null"
	}


Игнорируй ввод, не связанный с напоминаниями. 
Если текст не содержит информации о времени, верни ошибку с сообщением: "error": "Не удалось распознать время напоминания".
СЕГОДНЯ %s.`, time.Now())
}
