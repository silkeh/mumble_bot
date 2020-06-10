package mumble

import (
	"strings"
	"sync"
	"time"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"
	_ "layeh.com/gumble/opus" // Ensures Opus compatibility
)

// Preconfigured output events
const (
	Join  = "join"
	Leave = "leave"
)

// Client is a thread-safe mumble client
type Client struct {
	sync.Mutex
	*gumble.Client
	Out       chan string
	audioOut  sync.Mutex
	stopAudio bool
}

// NewClient initialises and returns a Mumble Client
func NewClient(server, user string) (c *Client, err error) {
	c = &Client{
		Out: make(chan string),
	}

	// Client configuration
	config := gumble.NewConfig()
	config.Username = user
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

// changeHandler handles room membership changes
func (c *Client) changeHandler(e *gumble.UserChangeEvent) {
	if e.Type.Has(gumble.UserChangeConnected) {
		c.Out <- Join
	} else if e.Type.Has(gumble.UserChangeDisconnected) {
		c.Out <- Leave
	}
}

// textMessageHandler handler text messages, and passes commands on
func (c *Client) textMessageHandler(e *gumble.TextMessageEvent) {
	if strings.HasPrefix(e.TextMessage.Message, "!") {
		c.Out <- e.TextMessage.Message
	}
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
