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

	var settings tbot.BotSettings
	if err := envconfig.Process("", &settings); err != nil {
		log.Fatalf("failed to load minio connect configuration: %s", err)
	}

	//cli := salutespeech.NewClient(context.Background(), settings.GigaAPIKey, salutespeech.WithHttpClient(&http.Client{Transport: &http.Transport{
	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//}}))
	//err := cli.GetToken()
	//fmt.Println(err)

	ctx, cancel := context.WithCancel(context.Background())
	go shutdown(cancel)

	g, _ := giga.NewGigaClient(context.Background(), settings.GigaAPIKey)
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
