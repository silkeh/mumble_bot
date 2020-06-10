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
	Messages    chan *gumble.TextMessage
	UserChanges chan *gumble.UserChangeEvent
	Audio       *AudioListener
	audioOut    sync.Mutex
	stopAudio   bool
}

// NewClient initialises and returns a Mumble Client.
func NewClient(server, user string) (c *Client, err error) {
	c = &Client{
		Messages:    make(chan *gumble.TextMessage),
		UserChanges: make(chan *gumble.UserChangeEvent),
		Audio:       new(AudioListener),
	}

	// Client configuration
	config := gumble.NewConfig()
	config.Username = user
	config.AttachAudio(c.Audio)
	config.Attach(gumbleutil.Listener{
		UserChange:  c.changeHandler,
		TextMessage: c.textMessageHandler,
	})

	// Create connection
	c.Client, err = gumble.Dial(server, config)
	if err != nil {
		return
	}

	c.Self.SetSelfDeafened(true)
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
	c.audioOut.Lock()
	defer c.audioOut.Unlock()

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
	c.audioOut.Lock()
	defer c.audioOut.Unlock()

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
