package main

import "fmt"

// CommandHandler is the function signature for a command handler.
type CommandHandler func(c *Client, cmd string, args ...string) (resp string)

var commandHandlers = map[string]CommandHandler {
	"!hold": commandHold,
	"!volume--": commandDecreaseVolume,
	"!volume++": commandIncreaseVolume,
	"!stop": commandStopAudio,
	"": commandSendSticker,
}

func commandHold (c *Client, cmd string, args ...string) (resp string) {
	file, ok := c.Config.Mumble.HoldMusic[args[0]]
	if !ok {
		return "Unknown hold music"
	}
	go c.PlayHold(file)
	return
}

func commandDecreaseVolume (c *Client, cmd string, args ...string) (resp string) {
	c.ChangeVolume(0.5)
	return
}

func commandIncreaseVolume (c *Client, cmd string, args ...string) (resp string) {
	c.ChangeVolume(2)
	return
}

func commandStopAudio (c *Client, cmd string, args ...string) (resp string) {
	c.Mumble.StopAudio()
	return
}

func commandSendSticker (c *Client, cmd string, args ...string) (resp string) {
	err := c.SendSticker(cmd[1:])
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return
}
