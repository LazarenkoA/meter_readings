package tbot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type pipe struct {
	msg     string
	name    string
	choice  map[string]*pipe
	prev    *pipe
	value   any // произвольное значение для сохранения на шаге
	handler func(p *pipe, rootMsg *tgbotapi.Message)
}

var (
	mapMetersNum sync.Map
)

func (t *tBot) initPipe() *pipe {
	metersChoice := make(map[string]*pipe)
	valueGetter := func(currentStep *pipe, rootMsg *tgbotapi.Message) {
		t.msgInterceptor.subscribe("pipe", func(msg *tgbotapi.Message) (breakProc bool) {
			if msg == nil {
				return false
			}

			t.deleteMessage(msg.Chat.ID, msg.MessageID)

			if len(msg.Photo) > 0 {
				v := t.imageRecognize(msg)
				currentStep.value = v
				t.editMsg(rootMsg, buildMsgTxt(currentStep), nil)
				return
			}

			if v, err := strconv.ParseFloat(msg.Text, 64); err == nil {
				currentStep.value = v
			}

			t.editMsg(rootMsg, buildMsgTxt(currentStep), nil)
			return true
		})
	}

	go t.getMeters(&metersChoice, valueGetter)

	return &pipe{
		msg: fmt.Sprintf("Показания текущего периода (%s):\n%s", time.Now().Format("01.2006"), t.readMeter()),
		choice: map[string]*pipe{
			"Заполнить электр...": {
				msg: "Выберите T1, T2, T3",
				choice: map[string]*pipe{
					"T1": {name: "T1", msg: "Введите значение T1 или отправьте фото счетчика", handler: valueGetter},
					"T2": {name: "T2", msg: "Введите значение T2 или отправьте фото счетчика", handler: valueGetter},
					"T3": {name: "T3", msg: "Введите значение T3 или отправьте фото счетчика", handler: valueGetter},
				},
				handler: func(*pipe, *tgbotapi.Message) {},
				value:   map[string]float64{},
			},
			"Заполнить воду": {
				msg:     "Выберите счетчик",
				choice:  metersChoice,
				handler: func(*pipe, *tgbotapi.Message) {},
				value:   map[string]float64{},
			},
			"Отправить": {
				handler: func(p *pipe, selfMsg *tgbotapi.Message) {
					t.sendMeter(p.prev, selfMsg.Chat.ID)
					t.deleteMessage(selfMsg.Chat.ID, selfMsg.MessageID)
					t.msgInterceptor.unsubscribe("pipe")
					t.callback = map[string]func(){}
				},
			},
		},
		handler: func(*pipe, *tgbotapi.Message) {},
	}
}

func (t *tBot) getMeters(metersChoice *map[string]*pipe, valueGetter func(currentStep *pipe, rootMsg *tgbotapi.Message)) {
	meters, err := t.mosRu.GetMeters(t.ctx)
	if err != nil {
		log.Println(fmt.Sprintf("get water meters error: %v", err))
	}

	for _, m := range meters {
		id, _ := m["id"]
		number, ok := m["number"]
		if !ok {
			continue
		}

		numberString, _ := number.(string)
		mapMetersNum.Store(numberString, id.(string))

		(*metersChoice)[numberString] = &pipe{name: numberString, msg: fmt.Sprintf("Введите значение счетчика %s или отправьте фото счетчика", numberString), handler: valueGetter}
	}
}

func (t *tBot) sendMeter(p *pipe, chatID int64) {
	data := meter{
		Water:       p.choice["Заполнить воду"].value.(map[string]float64),
		Electricity: p.choice["Заполнить электр..."].value.(map[string]float64),
		ChatID:      chatID,
	}

	if err := t.mosEnergosbytSend(data); err != nil && !errors.Is(err, errIsPlanning) {
		t.sendMsg(errors.Wrap(err, "mosenergosbyt send error").Error(), chatID, Buttons{})
	} else if errors.Is(err, errIsPlanning) {
		t.sendTTLMsg(fmt.Sprintf("Отправка электроэнергии запланирована на %d число", startingEnergosbyt), chatID, Buttons{}, time.Second*30)
	} else {
		t.sendTTLMsg("Показания по электроэнергии успешно отправлены", chatID, Buttons{}, time.Second*10)
	}

	if err := t.mosRuSend(data); err != nil && !errors.Is(err, errIsPlanning) {
		t.sendMsg(errors.Wrap(err, "mosRuSend send error").Error(), chatID, Buttons{})
	} else if errors.Is(err, errIsPlanning) {
		t.sendTTLMsg(fmt.Sprintf("Отправка воды запланирована на %d число", startingVodokanal), chatID, Buttons{}, time.Second*30)
	} else {
		t.sendTTLMsg("Показания по воде успешно отправлены", chatID, Buttons{}, time.Second*10)
	}
}

