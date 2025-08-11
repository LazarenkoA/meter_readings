package tbot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"meter_readings/deepgram"
	"meter_readings/deepseek"
	"meter_readings/mosenergosbyt"
	"meter_readings/node_mos_ru"
	"meter_readings/storage"
	"os"
	"time"
)

type TCallback map[string]func()

type options func(tb *tBot)

type tBot struct {
	bot            *tgbotapi.BotAPI
	callback       TCallback
	msgInterceptor msgInterceptor
	mosenergosbyt  IMosenergosbyt
	mosRu          IMos
	deepseek       IDeepseek
	deepgram       IDeepgram
	storage        storage.IStorage
	ctx            context.Context
	closer         *Closer
	ai             IAI
	pass           string
	reminder       *Reminder
}

func NewBot(ctx context.Context, settings BotSettings, opt ...options) (*tBot, error) {
	bot, err := tgbotapi.NewBotAPI(settings.Token)
	if err != nil {
		return nil, err
	}

	deepseek, _ := deepseek.NewDSClient(ctx, settings.DeepseekAPI)
	tb := &tBot{
		mosenergosbyt:  mosenergosbyt.NewClient(ctx, settings.MosELogin, settings.MosEPass),
		mosRu:          node_mos_ru.NewMosruAdapter(settings.MosRULogin, settings.MosRUPass),
		deepgram:       deepgram.NewDeepgram(settings.DeepgramKey),
		deepseek:       deepseek,
		ctx:            ctx,
		bot:            bot,
		callback:       make(TCallback),
		msgInterceptor: msgInterceptor{consumers: map[string]readUpdate{}},
		storage:        storage.NewFileStorage(),
		closer:         NewCloser(),
		pass:           settings.Pass,
	}

	for _, f := range opt {
		f(tb)
	}

	tb.reminder = NewReminder(tb)
	tb.closer.Append("storeReminderData", tb.reminder.StoreReminderData)

	return tb, nil
}

func (t *tBot) Run() {
	defer func() {
		t.closer.Close()
	}()

	t.scheduleEnergosbytMessages()
	t.scheduleVodokanalMessages()
	t.reminder.restoreReminderData()

	wdUpdate := t.run()
	for {
		var update tgbotapi.Update

		select {
		case <-t.ctx.Done():
			log.Println("bot stopped")
			return
		case update = <-wdUpdate:
		}

		msg := t.getMessage(update)
		if msg == nil {
			continue
		}

		// обработка команд кнопок
		if t.callbackQuery(update) {
			continue
		}

		if t.msgInterceptor.notify(msg) {
			continue
		}

		command := msg.Command()
		switch command {
		case "start":
			t.start(msg)
			continue
		case "reminder_list":
			t.reminderList(msg.Chat.ID)
			continue
		}

		if msg.Voice != nil && t.deepgram != nil {
			if filePath, err := t.downloadAudio(msg.Voice.FileID); err == nil {
				msg.Text, err = t.deepgram.STT(t.ctx, filePath)
				if err != nil {
					log.Println("deepgram STT error: ", err)
				}
				os.Remove(filePath)
			} else {
				log.Println("download voice error: ", err)
			}
		}

		if msg.Text != "" {
			go func() {
				if err := t.reminder.recognizeReminder(msg.Text, msg.Chat.ID); err != nil {
					t.sendMsg(fmt.Sprintf("Ошибка: %v", err), msg.Chat.ID, Buttons{})
				} else {
					t.deleteMessage(msg.Chat.ID, msg.MessageID)
					t.sendTTLMsg("👍🏻", msg.Chat.ID, Buttons{}, time.Second*10)
				}
			}()
		}
	}
}

func (t *tBot) run() tgbotapi.UpdatesChannel {
	_, _ = t.bot.Request(&tgbotapi.DeleteWebhookConfig{})

	dir, _ := os.Getwd()
	log.Println("bot running. Current working directory:", dir)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.AllowedUpdates = []string{"message", "chat_member", "callback_query", "chat_join_request", "my_chat_member"}
	return t.bot.GetUpdatesChan(u)
}

func (t *tBot) getMessage(update tgbotapi.Update) *tgbotapi.Message {
	if update.Message != nil {
		return update.Message
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.Message
	} else {
		return nil
	}
}

func (t *tBot) sendTTLMsg(msg string, chatID int64, buttons Buttons, ttl time.Duration) (*tgbotapi.Message, error) {
	m, err := t.sendMsg(msg, chatID, buttons)
	if err != nil {
		return nil, err
	}

	go func() {
		time.Sleep(ttl)
		t.deleteMessage(chatID, m.MessageID)
	}()

	return m, nil
}

func (t *tBot) sendMsg(msg string, chatID int64, buttons Buttons) (*tgbotapi.Message, error) {
	newmsg := tgbotapi.NewMessage(chatID, msg)
	newmsg.ParseMode = "HTML"
	return t.createButtonsAndSend(&newmsg, buttons)
}

func (t *tBot) deleteMessage(chatID int64, messageID int) {
	conf := tgbotapi.DeleteMessageConfig{
		ChatID:    chatID,
		MessageID: messageID,
	}

	if _, err := t.bot.Request(conf); err != nil {
		log.Println(errors.Wrap(err, "delete msg error"))
	}
}

func (t *tBot) createButtonsAndSend(msg tgbotapi.Chattable, buttons Buttons) (*tgbotapi.Message, error) {
	if len(buttons) > 0 {
		buttons.createButtons(msg, 3)
	}

	m, err := t.bot.Send(msg)
	if err == nil {
		for _, b := range buttons {
			t.callback[b.id] = func() {
				b.handler(b, &m)
			}
		}
	}

	return &m, err
}

func (t *tBot) callbackQuery(update tgbotapi.Update) bool {
	if update.CallbackQuery == nil || update.CallbackQuery.Message == nil {
		return false
	}

	if call, ok := t.callback[update.CallbackQuery.Data]; ok {
		call()
		delete(t.callback, update.CallbackQuery.Data)
	}

	return true
}

func (t *tBot) editMsg(msg *tgbotapi.Message, txt string, buttons Buttons) (*tgbotapi.Message, error) {
	editmsg := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, txt)
	editmsg.ParseMode = "HTML"

	if buttons == nil {
		editmsg.ReplyMarkup = msg.ReplyMarkup
		m, err := t.bot.Send(editmsg)
		return &m, err
	}

	return t.createButtonsAndSend(&editmsg, buttons)
}

func (t *tBot) downloadAudio(fileID string) (string, error) {
	log.Println("download voice message")

	file, err := t.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", errors.Wrap(err, "bot GetFile")
	}

	fileURL := file.Link(t.bot.Token)
	return downloadFile(t.ctx, fileURL)
}

func WithAI(ai IAI) options {
	return func(tb *tBot) {
		tb.ai = ai
	}
}
