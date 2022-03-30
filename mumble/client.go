package mumble

import (
	"strings"
	"sync"
	"time"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"
	_ "layeh.com/gumble/opus" // Ensures Opus compatibility
)

// Client is a thread-safe mumble client.
type Client struct {
	sync.Mutex
	*gumble.Client
	Messages      chan *gumble.TextMessage
	UserChanges   chan *gumble.UserChangeEvent
	audioOut      sync.Mutex
	audioMuted    bool
	audioDeafened bool
	stopAudio     bool
	selfMuted     bool
	selfDeafened  bool
}

// NewClient initialises and returns a Mumble Client.
func NewClient(server, user string, l gumble.AudioListener) (c *Client, err error) {
	c = &Client{
		Messages:    make(chan *gumble.TextMessage),
		UserChanges: make(chan *gumble.UserChangeEvent),
	}

	// Client configuration
	config := gumble.NewConfig()
	config.Username = user
	config.Attach(gumbleutil.Listener{
		UserChange:  c.changeHandler,
		TextMessage: c.textMessageHandler,
	})

	if l != nil {
		config.AttachAudio(l)
	}

	// Create connection
	c.Client, err = gumble.Dial(server, config)
	if err != nil {
		return
	}

	c.SetSelfDeafened(true)
	return
}

// changeHandler handles room membership changes.
func (c *Client) changeHandler(e *gumble.UserChangeEvent) {
	if e.Type.Has(gumble.UserChangeConnected) || e.Type.Has(gumble.UserChangeDisconnected) {
		c.UserChanges <- e
	}
}

// textMessageHandler handles text messages, and passes commands on.
func (c *Client) textMessageHandler(e *gumble.TextMessageEvent) {
	if strings.HasPrefix(e.TextMessage.Message, "!") {
		c.Messages <- &e.TextMessage
	}
}

// SendTextResponse sends a simple text response to the given message.
func (c *Client) SendTextResponse(e *gumble.TextMessage, msg string) {
	c.Send(&gumble.TextMessage{
		Sender:   c.Self,
		Channels: e.Channels,
		Message:  msg,
	})
}

// SendAudio sends the given 48 kHz 16-bit PCM audio to the main audio channel.
// This function waits for any earlier SendAudio() or StreamAudio() calls to finish.
func (c *Client) SendAudio(samples []int16) {
	c.LockAudio()
	defer c.UnlockAudio()

	out := c.AudioOutgoing()
	defer close(out)

	frameSize := c.Config.AudioFrameSize()
	ticker := time.NewTicker(c.Config.AudioInterval)
	for i := 0; i < len(samples)/frameSize; i++ {
		out <- samples[i*frameSize : i*frameSize+frameSize]

		if c.AudioStopped() {
			break
		}
		<-ticker.C
	}

	c.stopAudio = false
}

// StreamAudio sends the given 48 kHz 16-bit PCM audio to the main audio channel.
// This function waits for any earlier SendAudio() or StreamAudio() calls to finish.
func (c *Client) StreamAudio(ch <-chan int16) {
	c.LockAudio()
	defer c.UnlockAudio()

	out := c.AudioOutgoing()
	defer close(out)

	ticker := time.NewTicker(c.Config.AudioInterval)
	buf := make([]int16, 0, c.Config.AudioFrameSize())
	for s := range ch {
		buf = append(buf, s)
		if len(buf) == cap(buf) {
			out <- buf
			buf = make([]int16, 0, c.Config.AudioFrameSize())
			<-ticker.C
		}
	}

	if len(buf) > 0 {
		out <- buf
	}

	c.stopAudio = false
}

// AudioStopped returns true if audio should be stopped.
func (c *Client) AudioStopped() bool {
	c.Lock()
	defer c.Unlock()
	return c.stopAudio
}

// StopAudio requests all currently playing audio to be stopped.
func (c *Client) StopAudio() {
	c.Lock()
	defer c.Unlock()
	c.stopAudio = true
}

// LockAudio requests an audio play lock.
// This unmutes and undeafens, which is restored in UnlockAudio().
func (c *Client) LockAudio() {
	c.audioOut.Lock()
	c.audioMuted = c.SelfMuted()
	c.audioDeafened = c.SelfDeafened()
	c.SetSelfMuted(false)
}

// UnlockAudio releases the audio play lock.
// This restores the muted and deafened state when the lock was requested.
func (c *Client) UnlockAudio() {
	if c.audioDeafened {
		c.SetSelfDeafened(c.audioDeafened)
	} else if c.audioMuted {
		c.SetSelfMuted(c.audioMuted)
	}
	c.audioOut.Unlock()
}

// SelfMuted shows whether the client can transmit audio or not.
func (c *Client) SelfMuted() bool {
	c.Lock()
	defer c.Unlock()

	return c.selfMuted
}

// SetSelfMuted sets whether the client can transmit audio or not.
// Setting this to false also sets deafened to false.
// Note that this is bypassed by calling Client.Self.SetSelfMuted(), so don't.
func (c *Client) SetSelfMuted(muted bool) {
	c.Lock()
	defer c.Unlock()

	c.selfMuted = muted
	c.Self.SetSelfMuted(muted)
}

// SelfDeafened shows whether the client can receive audio or not.
func (c *Client) SelfDeafened() bool {
	c.Lock()
	defer c.Unlock()

	return c.selfDeafened
}

// SetSelfDeafened sets whether the client can receive audio or not.
// Setting this to true also sets muted to true.
// Note that this is bypassed by calling Client.Self.SetSelfDeafened(), so don't.
func (c *Client) SetSelfDeafened(deafened bool) {
	c.Lock()
	defer c.Unlock()

	c.selfDeafened = deafened
	c.Self.SetSelfDeafened(deafened)
}
