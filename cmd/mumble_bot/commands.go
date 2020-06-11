package main

import (
	"fmt"
	"github.com/silkeh/mumble_bot/bot"
)

// CommandHandler is the function signature for a command handler.
type CommandHandler func(c *bot.Client, cmd string, args ...string) (resp string)

// commandHandlers contains handlers for given commands.
var commandHandlers = map[string]CommandHandler{
	"!hold":     commandHold,
	"!play":     commandClip,
	"!volume--": commandDecreaseVolume,
	"!volume++": commandIncreaseVolume,
	"!stop":     commandStopAudio,
	"":          commandSendSticker,
}

func commandHold(c *bot.Client, cmd string, args ...string) (resp string) {
	file, ok := c.Config.Mumble.Music.Hold[args[0]]
	if !ok {
		return "Unknown hold music"
	}
	if err := c.PlayHold(file); err != nil {
		return fmt.Sprintf("Error playing hold music %q: %s", args[0], err)
	}
	return
}

func commandClip(c *bot.Client, cmd string, args ...string) (resp string) {
	if len(args) != 1 {
		return fmt.Sprintf("Usage: %s <clip>", cmd)
	}

	file, ok := c.Config.Mumble.Music.Clips[args[0]]
	if !ok {
		return fmt.Sprintf("Unknown music clip: %q", args[0])
	}

	if err := c.PlaySound(file); err != nil {
		return fmt.Sprintf("Error playing music clip %q: %s", args[0], err)
	}
	return
}

func commandDecreaseVolume(c *bot.Client, cmd string, args ...string) (resp string) {
	c.ChangeVolume(0.5)
	return fmt.Sprintf("Volume set to %.1f", c.Volume())
}

func commandIncreaseVolume(c *bot.Client, cmd string, args ...string) (resp string) {
	c.ChangeVolume(2)
	return fmt.Sprintf("Volume set to %.1f", c.Volume())
}

func commandStopAudio(c *bot.Client, cmd string, args ...string) (resp string) {
	c.Mumble.StopAudio()
	return
}

func commandSendSticker(c *bot.Client, cmd string, args ...string) (resp string) {
	if len(args) != 1 {
		return fmt.Sprintf("Usage: %s <sticker>", cmd)
	}

	err := c.SendSticker(args[0])
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return
}
