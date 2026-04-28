package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPPort  string `env:"HTTP_PORT" env-default:"8080"`
	DBHost    string `env:"DB_HOST" env-default:"localhost"`
	DBPort    string `env:"DB_PORT" env-default:"5432"`
	DBUser    string `env:"DB_USER" env-default:"pguser"`
	DBPass    string `env:"DB_PASSWORD" env-default:"pgpassword"`
	DBName    string `env:"DB_NAME" env-default:"room_booking"`
	JWTSecret string `env:"JWT_SECRET" env-default:"secret-key"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) DatabaseURL() string {
	return "postgres://" + c.DBUser + ":" + c.DBPass + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=disable"
}
