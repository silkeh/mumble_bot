package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/silkeh/mumble_bot/mumble"
	"github.com/silkeh/mumble_bot/telegram"
	"github.com/silkeh/mumble_bot/util"
	"github.com/tucnak/telebot"
)

func main() {
	var err error
	var mumbleServer, mumbleUser string
	var telegramToken, telegramTarget string

	flag.StringVar(&mumbleServer, "mumble.server", "localhost:64738", "Mumble server")
	flag.StringVar(&mumbleUser, "mumble.user", "MatrixBot", "Mumble user name")
	flag.StringVar(&telegramToken, "telegram.token", "", "Telegram API token")
	flag.StringVar(&telegramTarget, "telegram.target", "", "Telegram target user/group")
	flag.Parse()

	util.SetStringFromEnv(&mumbleServer, "MUMBLE_SERVER")
	util.SetStringFromEnv(&mumbleUser, "MUMBLE_USER")
	util.SetStringFromEnv(&telegramTarget, "TELEGRAM_TARGET")
	util.SetStringFromEnv(&telegramToken, "TELEGRAM_TOKEN")

	if telegramToken == "" || telegramTarget == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Telegram
	tel, err := telegram.NewClient(telegramToken, telegramTarget)
	if err != nil {
		log.Fatal("Error connecting to Telegram:", err)
	}
	go tel.Start()
	defer tel.Stop()

	// Mumble
	mum, err := mumble.NewClient(mumbleServer, mumbleUser)
	if err != nil {
		log.Fatal("Error connecting to Mumble:", err)
	}
	defer mum.Disconnect()

	log.Printf("Waiting for events...")
	for c := range mum.Out {
		var sticker *telebot.Sticker

		switch {
		// First join
		case c == mumble.Join && len(mum.Users) == 2:
			sticker = telegram.Stickers["welcome"]
		// Command
		case strings.HasPrefix(c, "!"):
			sticker, _ = telegram.Stickers[c[1:]]
		}

		if sticker != nil {
			_, err := tel.SendSticker(sticker)

			if err != nil {
				log.Printf("Error posting sticker: %s", err)
			}
		}
	}
}
