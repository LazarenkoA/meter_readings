package tbot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"log"
	"os"
	"time"
)

const (
	startingEnergosbyt = 16 // с 15 дня принимаются показания +1
	startingVodokanal  = 20
)

var (
	errIsPlanning = errors.New("shipping is scheduled")
)

func (t *tBot) start(chatID int64, msg *tgbotapi.Message) {
	txt := fmt.Sprintf("Привет %s %s\n Введите пароль", msg.From.FirstName, msg.From.LastName)

	if !t.checkPass(msg.Text, msg.Chat.ID) {
		txt := []string{fmt.Sprintf("Привет %s %s", msg.From.FirstName, msg.From.LastName)}
		txt = append(txt, "Введите пароль")
	} else {
		t.stepPipe(t.initPipe(), nil, chatID)
		return
	}

	b := &Button{
		caption: "Отмена",
		handler: func(self *Button, selfMsg *tgbotapi.Message) {
			t.msgInterceptor.unsubscribe("start")
			t.deleteMessage(msg.Chat.ID, selfMsg.MessageID)
			t.callback = map[string]func(){}
		},
	}

	rootMsg, err := t.sendMsg(txt, chatID, Buttons{b})
	if err != nil {
		log.Println(errors.Wrap(err, "send msg error"))
		return
	}

	t.msgInterceptor.subscribe("start", func(msg *tgbotapi.Message) (breakProc bool) {
		if msg == nil {
			return false
		}

		t.deleteMessage(msg.Chat.ID, msg.MessageID)
		if t.checkPass(msg.Text, msg.Chat.ID) {
			go t.msgInterceptor.unsubscribe("start")
			t.deleteMessage(msg.Chat.ID, rootMsg.MessageID)
			t.sendTTLMsg("Пароль принят", msg.Chat.ID, Buttons{}, time.Second*10)
		}

		return true
	})
}

func (t *tBot) mosEnergosbytSend(data meter) error {
	if len(data.Electricity) == 0 {
		return errors.New("the testimony was not transmitted")
	}

	T1, T2, T3 := 0, 0, 0

	t1, _ := data.Electricity["T1"]
	t2, _ := data.Electricity["T2"]
	t3, _ := data.Electricity["T3"]

	T1, T2, T3 = int(t1), int(t2), int(t3)

	if T1 == 0 {
		return errors.New("T1 value is not filled in")
	}
	if T2 == 0 {
		return errors.New("T2 value is not filled in")
	}
	if T3 == 0 {
		return errors.New("T3 value is not filled in")
	}

	if time.Now().Day() < startingEnergosbyt {
		if err := t.storage.StoreObject("energosbyt", data); err == nil {
			t.scheduleEnergosbytMessages()
		}

		return errIsPlanning
	}

	if err := t.mosenergosbyt.Auth(); err != nil {
		return errors.Wrap(err, "mosenergosbyt auth error")
	}

	return t.mosenergosbyt.SetReadings(T1, T2, T3)
}

func (t *tBot) mosVodokanalSend() error {

	return errIsPlanning

}

func (t *tBot) scheduleEnergosbytMessages() {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
		}
	}()

	data, err := t.storage.RestoreObject("energosbyt")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Println(errors.Wrap(err, "restore energosbyt data error"))
		}
		return
	}

	send := func(T1, T2, T3 int) (err error) {
		defer func() {
			if err != nil {
				t.sendMsg(html.EscapeString(fmt.Sprintf("не удалось отправить запланированное сообщение в мосэнергосбыт: %s", err)), int64(data["chatid"].(int)), Buttons{})
			} else {
				t.sendMsg("успешно переданы показания в мосэнергосбыт", int64(data["chatid"].(int)), Buttons{})
			}

			t.storage.DeleteObject("energosbyt")
		}()

		if err := t.mosenergosbyt.Auth(); err != nil {
			return errors.Wrap(err, "mosenergosbyt auth error")
		}
		return errors.Wrap(t.mosenergosbyt.SetReadings(T1, T2, T3), "send to mosenergosbyt error")
	}

	v, ok := data["electricity"]
	if !ok {
		log.Println("bad file struct")
		return
	}

	T1, T2, T3 := v.(map[any]any)["T1"].(int), v.(map[any]any)["T2"].(int), v.(map[any]any)["T3"].(int)
	if time.Now().Day() >= startingEnergosbyt {
		send(T1, T2, T3)
	} else {
		targetTime := time.Date(time.Now().Year(), time.Now().Month(), startingEnergosbyt, 9, 0, 0, 0, time.Local)
		time.AfterFunc(time.Until(targetTime), func() {
			send(T1, T2, T3)
		})
	}
}
