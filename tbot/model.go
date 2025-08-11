package tbot

import (
	"context"
	"meter_readings/deepseek"
	"meter_readings/node_mos_ru"
)

type BotSettings struct {
	Token       string `envconfig:"BOT_TOKEN"`
	Pass        string `envconfig:"PASSWORD"`
	MosELogin   string `envconfig:"MOS_ENERG_LOGIN"`
	MosEPass    string `envconfig:"MOS_ENERG_PASS"`
	MosRULogin  string `envconfig:"MOS_RU_LOGIN"`
	MosRUPass   string `envconfig:"MOS_RU_PASS"`
	DeepseekAPI string `envconfig:"DEEPSEEK_API_KEY"`
	GigaAPIKey  string `envconfig:"GIGA_API_KEY"`
	DeepgramKey string `envconfig:"DEEPGRAM_KEY"`
}

type meter struct {
	Water       map[string]float64
	Electricity map[string]float64
	ChatID      int64
}

type IMosenergosbyt interface {
	SetReadings(T1, T2, T3 int) error
	Auth() error
}

type IMos interface {
	SendReadingsWater(ctx context.Context, data *node_mos_ru.Readings) error
	GetMeters(ctx context.Context) ([]map[string]interface{}, error)
}

type IAI interface {
	GetWaterMeter(imgPath string) (float64, error)
}

type IDeepseek interface {
	GetReminderCharacteristics(msgText string) (*deepseek.ReminderCharacteristics, error)
}

type IDeepgram interface {
	STT(ctx context.Context, filePath string) (string, error)
}
