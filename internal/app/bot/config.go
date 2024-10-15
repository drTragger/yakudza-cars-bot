package bot

import "github.com/drTragger/yakudza-cars-bot/storage"

type Config struct {
	BotToken    string `toml:"bot_token"`
	LoggerLevel string `toml:"logger_level"`
	Storage     *storage.Config
	Admin       AdminConfig
	GroupID     string `toml:"group_id"`
}

type AdminConfig struct {
	AdminIDs []string `toml:"admin_ids"`
}

func NewConfig() *Config {
	return &Config{
		LoggerLevel: "debug",
		Storage:     storage.NewConfig(),
	}
}
