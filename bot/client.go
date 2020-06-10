package bot

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
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
	volume   float32
}

// NewClient initializes the client with a given config.
// Either Matrix or Telegram may be configured, not both at the same time.
func NewClient(config *Config) (c *Client, err error) {
	c = &Client{Config: config, volume: 1}

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

// ChangeVolume changes the volume of any Mumble audio played.
func (c *Client) ChangeVolume(f float32) {
	c.Lock()
	defer c.Unlock()
	c.volume *= f
}

// Volume returns the current volume level.
func (c *Client) Volume() float32 {
	c.Lock()
	defer c.Unlock()
	return c.volume
}

// PlayHold plays hold music from a raw 16-bit 48k PCM file in a loop until Mumble.StopAudio() is called.
func (c *Client) PlayHold(path string) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	bytes, err := ioutil.ReadAll(fh)
	if err != nil {
		return err
	}

	// Play the loop in separate threads
	ch := make(chan int16)
	go c.Mumble.StreamAudio(ch)
	go c.loopRaw(ch, bytes)
	return nil
}

// PlaySound plays a sound file containing raw 16-bit 48k PCM file once,
// or until Mumble.StopAudio() is called.
func (c *Client) PlaySound(path string) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	bytes, err := ioutil.ReadAll(fh)
	if err != nil {
		return err
	}

	go c.playRaw(bytes)
	return nil
}

// playRaw plays a byte slice containing 16-bit 48k PCM audio once.
func (c *Client) playRaw(bytes []byte) {
	// Decode and adjust audio
	volume := c.Volume()
	samples := make([]int16, len(bytes)/2)
	for i := range samples {
		samples[i] = int16(float32(int16(binary.LittleEndian.Uint16(bytes[i*2:i*2+2]))) * volume)
	}

	// Play the audio
	c.Mumble.Self.SetSelfMuted(false)
	c.Mumble.SendAudio(samples)
	c.Mumble.Self.SetSelfDeafened(true)
}

// loopRaw loops a byte containing 16-bit 48k PCM audio in a loop,
// with volume adjusted on the fly.
func (c *Client) loopRaw(ch chan<- int16, bytes []byte) {
	c.Mumble.Self.SetSelfMuted(false)
	defer c.Mumble.Self.SetSelfDeafened(true)
	defer close(ch)

	audioFrameSize := c.Mumble.Config.AudioFrameSize()
	volume := c.Volume()
	for i := 0; true; i = (i + 1) % (len(bytes) / 2) {
		ch <- int16(float32(int16(binary.LittleEndian.Uint16(bytes[i*2:i*2+2]))) * volume)

		// Do the slow updates every frame
		if i%audioFrameSize == 0 {
			volume = c.Volume()
			if c.Mumble.AudioStopped() {
				break
			}
		}
	}
}

// Stop stops this client.
func (c *Client) Stop() {
	c.Telegram.Stop()
	c.Mumble.Disconnect()
}
