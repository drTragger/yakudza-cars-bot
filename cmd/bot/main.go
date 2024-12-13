package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/drTragger/yakudza-cars-bot/internal/app/bot"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/yanzay/tbot/v2"
	"log"
	"os"
	"time"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "path", "configs/config.toml", "Path to configs file in .toml format")
}

func main() {
	config := bot.NewConfig()
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Println("Could not find configs file. Using default values:", err)
	}
	tgBot := tbot.New(config.BotToken)

	logger := logrus.New()

	file, err := os.OpenFile("storage/logs/logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	logger.SetOutput(file)

	server := bot.New(config, logger, tgBot.Client(), tgBot, cache.New(24*time.Hour, 48*time.Hour))

	log.Fatal(server.Start())
}
