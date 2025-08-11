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

type reminderInfo struct {
	topic    string
	schedule struct {
		once *struct {
			when time.Time
		}
		repeat *struct {
			cron string
		}
	}
}

type Reminder struct {
	cron         *cron.Cron
	reminderData reminder
	Mx           sync.RWMutex
	DsMx         sync.Mutex
	bot          *tBot
	timer        []*time.Timer
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
		timer:        make([]*time.Timer, 0),
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
	r.Mx.Unlock()

	r.StoreReminderData()
	r.reminderReStart()
	return nil
}

func (r *Reminder) StoreReminderData() {
	r.Mx.Lock()
	defer r.Mx.Unlock()

	for k, rInfo := range r.reminderData {
		for i := len(rInfo) - 1; i >= 0; i-- {
			if rInfo[i].Completed {
				r.reminderData[k] = append(append([]*deepseek.ReminderCharacteristics{}, r.reminderData[k][:i]...), r.reminderData[k][i+1:]...)
			}
		}
	}

	if err := r.bot.storage.StoreObject("reminder", r.reminderData); err != nil {
		log.Println("ERROR: ", err)
	}
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
	for _, t := range r.timer {
		t.Stop()
	}

	r.cron.Stop()
	for _, e := range r.cron.Entries() {
		r.cron.Remove(e.ID)
	}

	r.reminderStart()
}

func (r *Reminder) reminderStart() {
	r.Mx.RLock()
	defer r.Mx.RUnlock()

	for chatID, rInfo := range r.reminderData {
		for _, item := range rInfo {
			if !item.RunAtTime.IsZero() && item.RunAtTime.After(time.Now()) {
				r.timer = append(r.timer, time.AfterFunc(time.Until(item.RunAtTime), func() {
					_, _ = r.bot.sendMsg(item.Topic, chatID, Buttons{})
					item.Completed = true
				}))
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

func (r *Reminder) reminderList(chatID int64) []reminderInfo {
	r.Mx.RLock()
	defer r.Mx.RUnlock()

	result := make([]reminderInfo, 0, len(r.reminderData))
	for _, r := range r.reminderData[chatID] {
		if r.Completed || !r.RunAtTime.IsZero() && r.RunAtTime.Before(time.Now()) {
			continue
		}

		item := reminderInfo{topic: r.Topic}
		if r.Cron != "" {
			item.schedule.repeat = &struct{ cron string }{cron: r.Cron}
		}
		if !r.RunAtTime.IsZero() {
			item.schedule.once = &struct{ when time.Time }{when: r.RunAtTime}
		}

		result = append(result, item)
	}

	return result
}
