package deepseek

import "time"

type ReminderCharacteristics struct {
	Topic     string
	RunAt     string `json:"run_at"`
	RunAtTime time.Time
	Cron      string
	Error     string
	Completed bool
}
