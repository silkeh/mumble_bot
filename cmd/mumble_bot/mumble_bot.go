package main

import (
	"flag"
	"log"
	"strings"

	"github.com/silkeh/mumble_bot/mumble"
)

func main() {
	var configFile string

	flag.StringVar(&configFile, "config", "config.yaml", "Configuration file")
	flag.Parse()

	config, err := LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config file %q: %s", configFile, err)
	}

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Error creating client: %s", err)
	}
	defer client.Stop()

	log.Printf("Waiting for events...")
	for c := range client.Mumble.Out {
		if c == mumble.Join && len(client.Mumble.Users) > 1 {
			client.SendSticker("welcome")
		}

		cmd := strings.Split(c, " ")
		switch cmd[0] {
		case "!hold":
			go client.PlayHold(cmd[1:]...)
		case "!volume--":
			client.ChangeVolume(0.5)
		case "!volume++":
			client.ChangeVolume(2)
		case "!stop":
			client.Mumble.StopAudio()
		default:
			client.SendSticker(cmd[0][1:])
		}
	}
}
