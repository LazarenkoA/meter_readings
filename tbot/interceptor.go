package tbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
)

type readUpdate func(msg *tgbotapi.Message) (breakProc bool)

type msgInterceptor struct {
	consumers map[string]readUpdate
	mx        sync.RWMutex
}

func (i *msgInterceptor) subscribe(key string, c readUpdate) {
	i.mx.Lock()
	defer i.mx.Unlock()

	i.consumers[key] = c
}

func (i *msgInterceptor) unsubscribe(key string) {
	i.mx.Lock()
	defer i.mx.Unlock()

	delete(i.consumers, key)
}

func (i *msgInterceptor) notify(msg *tgbotapi.Message) (breakProc bool) {
	i.mx.RLock()
	defer i.mx.RUnlock()

	for _, c := range i.consumers {
		if c(msg) {
			return true
		}
	}

	return false
}
