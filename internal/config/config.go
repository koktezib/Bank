package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	DBHost, DBPort, DBUser, DBPass, DBName               string
	JWTSecret                                            string
	SMTPHost, SMTPUser, SMTPPass                         string
	SMTPPort                                             int
	PGPPrivateKey, PGPPublicKey, PGPPrivateKeyPassphrase string
	HMACSecret                                           string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Нет .env-файла, читаем из окружения")
	}
	return &Config{
		DBHost:                  os.Getenv("DB_HOST"),
		DBPort:                  os.Getenv("DB_PORT"),
		DBUser:                  os.Getenv("DB_USER"),
		DBPass:                  os.Getenv("DB_PASS"),
		DBName:                  os.Getenv("DB_NAME"),
		JWTSecret:               os.Getenv("JWT_SECRET"),
		SMTPHost:                os.Getenv("SMTP_HOST"),
		SMTPPort:                atoiOrDefault(os.Getenv("SMTP_PORT"), 587),
		SMTPUser:                os.Getenv("SMTP_USER"),
		SMTPPass:                os.Getenv("SMTP_PASS"),
		HMACSecret:              os.Getenv("HMAC_SECRET"),
		PGPPrivateKey:           os.Getenv("PGP_PRIVATE_KEY"),
		PGPPublicKey:            os.Getenv("PGP_PUBLIC_KEY"),
		PGPPrivateKeyPassphrase: os.Getenv("PGP_PASSPHRASE"),
	}
}

func atoiOrDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}
