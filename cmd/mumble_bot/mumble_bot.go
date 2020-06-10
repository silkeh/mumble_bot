package main

import (
	"flag"
	"layeh.com/gumble/gumble"
	"log"
	"strings"
)

func handleUserChange(c *Client, e *gumble.UserChangeEvent) {
	switch {
	case e.Type.Has(gumble.UserChangeConnected):
		// First join
		if len(c.Mumble.Users) == 2 {
			c.SendSticker("welcome")
		}
	case e.Type.Has(gumble.UserChangeConnected):
		// Last leave
		if len(c.Mumble.Users) == 1 {
			c.SendSticker("goodbye")
		}
	}
}

func handleTextMessage(c *Client, e *gumble.TextMessage) {
	res := ""
	cmd := strings.Split(e.Message, " ")
	if f, ok := commandHandlers[cmd[0]]; ok {
		res = f(c, cmd[0], cmd[1:]...)
	} else {
		res = commandHandlers[""](c, cmd[0], cmd[1:]...)
	}
	if res != "" {
		c.Mumble.SendTextResponse(e, res)
	}
}

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
	for {
		select {
		case e := <-client.Mumble.UserChanges:
			handleUserChange(client, e)
		case msg := <-client.Mumble.Messages:
			handleTextMessage(client, msg)
		}
	}
}
