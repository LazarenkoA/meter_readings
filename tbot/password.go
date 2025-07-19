package tbot

import (
	"github.com/pkg/errors"
	"log"
	"strconv"
	"sync"
)

var (
	verified = map[string]bool{}
	mx       sync.RWMutex
	one      sync.Once
)

const (
	verifiedFileName = "verified"
)

func (t *tBot) checkPass(p string, chatID int64) bool {
	mx.Lock()
	defer mx.Unlock()

	one.Do(func() {
		if data, err := t.storage.RestoreObject(verifiedFileName); err == nil {
			verified = castMap[bool](data)
		}
	})

	key := strconv.Itoa(int(chatID))
	if v, _ := verified[key]; v {
		return true
	}

	if v := p == t.pass; v {
		verified[key] = v
		return true
	}

	t.closer.Append("verifiedFileName", func() {
		err := t.storage.StoreObject(verifiedFileName, verified)
		if err != nil {
			log.Println(errors.Wrap(err, "store verified file error"))
		}
	})

	return false
}
