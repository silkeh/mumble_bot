package main

import (
	"fmt"
	"github.com/silkeh/mumble_bot/bot"
	"os"
	"path"
	"strconv"
	"strings"
)

// commandPrefix contains the prefix for all commands.
var commandPrefix = "!"

// CommandHandler is the function signature for a command handler.
type CommandHandler func(c *bot.Client, cmd string, args ...string) (resp string)

// commandHandlers contains handlers for given commands.
var commandHandlers = map[string]CommandHandler{
	"!hold":     commandHold,
	"!play":     commandClip,
	"!volume":   commandSetVolume,
	"!volume--": commandDecreaseVolume,
	"!volume++": commandIncreaseVolume,
	"!stop":     commandStopAudio,
	"!sticker":  commandSendSticker,
}

func handleCommand(c *bot.Client, s string) string {
	cmd, args := parseCommand(s)
	if f, ok := commandHandlers[cmd]; ok {
		return f(c, cmd, args...)
	}

	return commandDefault(c, cmd, args...)
}

func commandHold(c *bot.Client, cmd string, args ...string) (resp string) {
	if len(args) != 1 {
		return fmt.Sprintf("Usage: %s <name>", cmd)
	}

	file := path.Join(c.Config.Mumble.Sounds.Hold, args[0] + ".raw")
	if _, err := os.Stat(file); os.IsNotExist(err) {
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

	file := path.Join(c.Config.Mumble.Sounds.Clips, args[0] + ".raw")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return "Unknown hold music"
	}

	if err := c.PlaySound(file); err != nil {
		return fmt.Sprintf("Error playing music clip %q: %s", args[0], err)
	}
	return
}

func commandSetVolume(c *bot.Client, cmd string, args ...string) (resp string) {
	usage := fmt.Sprintf("Usage: %s %v-%v", cmd, bot.MinVolume, bot.MaxVolume)
	if len(args) != 1 {
		return usage
	}

	v, err := strconv.ParseUint(args[0], 10, 8)
	if err != nil || v > bot.MaxVolume || v < bot.MinVolume {
		return usage
	}

	c.SetVolume(uint8(v%256))
	return fmt.Sprintf("Volume set to %v", c.Volume())
}

func commandDecreaseVolume(c *bot.Client, cmd string, args ...string) (resp string) {
	c.ChangeVolume(-1)
	return fmt.Sprintf("Volume set to %v", c.Volume())
}

func commandIncreaseVolume(c *bot.Client, cmd string, args ...string) (resp string) {
	c.ChangeVolume(1)
	return fmt.Sprintf("Volume set to %v", c.Volume())
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

func commandDefault(c *bot.Client, cmd string, args ...string) (resp string) {
	// Ignore non-commands
	if !strings.HasPrefix(cmd, commandPrefix) {
		return ""
	}

	// Resolve any configured aliases
	if alias, ok := c.Config.Mumble.Alias[cmd[len(commandPrefix):]]; ok {
		return handleCommand(c, commandPrefix + alias)
	}

	return fmt.Sprintf("Unknown command: %s", cmd)
}
