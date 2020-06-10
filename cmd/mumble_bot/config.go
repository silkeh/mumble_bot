package main

import (
	"github.com/silkeh/mumble_bot/matrix"
	"github.com/tucnak/telebot"
	"gopkg.in/yaml.v2"
	"os"
)

// MumbleConfig represents configuration for a Mumble client.
type MumbleConfig struct {
	Server, User string
	HoldMusic map[string]string `yaml:"hold_music"`
}

// TelegramConfig represents configuration for a Telegram client.
type TelegramConfig struct {
	Token, Target string
	Stickers map[string]*telebot.Sticker
}

// MatrixConfig represents configuration for a Matrix client.
type MatrixConfig struct {
	Server, User, Token, Room string
	Stickers map[string]*matrix.Sticker
}

// Config represents configuration for a Client.
type Config struct {
	Mumble *MumbleConfig
	Matrix *MatrixConfig
	Telegram *TelegramConfig
}

// LoadConfig loads a YAML configuration file.
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	config := new(Config)
	yaml.NewDecoder(file).Decode(&config)
	return config, nil
}
