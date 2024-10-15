package bot

import (
	"github.com/drTragger/yakudza-cars-bot/internal/app/models"
	"github.com/drTragger/yakudza-cars-bot/storage"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yanzay/tbot/v2"
)

const StartYear = 2014

var (
	userStates    = make(map[string]string)
	carData       = make(map[string]*models.CarDetails)
	carOption     = &models.CarOption{}
	feedbackToAdd = &models.Feedback{}
	prices        = []*models.PriceRange{
		{Title: "До 10 000$", Min: 0, Max: 9999},
		{Title: "10 000$ - 15 000$", Min: 10000, Max: 14999},
		{Title: "15 000$ - 20 000$", Min: 15000, Max: 19999},
		{Title: "20 000$ - 30 000$", Min: 20000, Max: 29999},
		{Title: "30 000$ - 35 000$", Min: 30000, Max: 34999},
		{Title: "35 000$ +", Min: 35000, Max: 1000000},
	}
	shownOptionIDs   = make(map[string][]int)
	shownFeedbackIDs = make(map[string][]int)
	carDataMutex     sync.Mutex
)

type Bot struct {
	client  *tbot.Client
	bot     *tbot.Server
	config  *Config
	logger  *logrus.Logger
	storage *storage.Storage
}

func New(config *Config, logger *logrus.Logger, client *tbot.Client, bot *tbot.Server) *Bot {
	return &Bot{
		config: config,
		logger: logger,
		client: client,
		bot:    bot,
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

func (b *Bot) LogHandler(m *tbot.Message, answer string) {
	location, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		b.logger.Info("Failed loading location ", err.Error())
	}

	b.logger.Printf("%s\nUsername: %s\nChat ID: %s\nMessage: %s\nAnswer: %s", time.Now().In(location).Format(TimeLayout), m.From.Username, m.Chat.ID, m.Text, answer)
}

func (b *Bot) LogCallbackHandler(cq *tbot.CallbackQuery, answer string) {
	location, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		b.logger.Info("Failed loading location ", err.Error())
	}

	b.logger.Printf("%s\nUsername: %s\nChat ID: %s\nMessage: %s\nnAnswer: %s", time.Now().In(location).Format(TimeLayout), cq.From.Username, cq.Message.Chat.ID, cq.Message.Text, answer)
}
