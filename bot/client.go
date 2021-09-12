package bot

import (
	"fmt"
	"io"
	"math"
	"strings"
	"sync"

	"github.com/silkeh/mumble_bot/matrix"
	"github.com/silkeh/mumble_bot/mumble"
	"github.com/silkeh/mumble_bot/telegram"
	"layeh.com/gumble/gumble"
)

// Client is a thread-safe multi-chat client.
type Client struct {
	sync.Mutex
	Config   *Config
	Mumble   *mumble.Client
	Matrix   *matrix.Client
	Telegram *telegram.Client
	commands map[string]CommandHandler
	volume   int8
}

const (
	// MinVolume represents the minimum gain in dB that can be set.
	MinVolume = -90

	// MaxVolume represents the maximum gain in dB that can be set.
	MaxVolume = 30

	// DefaultVolume represents the default volume.
	DefaultVolume = -18
)

const (
	joinHook       = "join"
	leaveHook      = "leave"
	firstJoinHook  = "first_join"
	lastLeaveHook  = "last_leave"
	defaultSubject = "default"
)

// NewClient initializes the client with a given config.
// Either Matrix or Telegram may be configured, not both at the same time.
func NewClient(config *Config) (c *Client, err error) {
	c = &Client{Config: config, volume: DefaultVolume, commands: defaultCommands}

	// Check if Matrix and Telegram aren't enabled at the same time.
	if config.Telegram != nil && config.Matrix != nil {
		return nil, fmt.Errorf("both Telegram and Matrix may not be configured at the same time")
	}

	// Telegram
	if config.Telegram != nil {
		c.Telegram, err = telegram.NewClient(config.Telegram.Token, config.Telegram.Target)
		if err != nil {
			return nil, fmt.Errorf("connecting to Telegram: %w", err)
		}
		go c.Telegram.Start()
	}

	// Matrix
	if config.Matrix != nil {
		c.Matrix, err = matrix.NewClient(config.Matrix.Server, config.Matrix.User, config.Matrix.Token, config.Matrix.Room)
		if err != nil {
			return nil, fmt.Errorf("connecting to Matrix: %w", err)
		}
	}

	// Mumble
	c.Mumble, err = mumble.NewClient(config.Mumble.Server, config.Mumble.User)
	if err != nil {
		return nil, fmt.Errorf("connecting to Mumble: %w", err)
	}

	return
}

// Run the client.
func (c *Client) Run() error {
	for {
		select {
		case e := <-c.Mumble.UserChanges:
			c.handleUserChange(e)
		case msg := <-c.Mumble.Messages:
			c.handleTextMessage(msg)
		}
	}
}

func (c *Client) handleUserChange(e *gumble.UserChangeEvent) {
	switch {
	case e.Type.Has(gumble.UserChangeConnected):
		if len(c.Mumble.Users) == 2 {
			c.ExecuteHook(firstJoinHook, e.User.Name)
		}
		c.ExecuteHook(joinHook, e.User.Name)
	case e.Type.Has(gumble.UserChangeDisconnected):
		if len(c.Mumble.Users) == 1 {
			c.ExecuteHook(lastLeaveHook, e.User.Name)
		}
		c.ExecuteHook(leaveHook, e.User.Name)
	}
}

func (c *Client) handleTextMessage(e *gumble.TextMessage) {
	if !strings.HasPrefix(e.Message, c.Config.Mumble.CommandPrefix) {
		return
	}

	res := c.HandleCommand(strings.TrimPrefix(e.Message, c.Config.Mumble.CommandPrefix))
	if res != "" {
		c.Mumble.SendTextResponse(e, res)
	}
}

// HandleCommand handles a bot command.
func (c *Client) HandleCommand(s string) string {
	cmd, args := parseCommand(s)
	if f, ok := c.commands[cmd]; ok {
		return f(c, cmd, args...)
	}

	return commandDefault(c, cmd, args...)
}

// ExecuteHook executes a configured hook.
func (c *Client) ExecuteHook(name, subject string) string {
	actions, ok := c.Config.Mumble.Hooks[name]
	if !ok {
		return ""
	}

	command, ok := actions[subject]
	if !ok {
		command, ok = actions[defaultSubject]
		if !ok {
			return ""
		}
	}

	return c.HandleCommand(command)
}

// SendSticker sends a sticker to a either Matrix or Telegram.
func (c *Client) SendSticker(name string) error {
	if c.Telegram != nil {
		s, ok := c.Config.Telegram.Stickers[name]
		if !ok {
			return fmt.Errorf("unknown sticker: %q", name)
		}
		_, err := c.Telegram.SendSticker(s)
		if err != nil {
			return err
		}
	}

	if c.Matrix != nil {
		s, ok := c.Config.Matrix.Stickers[name]
		if !ok {
			return fmt.Errorf("unknown sticker: %q", name)
		}
		_, err := c.Matrix.SendSticker(s)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetVolume sets the volume of any Mumble audio played.
func (c *Client) SetVolume(n int8) {
	c.Lock()
	defer c.Unlock()
	c.volume = bound8(n, MinVolume, MaxVolume)
}

// ChangeVolume changes the volume of any Mumble audio played.
func (c *Client) ChangeVolume(n int8) {
	c.Lock()
	defer c.Unlock()
	c.volume = bound8(c.volume+n, MinVolume, MaxVolume)
}

// Volume returns the current volume gain in dB.
func (c *Client) Volume() int8 {
	c.Lock()
	defer c.Unlock()
	return c.volume
}

// Gain returns the current volume amplitude ratio.
func (c *Client) gain() float64 {
	c.Lock()
	defer c.Unlock()
	return math.Pow(10, float64(c.volume)/20)
}

// PlayHold plays hold music from a raw 16-bit 48k PCM file in a loop until Mumble.StopAudio() is called.
func (c *Client) PlayHold(path string) error {
	return c.playFile(path, -1)
}

// PlaySound plays a sound file containing raw 16-bit 48k PCM file once,
// or until Mumble.StopAudio() is called.
func (c *Client) PlaySound(path string) error {
	return c.playFile(path, 1)
}

// playFile plays an arbitrary audio file.
func (c *Client) playFile(path string, count int) error {
	fh, err := OpenSoundFile(path)
	if err != nil {
		return err
	}

	// Play the loop in separate threads
	ch := make(chan int16)
	go c.Mumble.StreamAudio(ch)
	go c.playRaw(ch, NewAudioLoop(fh, count))
	return nil
}

// playRaw loops a byte containing 16-bit 48k PCM audio a number of times,
// with volume adjusted on the fly.
func (c *Client) playRaw(ch chan<- int16, stream AudioStream) {
	defer stream.Close()
	defer close(ch)

	buf := make([]int16, c.Mumble.Config.AudioFrameSize())
	for i := 0; true; i++ {
		// Do the slow updates every stream
		volume := c.gain()
		if c.Mumble.AudioStopped() {
			break
		}

		// Read the audio from the stream
		n, err := stream.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}

		// Stream the audio
		for _, sample := range buf[:n] {
			ch <- int16(float64(sample) * volume)
		}

		// Stop if the file/stream has ended
		if err == io.EOF {
			break
		}
	}
}

// Stop stops this client.
func (c *Client) Stop() {
	c.Telegram.Stop()
	c.Mumble.Disconnect()
}

func bound8(v, min, max int8) int8 {
	if v >= max {
		return max
	}
	if v <= min {
		return min
	}
	return v
}
