package tbot

import (
	"github.com/hanagantig/cron"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"log"
	"meter_readings/deepseek"
	"sync"
	"time"
)

type Reminder struct {
	cron         *cron.Cron
	reminderData reminder
	Mx           sync.RWMutex
	DsMx         sync.Mutex
	bot          *tBot
	timer        *time.Timer
}

type reminder map[int64][]*deepseek.ReminderCharacteristics

var (
	errFrequentRequests = errors.New("requests are too frequent")
)

func NewReminder(bot *tBot) *Reminder {
	return &Reminder{
		bot:          bot,
		cron:         cron.New(),
		reminderData: reminder{},
	}
}

func (r *Reminder) recognizeReminder(txt string, chatID int64) error {
	if r.bot.deepseek == nil {
		return nil
	}

	if r.DsMx.TryLock() {
		defer r.DsMx.Unlock()
	} else {
		return errFrequentRequests // если уже распозноется какое-то сообщение, остальные скипаются
	}

	info, err := r.bot.deepseek.GetReminderCharacteristics(txt)
	if err != nil {
		return err
	}

	r.Mx.Lock()
	r.reminderData[chatID] = append(r.reminderData[chatID], info)
	if err := r.bot.storage.StoreObject("reminder", r.reminderData); err != nil {
		r.Mx.Unlock()
		return err
	}
	r.Mx.Unlock()

	r.reminderReStart()
	return nil
}

func (r *Reminder) restoreReminderData() {
	err := r.bot.storage.RestoreAsObject("reminder", func(data []byte) error {
		r.Mx.Lock()
		defer r.Mx.Unlock()

		return yaml.Unmarshal(data, &r.reminderData)
	})

	if err != nil {
		log.Println(errors.Wrap(err, "RestoreObject error"))
		return
	}

	r.reminderReStart()
}

func (r *Reminder) reminderReStart() {
	if r.timer != nil {
		r.timer.Stop()
	}

	r.cron.Stop()

	r.reminderStart()
}

func (r *Reminder) reminderStart() {
	r.Mx.RLock()
	defer r.Mx.RUnlock()

	for chatID, rInfo := range r.reminderData {
		for _, item := range rInfo {
			if !item.RunAtTime.IsZero() && item.RunAtTime.After(time.Now()) {
				r.timer = time.AfterFunc(time.Until(item.RunAtTime), func() {
					_, _ = r.bot.sendMsg(item.Topic, chatID, Buttons{})
				})
				continue
			}

			if item.Cron != "" {
				_, err := r.cron.AddFunc(item.Cron, "", func() {
					_, _ = r.bot.sendMsg(item.Topic, chatID, Buttons{})
				})
				if err != nil {
					log.Println(errors.Wrap(err, "cron error"))
				}
			}
		}
	}

	if len(r.cron.Entries()) > 0 {
		r.cron.Start()
	}
}
