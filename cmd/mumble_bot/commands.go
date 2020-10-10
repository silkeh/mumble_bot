package main

import (
	"bytes"
	"fmt"
	"html/template"
	"path"
	"strconv"
	"strings"

	"github.com/justinian/dice"
	"github.com/silkeh/mumble_bot/bot"
)

// commandPrefix contains the prefix for all commands.
var commandPrefix = "!"

// soundExtension contains the filename extension for all sound files.
var soundExtensions = []string{".raw", ".opus"}

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
	"!roll":     commandDiceRoll,
}

func handleCommand(c *bot.Client, s string) string {
	cmd, args := parseCommand(s)
	if f, ok := commandHandlers[cmd]; ok {
		return f(c, cmd, args...)
	}

	return commandDefault(c, cmd, args...)
}

var templates *template.Template

type soundUsageParams struct {
	Command string
	Files   []string
}

var soundUsage = `
Usage: {{.Command}} &lt;name&gt;<br/>
Where &lt;name&gt; is one of:
<ul>
{{range .Files}}
<li>{{.}}</li>
{{end}}
</ul>
`

func init() {
	templates = template.Must(template.New("sound").Parse(soundUsage))
}

func commandHold(c *bot.Client, cmd string, args ...string) (resp string) {
	if len(args) < 1 {
		return renderSoundUsage(cmd, c.Config.Mumble.Sounds.Hold)
	}

	file, err := findFile(
		path.Join(c.Config.Mumble.Sounds.Clips, strings.Join(args, " ")),
		soundExtensions...)
	if err != nil {
		return err.Error()
	}

	if err := c.PlayHold(file); err != nil {
		return fmt.Sprintf("Error playing hold music %q: %s", args[0], err)
	}
	return
}

func commandClip(c *bot.Client, cmd string, args ...string) (resp string) {
	if len(args) < 1 {
		return renderSoundUsage(cmd, c.Config.Mumble.Sounds.Clips)
	}

	file, err := findFile(
		path.Join(c.Config.Mumble.Sounds.Clips, strings.Join(args, " ")),
		soundExtensions...)
	if err != nil {
		return err.Error()
	}

	if err := c.PlaySound(file); err != nil {
		return fmt.Sprintf("Error playing music clip %q: %s", args[0], err)
	}
	return
}

func commandSetVolume(c *bot.Client, cmd string, args ...string) (resp string) {
	if len(args) != 1 {
		return fmt.Sprintf("Volume is %v/%v", c.Volume(), bot.MaxVolume)
	}

	v, err := strconv.ParseUint(args[0], 10, 8)
	if err != nil || v > bot.MaxVolume || v < bot.MinVolume {
		return fmt.Sprintf("Usage: %s %v-%v", cmd, bot.MinVolume, bot.MaxVolume)
	}

	c.SetVolume(uint8(v % 256))
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
		return handleCommand(c, commandPrefix+alias)
	}

	return fmt.Sprintf("Unknown command: %s", cmd)
}

func renderSoundUsage(command, path string) string {
	files, err := listFiles(path, soundExtensions...)
	if err != nil {
		return err.Error()
	}
	params := struct {
		Command string
		Files   []string
	}{
		command,
		files,
	}
	usage, err := renderTemplate("sound", params)
	if err != nil {
		return err.Error()
	}
	return usage
}

func renderTemplate(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, name, data)
	return buf.String(), err
}

func commandDiceRoll(c *bot.Client, cmd string, args ...string) (resp string) {
	if len(args) != 1 {
		return fmt.Sprintf("Usage: %s &lt;description&gt;<br/>Example: %s 4d20", cmd, cmd)
	}

	result, _, err := dice.Roll(args[0])
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	switch r := result.(type) {
	case dice.StdResult:
		resp = fmt.Sprintf("%v", r.Total)
		if len(r.Rolls) > 1 {
			resp += fmt.Sprintf(" (%s)", intJoin(r.Rolls, "+"))
		}
		if len(r.Dropped) > 0 {
			resp += fmt.Sprintf(" (dropped %s)", intJoin(r.Dropped, ", "))
		}
	case dice.FudgeResult:
		resp = fmt.Sprintf("%v", r.Total)
		if len(r.Rolls) > 1 {
			resp += fmt.Sprintf(" (%s)", intJoin(r.Rolls, "+"))
		}
	case dice.VsResult:
		resp = fmt.Sprintf("successes: %v", r.Successes)
		if len(r.Rolls) > 1 {
			resp += fmt.Sprintf(" (%s)", intJoin(r.Rolls, ", "))
		}
	default:
		resp = result.String()
	}

	return
}

func intJoin(elems []int, sep string) string {
	strs := make([]string, len(elems))
	for i, v := range elems {
		strs[i] = fmt.Sprintf("%v", v)
	}
	return strings.Join(strs, sep)
}
