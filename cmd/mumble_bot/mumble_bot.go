package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/silkeh/mumble_bot/bot"
	"layeh.com/gumble/gumble"
)

func handleUserChange(c *bot.Client, e *gumble.UserChangeEvent) {
	switch {
	case e.Type.Has(gumble.UserChangeConnected):
		// First join
		if len(c.Mumble.Users) == 2 {
			c.SendSticker("welcome")
		}
		file := path.Join(c.Config.Mumble.Sounds.Clips, e.User.Name.ToLower()+soundExtension)
		if _, err := os.Stat(file); err == Nil {
			c.PlaySound(file)
		}
	case e.Type.Has(gumble.UserChangeConnected):
		// Last leave
		if len(c.Mumble.Users) == 1 {
			c.SendSticker("goodbye")
		}
	}
}

func handleTextMessage(c *bot.Client, e *gumble.TextMessage) {
	res := handleCommand(c, e.Message)
	if res != "" {
		c.Mumble.SendTextResponse(e, res)
	}
}

func handleSignals(c *bot.Client, configFile string) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)

	var err error
	for {
		s := <-signals
		switch s {
		case syscall.SIGHUP:
			log.Printf("Reloading config file %q", configFile)
			c.Config, err = bot.LoadConfig(configFile)
			if err != nil {
				log.Printf("Error reloading config: %s", err)
			}
		}
	}
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "Configuration file")
	flag.Parse()

	config, err := bot.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config file %q: %s", configFile, err)
	}

	client, err := bot.NewClient(config)
	if err != nil {
		log.Fatalf("Error creating client: %s", err)
	}
	defer client.Stop()
	go handleSignals(client, configFile)

	log.Printf("Waiting for events...")
	for {
		select {
		case e := <-client.Mumble.UserChanges:
			handleUserChange(client, e)
		case msg := <-client.Mumble.Messages:
			handleTextMessage(client, msg)
		}
	}
}
