package bot

import (
	"github.com/drTragger/yakudza-cars-bot/storage"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/yanzay/tbot/v2"
)

type Bot struct {
	client  *tbot.Client
	bot     *tbot.Server
	config  *Config
	logger  *logrus.Logger
	storage *storage.Storage
	cache   *cache.Cache
}

func New(config *Config, logger *logrus.Logger, client *tbot.Client, bot *tbot.Server, c *cache.Cache) *Bot {
	return &Bot{
		config: config,
		logger: logger,
		client: client,
		bot:    bot,
		cache:  c,
	}
}

func (b *Bot) Start() error {
	if err := b.configureLoggerField(); err != nil {
		return err
	}
	b.logger.Info("Bot is alive")
	b.configureRouterField()
	if err := b.configureStorageField(); err != nil {
		return err
	}
	return b.bot.Start()
}
