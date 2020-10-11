package bot

import (
	"fmt"
	"io"
	"sync"

	"github.com/silkeh/mumble_bot/matrix"
	"github.com/silkeh/mumble_bot/mumble"
	"github.com/silkeh/mumble_bot/telegram"
)

// Client is a thread-safe multi-chat client.
type Client struct {
	sync.Mutex
	Config   *Config
	Mumble   *mumble.Client
	Matrix   *matrix.Client
	Telegram *telegram.Client
	volume   uint8
}

const (
	// MinVolume represents the minimum volume that can be set.
	MinVolume = 0

	// MaxVolume represents the maximum volume that can be set.
	MaxVolume = 16
)

// NewClient initializes the client with a given config.
// Either Matrix or Telegram may be configured, not both at the same time.
func NewClient(config *Config) (c *Client, err error) {
	c = &Client{Config: config, volume: 15}

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
func (c *Client) SetVolume(n uint8) {
	c.Lock()
	defer c.Unlock()
	c.volume = bound8(n, MinVolume, MaxVolume)
}

// ChangeVolume changes the volume of any Mumble audio played.
func (c *Client) ChangeVolume(n int8) {
	c.Lock()
	defer c.Unlock()
	c.volume = bound8(uint8(int8(c.volume)+n), MinVolume, MaxVolume)
}

// Volume returns the current volume level.
func (c *Client) Volume() uint8 {
	c.Lock()
	defer c.Unlock()
	return c.volume
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
		volume := MaxVolume - c.Volume()
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
			ch <- sample >> volume
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

func bound8(v, min, max uint8) uint8 {
	if v >= max {
		return max
	}
	if v <= min {
		return min
	}
	return v
}
