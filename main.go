package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
	"meter_readings/giga"
	"meter_readings/tbot"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	_ = godotenv.Load()
	//
	//ds, _ := deepseek.NewDSClient(context.Background(), os.Getenv("DEEPSEEK_API_KEY"))
	//res, _ := ds.GetReminderCharacteristics("напомни завтра утром позавтракать")
	//fmt.Println(res)
	//
	//res, _ = ds.GetReminderCharacteristics("напомни в сл. пятницу в 21ч встреча")
	//fmt.Println(res)
	//
	//res, _ = ds.GetReminderCharacteristics("напоминай каждый месяц 10ч о том что нужно подать показания счетчиков")
	//fmt.Println(res)
	//
	//res, _ = ds.GetReminderCharacteristics("привет как дела")
	//fmt.Println(res)
	//
	//return

	var settings tbot.BotSettings
	if err := envconfig.Process("", &settings); err != nil {
		log.Fatalf("failed to load minio connect configuration: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go shutdown(cancel)

	g, _ := giga.NewGigaClient(context.Background(), os.Getenv("GIGA_API_KEY"))
	bot, err := tbot.NewBot(ctx, settings, tbot.WithAI(g))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	bot.Run()
}

func shutdown(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	cancel()
	log.Println("shutting down")
}
