package bot

import (
	"bytes"
	"fmt"
	"html/template"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/justinian/dice"
)

// CommandHandler is the function signature for a command handler.
type CommandHandler func(c *Client, cmd string, args ...string) (resp string)

// SoundExtension contains the filename extension for all sound files.
const SoundExtension = ".opus"

// commandHandlers contains handlers for given commands.
var defaultCommands = map[string]CommandHandler{
	"hold":     CommandHold,
	"play":     CommandClip,
	"volume":   CommandSetVolume,
	"volume--": CommandDecreaseVolume,
	"volume++": CommandIncreaseVolume,
	"stop":     CommandStopAudio,
	"sticker":  CommandSendSticker,
	"roll":     CommandDiceRoll,
	"shell":    CommandShell,
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

// CommandHold plays a given sound file in a loop (like hold music).
func CommandHold(c *Client, cmd string, args ...string) (resp string) {
	if len(args) < 1 {
		return renderSoundUsage(cmd, c.Config.Mumble.Sounds.Hold)
	}

	name := strings.Join(args, " ")
	file := path.Join(c.Config.Mumble.Sounds.Clips, name)
	if err := c.PlayHold(file + SoundExtension); err != nil {
		return fmt.Sprintf("Error playing hold music %q: %s", args[0], err)
	}
	return fmt.Sprintf("Please hold. Now playing %q...", name)
}

// CommandClip plays a sound file once.
func CommandClip(c *Client, cmd string, args ...string) (resp string) {
	if len(args) < 1 {
		return renderSoundUsage(cmd, c.Config.Mumble.Sounds.Clips)
	}

	name := strings.Join(args, " ")
	file := path.Join(c.Config.Mumble.Sounds.Clips, name)
	if err := c.PlaySound(file + SoundExtension); err != nil {
		return fmt.Sprintf("Error playing music clip %q: %s", name, err)
	}
	return fmt.Sprintf("Now playing %q...", name)
}

// CommandSetVolume sets the volume of the bot to a given value.
func CommandSetVolume(c *Client, cmd string, args ...string) (resp string) {
	if len(args) != 1 {
		return fmt.Sprintf("Volume is %+v dB (max %+v dB)", c.Volume(), MaxVolume)
	}

	v, err := strconv.ParseInt(args[0], 10, 8)
	if err != nil || v > MaxVolume || v < MinVolume {
		return fmt.Sprintf("Usage: %s %v-%v (in dB)", cmd, MinVolume, MaxVolume)
	}

	c.SetVolume(int8(v % 256))
	return fmt.Sprintf("Volume set to %+v dB", c.Volume())
}

// CommandDecreaseVolume decreases the volume by one step.
func CommandDecreaseVolume(c *Client, cmd string, args ...string) (resp string) {
	c.ChangeVolume(-3)
	return fmt.Sprintf("Volume set to %+v dB", c.Volume())
}

// CommandIncreaseVolume increases the volume by one step.
func CommandIncreaseVolume(c *Client, cmd string, args ...string) (resp string) {
	c.ChangeVolume(3)
	return fmt.Sprintf("Volume set to %+v dB", c.Volume())
}

// CommandStopAudio stops any playing audio.
func CommandStopAudio(c *Client, cmd string, args ...string) (resp string) {
	c.Mumble.StopAudio()
	return
}

// CommandSendSticker sends a sticker to a linked chat platform.
func CommandSendSticker(c *Client, cmd string, args ...string) (resp string) {
	if len(args) != 1 {
		return fmt.Sprintf("Usage: %s <sticker>", cmd)
	}

	err := c.SendSticker(args[0])
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return
}

func commandDefault(c *Client, cmd string, args ...string) (resp string) {
	// Resolve any configured aliases
	if alias, ok := c.Config.Mumble.Alias[cmd]; ok {
		if len(args) > 0 {
			alias += " " + strings.Join(args, " ")
		}
		return c.HandleCommand(alias)
	}

	return fmt.Sprintf("Unknown command: %s", cmd)
}

func renderSoundUsage(command, path string) string {
	files, err := listFiles(path, SoundExtension)
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

// CommandDiceRoll rolls a (set of) dice and prints the result.
// See https://github.com/justinian/dice for features and syntax.
func CommandDiceRoll(c *Client, cmd string, args ...string) (resp string) {
	if len(args) != 1 {
		return fmt.Sprintf("Usage: %s &lt;description&gt;<br/>Example: %s 4d20", cmd, cmd)
	}

	result, _, err := dice.Roll(args[0])
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	resp = fmt.Sprintf("Rolled %s: ", args[0])
	switch r := result.(type) {
	case dice.StdResult:
		resp += fmt.Sprintf("%v", r.Total)
		if len(r.Rolls) > 1 {
			resp += fmt.Sprintf(" (%s)", intJoin(r.Rolls, "+"))
		}
		if len(r.Dropped) > 0 {
			resp += fmt.Sprintf(" (dropped %s)", intJoin(r.Dropped, ", "))
		}
	case dice.FudgeResult:
		resp += fmt.Sprintf("%v", r.Total)
		if len(r.Rolls) > 1 {
			resp += fmt.Sprintf(" (%s)", intJoin(r.Rolls, "+"))
		}
	case dice.VsResult:
		resp += fmt.Sprintf("successes: %v", r.Successes)
		if len(r.Rolls) > 1 {
			resp += fmt.Sprintf(" (%s)", intJoin(r.Rolls, ", "))
		}
	default:
		resp += result.String()
	}

	return
}

// CommandShell executes a shell script in the configured script directory
func CommandShell(c *Client, cmd string, args ...string) (resp string) {
	if c.Config.Mumble.Script.Directory == "" {
		return fmt.Sprintf("%q command is not enabled", cmd)
	}

	if len(args) == 0 {
		return fmt.Sprintf("Usage: %s &lt;scripts&gt; [arguments...]", cmd)
	}

	cmd = args[0]
	args = args[1:]
	if strings.Contains(cmd, `/`) || strings.Contains(cmd, `\`) {
		return fmt.Sprintf("Invalid command: %q", cmd)
	}

	cmd = path.Join(c.Config.Mumble.Script.Directory, cmd)
	out := new(bytes.Buffer)
	exe := exec.Command(cmd, args...)
	exe.Stdout = out
	exe.Stderr = out

	if err := exe.Run(); err != nil {
		return fmt.Sprintf("Error: %s<br/><pre>%s</pre>", err, out.String())
	}

	return out.String()
}
