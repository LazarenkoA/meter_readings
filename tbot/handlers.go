package tbot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"log"
	"meter_readings/node_mos_ru"
	"os"
	"strings"
	"time"
)

const (
	startingEnergosbyt = 16 // с 15 дня принимаются показания +1
	startingVodokanal  = 10
	payerCode          = "2290715619"
	flat               = "82"
)

var (
	errIsPlanning = errors.New("shipping is scheduled")
)

func (t *tBot) start(msg *tgbotapi.Message) {
	txt := fmt.Sprintf("Привет %s %s\n Введите пароль", msg.From.FirstName, msg.From.LastName)

	if !t.checkPass(msg.Text, msg.Chat.ID) {
		txt := []string{fmt.Sprintf("Привет %s %s", msg.From.FirstName, msg.From.LastName)}
		txt = append(txt, "Введите пароль")
	} else {
		t.stepPipe(t.initPipe(), nil, msg.Chat.ID)
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

	rootMsg, err := t.sendMsg(txt, msg.Chat.ID, Buttons{b})
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

func (t *tBot) reminderList(chatID int64) {
	list := t.reminder.reminderList(chatID)
	str := strings.Builder{}

	var pReminders []string
	var oReminders []string

	for _, l := range list {
		if l.schedule.once != nil {
			oReminders = append(oReminders, fmt.Sprintf("<code>%s</code> - %s", l.schedule.once.when.Format("02-01-2006 15:04:05"), l.topic))
		}
		if l.schedule.repeat != nil {
			pReminders = append(pReminders, fmt.Sprintf("<code>%s</code> - %s", l.schedule.repeat.cron, l.topic))
		}
	}

	if len(pReminders) > 0 {
		str.WriteString("Периодические напоминания:\n")
		str.WriteString(strings.Join(pReminders, "\n"))
	}

	if len(oReminders) > 0 {
		str.WriteString("\n\nПредстоящие напоминания:\n")
		str.WriteString(strings.Join(oReminders, "\n"))
	}

	t.sendTTLMsg(str.String(), chatID, Buttons{}, time.Second*30)
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

func (t *tBot) mosRuSend(data meter) error {
	if time.Now().Day() < startingVodokanal {
		if err := t.storage.StoreObject("vodokanal", data); err == nil {
			t.scheduleVodokanalMessages()
		}

		return errIsPlanning
	}

	return t.mosRu.SendReadingsWater(t.ctx, t.buildDataReadings(data))
}

func (t *tBot) buildDataReadings(data meter) *node_mos_ru.Readings {
	result := &node_mos_ru.Readings{
		PayerCode: payerCode,
		Flat:      flat,
		Items:     make([]node_mos_ru.ReadingsItem, 0, len(data.Water)),
	}

	for k, v := range data.Water {
		deviceId, ok := mapMetersNum.Load(k)
		if !ok {
			fmt.Sprintf("ERROR: no id is defined for the device %s", k)
			continue
		}

		result.Items = append(result.Items, node_mos_ru.ReadingsItem{
			Indication: v,
			DeviceId:   deviceId.(string),
			Period:     EndYear().Format(time.DateOnly),
		})
	}

	return result
}

func (t *tBot) scheduleVodokanalMessages() {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
		}
	}()

	restoredData, err := t.storage.RestoreObject("vodokanal")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Println(errors.Wrap(err, "restore vodokanal data error"))
		}
		return
	}

	send := func(data meter) (err error) {
		defer func() {
			if err != nil {
				t.sendMsg(html.EscapeString(fmt.Sprintf("не удалось отправить запланированное сообщение в mos.ru: %s", err)), data.ChatID, Buttons{})
			} else {
				t.sendMsg("успешно переданы показания по воде", data.ChatID, Buttons{})
			}

			t.storage.DeleteObject("vodokanal")
		}()

		return t.mosRu.SendReadingsWater(t.ctx, t.buildDataReadings(data))
	}

	cast := func(object map[string]interface{}) meter {
		water, ok := object["water"]
		if !ok {
			return meter{}
		}

		result := meter{
			Water:  map[string]float64{},
			ChatID: int64(object["chatid"].(int)),
		}

		for k, v := range water.(map[interface{}]interface{}) {
			result.Water[k.(string)], _ = v.(float64)
		}

		return result
	}

	if time.Now().Day() >= startingVodokanal {
		send(cast(restoredData))
	} else {
		targetTime := time.Date(time.Now().Year(), time.Now().Month(), startingVodokanal, 9, 0, 0, 0, time.Local)
		time.AfterFunc(time.Until(targetTime), func() {
			send(cast(restoredData))
		})
	}
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
