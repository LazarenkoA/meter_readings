package tbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	uuid "github.com/google/uuid"
)

type BHandler func(*Button, *tgbotapi.Message)

type Button struct {
	caption string
	handler BHandler
	id      string
}

type Buttons []*Button

func (bts Buttons) createButtons(msg tgbotapi.Chattable, countColum int) {
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	var Buttons []tgbotapi.InlineKeyboardButton

	switch v := msg.(type) {
	case *tgbotapi.EditMessageTextConfig:
		v.ReplyMarkup = &keyboard
	case *tgbotapi.EditMessageCaptionConfig:
		v.ReplyMarkup = &keyboard
	case *tgbotapi.MessageConfig:
		v.ReplyMarkup = &keyboard
	case *tgbotapi.PhotoConfig:
		v.ReplyMarkup = &keyboard
	}

	for i, _ := range bts {
		currentButton := bts[i]
		if currentButton.id == "" {
			currentButton.id = uuid.NewString()
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(currentButton.caption, currentButton.id)
		Buttons = append(Buttons, btn)
	}

	keyboard.InlineKeyboard = bts.breakButtonsByColum(Buttons, countColum)
}

func (bts Buttons) breakButtonsByColum(Buttons []tgbotapi.InlineKeyboardButton, countColum int) [][]tgbotapi.InlineKeyboardButton {
	end := 0
	var result [][]tgbotapi.InlineKeyboardButton

	for i := 1; i <= int(float64(len(Buttons)/countColum)); i++ {
		end = i * countColum
		start := (i - 1) * countColum
		if end > len(Buttons) {
			end = len(Buttons)
		}

		row := tgbotapi.NewInlineKeyboardRow(Buttons[start:end]...)
		result = append(result, row)
	}
	if len(Buttons)%countColum > 0 {
		row := tgbotapi.NewInlineKeyboardRow(Buttons[end:len(Buttons)]...)
		result = append(result, row)
	}

	return result
}
