package config

import (
	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Port        string   `env:"PORT" envDefault:"9998"`
	Issuer      string   `env:"ISSUER" envDefault:"http://localhost:9998"`
	RedirectURI []string `env:"REDIRECT_URI"`
	UsersFile   string   `env:"USERS_FILE"`
}

func Parse() (Config, error) {
	return env.ParseAs[Config]()
}