func (t *tBot) stepPipe(p *pipe, rootMsg *tgbotapi.Message, chatID int64) {
	if p == nil {
		return
	}

	buttons := make(Buttons, 0, len(p.choice))
	for k, v := range p.choice {
		b := &Button{
			caption: k,
			handler: func(self *Button, selfMsg *tgbotapi.Message) {
				v.prev = p
				t.stepPipe(v, selfMsg, chatID)
			},
		}

		buttons = append(buttons, b)
	}

	sort.Slice(buttons, func(i, j int) bool {
		return strings.Compare(buttons[i].caption, buttons[j].caption) < 0
	})

	if p.prev != nil {
		buttons = append(buttons, &Button{
			caption: "Назад",
			handler: func(self *Button, selfMsg *tgbotapi.Message) {
				t.msgInterceptor.unsubscribe("pipe")
				if tmp, ok := p.prev.value.(map[string]float64); ok && p.value != nil {
					tmp[p.name] = p.value.(float64)
				}
				t.stepPipe(p.prev, selfMsg, chatID)
			},
		})
	}

	cancel := &Button{
		caption: "Отмена",
		handler: func(self *Button, selfMsg *tgbotapi.Message) {
			t.deleteMessage(selfMsg.Chat.ID, selfMsg.MessageID)
			t.msgInterceptor.unsubscribe("pipe")
			t.callback = map[string]func(){}
		},
	}

	buttons = append(buttons, cancel)
	if rootMsg == nil {
		if _, err := t.sendMsg(p.msg, chatID, buttons); err != nil {
			fmt.Println(errors.Wrap(err, "send msg error").Error())
		}
	} else {
		m, err := t.editMsg(rootMsg, buildMsgTxt(p), buttons)
		if err != nil {
			log.Println(errors.Wrap(err, "edit msg error").Error())
			m = rootMsg
		}
		p.handler(p, m)
	}
}

func (t *tBot) readMeter() string {
	b := strings.Builder{}

	data, err := t.storage.RestoreObject("energosbyt")
	if err == nil {
		v, ok := data["electricity"]
		if !ok {
			log.Println("bad file struct")
			return ""
		}

		b.WriteString("Электричество:\n")
		T1, T2, T3 := v.(map[any]any)["T1"].(int), v.(map[any]any)["T2"].(int), v.(map[any]any)["T3"].(int)
		b.WriteString(fmt.Sprintf("\t%d\n", T1))
		b.WriteString(fmt.Sprintf("\t%d\n", T2))
		b.WriteString(fmt.Sprintf("\t%d\n", T3))
	}

	return b.String()
}

func (t *tBot) imageRecognize(msg *tgbotapi.Message) float64 {
	if len(msg.Photo) == 0 {
		return 0
	}

	if t.ai == nil {
		t.sendTTLMsg("для работы с распознованием фото нужно подключить AI", msg.Chat.ID, Buttons{}, time.Second*2)
		return 0
	}

	imgPath, err := t.downloadImg(msg.Photo[len(msg.Photo)-1])
	if err != nil {
		t.sendTTLMsg(fmt.Sprintf("не удалось скачать файл: %s", err.Error()), msg.Chat.ID, Buttons{}, time.Second*2)
		return 0
	}

	v, err := t.ai.GetWaterMeter(imgPath)
	if err != nil {
		log.Printf("imageRecognize error: %s\n", err.Error())
		t.sendTTLMsg("не удалось распознать значение", msg.Chat.ID, Buttons{}, time.Second*2)
		return 0
	}

	return v
}

func (t *tBot) downloadImg(photo tgbotapi.PhotoSize) (string, error) {
	file, err := t.bot.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
	if err != nil {
		return "", errors.Wrap(err, "get file error")
	}

	fileURL, err := t.bot.GetFileDirectURL(file.FileID)
	if err != nil {
		return "", err
	}

	return t.downloadFile(fileURL, filepath.Base(file.FilePath))
}

func buildMsgTxt(p *pipe) string {
	msg := []string{p.msg}
	if p.value != nil {
		msg = append(msg, fmt.Sprintf("Текущее значение: \n%s", valAsString(p.value)))
	}

	return strings.Join(msg, "\n")
}

func valAsString(v any) string {
	switch t := v.(type) {
	case int, int64, int32:
		return fmt.Sprintf("%d", t)
	case float32, float64:
		return fmt.Sprintf("%.2f", t)
	case []string:
		return strings.Join(t, "\n")
	case map[string]string:
		tmp := make([]string, 0, len(t))
		for k, v := range t {
			tmp = append(tmp, fmt.Sprintf("%s - %s", k, v))
		}
		return strings.Join(tmp, "\n")
	case map[string]float64:
		tmp := make([]string, 0, len(t))
		for k, v := range t {
			tmp = append(tmp, fmt.Sprintf("%s - %.2f", k, v))
		}
		return strings.Join(tmp, "\n")
	}

	return fmt.Sprintf("%v", v)
}
