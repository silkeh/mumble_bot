package mumble

import (
	"strings"
	"sync"

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
	Out chan string
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
