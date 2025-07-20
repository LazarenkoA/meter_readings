package deepseek

import "time"

type ReminderCharacteristics struct {
	Topic     string    `json:"topic"`
	RunAt     string    `json:"run_at"`
	RunAtTime time.Time `json:"-"`
	Cron      string    `json:"cron"`
	Error     string    `json:"error"`
}
