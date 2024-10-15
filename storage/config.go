package storage

type Config struct {
	Host     string `toml:"db_host"`
	Port     int16  `toml:"db_port"`
	User     string `toml:"db_user"`
	Password string `toml:"db_password"`
	DataBase string `toml:"db_database"`
}

func NewConfig() *Config {
	return &Config{}
}
