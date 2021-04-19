package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/silkeh/mumble_bot/api"
	"github.com/silkeh/mumble_bot/bot"
)

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

	if config.API != nil {
		log.Printf("API listening on %q", config.API.Address)
		go func() {
			log.Fatal(api.NewAPI(client, nil).ListenAndServe(config.API.Address))
		}()
	}

	log.Printf("Waiting for events...")
	log.Fatal(client.Run())
}
